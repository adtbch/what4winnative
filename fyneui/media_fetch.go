package fyneui

import (
	"log"
	"sync"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
)

// mediaFetchJob is a bounded, deduped media load request. It avoids spawning
// one goroutine per bubble (RAM/CPU spike on long lists) while preserving the
// existing fetch order: thumbnail -> cache -> download (images/stickers only).
type mediaFetchJob struct {
	cv     *ChatView
	chatID string
	msgID  string
	kind   int
	img    *canvas.Image
	box    fyne.CanvasObject
	gen    uint64 // stale-guard: must match cv generation when applying
}

const mediaFetchWorkers = 3

var (
	mediaFetchOnce   sync.Once
	mediaFetchQueue  chan mediaFetchJob
	mediaFetchInflight sync.Map // key msgID -> struct{}
)

func startMediaFetchers() {
	mediaFetchOnce.Do(func() {
		mediaFetchQueue = make(chan mediaFetchJob, 256)
		for i := 0; i < mediaFetchWorkers; i++ {
			go mediaFetchWorker()
		}
	})
}

func enqueueMediaFetch(cv *ChatView, chatID, msgID string, kind int, img *canvas.Image, box fyne.CanvasObject) {
	if cv == nil || cv.fa == nil || cv.fa.api == nil || msgID == "" || img == nil {
		return
	}
	startMediaFetchers()
	if _, loaded := mediaFetchInflight.LoadOrStore(msgID, struct{}{}); loaded {
		return
	}
	job := mediaFetchJob{cv: cv, chatID: chatID, msgID: msgID, kind: kind, img: img, box: box, gen: cv.mediaGen()}
	select {
	case mediaFetchQueue <- job:
	default:
		// queue full: drop rather than block UI; item can be fetched on re-render
		mediaFetchInflight.Delete(msgID)
	}
}

func mediaFetchWorker() {
	for job := range mediaFetchQueue {
		func() {
			defer mediaFetchInflight.Delete(job.msgID)
			if !job.active() {
				return
			}
			var dataURL string
			api := job.cv.fa.api
			if job.kind == mediaKindVideo {
				dataURL = api.GetVideoThumbnail(job.msgID)
			} else {
				dataURL = api.GetImageThumbnail(job.msgID)
				// images/stickers: thumbnail -> cache -> download (same as before)
				if dataURL == "" {
					if u, err := api.GetCachedImage(job.msgID); err == nil {
						dataURL = u
					}
				}
				if dataURL == "" && job.active() {
					if u, err := api.DownloadMedia(job.chatID, job.msgID); err == nil {
						dataURL = u
					} else {
						log.Println("DownloadMedia:", job.msgID, err)
					}
				}
			}
			if dataURL == "" || !job.active() {
				return
			}
			res := dataURLResource(job.msgID, dataURL)
			if res == nil {
				return
			}
			fyne.Do(func() {
				if !job.active() {
					return
				}
				job.img.Resource = res
				job.img.Refresh()
				if job.box != nil {
					job.box.Refresh()
				}
			})
		}()
	}
}

func (j mediaFetchJob) active() bool {
	return j.cv != nil && j.cv.isCurrentMediaGen(j.gen) && j.cv.jid == j.chatID
}

func (cv *ChatView) mediaGen() uint64 {
	cv.mu.Lock()
	defer cv.mu.Unlock()
	return cv.mediaGeneration
}

func (cv *ChatView) bumpMediaGen() {
	cv.mu.Lock()
	cv.mediaGeneration++
	cv.mu.Unlock()
}

func (cv *ChatView) isCurrentMediaGen(gen uint64) bool {
	cv.mu.Lock()
	defer cv.mu.Unlock()
	return cv.mediaGeneration == gen
}
