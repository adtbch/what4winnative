package voip

import (
	"context"
	"errors"
	"log"

	meowcaller "github.com/purpshell/meowcaller"
)

// SetOnVideoFrame registers a callback for inbound H.264 access units (Annex-B,
// one per frame). The callback runs on the media goroutine, so it must not block.
func (m *Manager) SetOnVideoFrame(fn func([]byte)) {
	if m == nil {
		return
	}
	m.mu.Lock()
	m.onVideoFrame = fn
	m.mu.Unlock()
}

// SetOnVideoKeyframeRequest registers a callback fired when the peer asks for an
// IDR frame (WhatsApp PLI/FIR). The frontend encoder must force a keyframe.
func (m *Manager) SetOnVideoKeyframeRequest(fn func()) {
	if m == nil {
		return
	}
	m.mu.Lock()
	m.onVideoKeyframe = fn
	m.mu.Unlock()
}

// attachVideo wires the inbound video sink and state callbacks for a call.
func (m *Manager) attachVideo(call *meowcaller.Call) {
	call.ReceiveVideo(meowcaller.VideoSinkFunc(func(accessUnit []byte) {
		if len(accessUnit) == 0 {
			return
		}
		m.mu.Lock()
		fn := m.onVideoFrame
		m.mu.Unlock()
		if fn != nil {
			fn(accessUnit)
		}
	}))
	call.OnVideoState(func(state meowcaller.VideoState) {
		log.Printf("voip: video state active=%v upgrade=%v orientation=%d raw=%d",
			state.Active, state.Upgrade, state.Orientation, state.Raw)
		m.notify()
	})
	call.OnVideoKeyframeRequest(func() {
		m.mu.Lock()
		fn := m.onVideoKeyframe
		m.mu.Unlock()
		if fn != nil {
			fn()
		}
	})
}

// StartVideo requests an audio-to-video upgrade. Outbound frames stay gated until
// the peer acknowledges the transition.
func (m *Manager) StartVideo() error {
	call, err := m.current()
	if err != nil {
		return err
	}
	if err := call.StartVideo(); err != nil {
		return err
	}
	m.notify()
	return nil
}

// AcceptVideo accepts a pending peer video upgrade.
func (m *Manager) AcceptVideo() error {
	call, err := m.current()
	if err != nil {
		return err
	}
	if err := call.AcceptVideo(); err != nil {
		return err
	}
	m.notify()
	return nil
}

// StopVideo stops outbound video while keeping audio and peer video.
func (m *Manager) StopVideo() error {
	call, err := m.current()
	if err != nil {
		return err
	}
	if err := call.StopVideo(); err != nil {
		return err
	}
	m.notify()
	return nil
}

// PlaceVideo places an outbound call that advertises WhatsApp video from the start.
func (m *Manager) PlaceVideo(ctx context.Context, target string) error {
	if m == nil || m.client == nil {
		return errors.New("voip is not ready")
	}
	if err := m.ensureIdle(); err != nil {
		return err
	}
	call, err := m.client.CallWithOptions(ctx, target, meowcaller.CallOptions{Video: true})
	if err != nil {
		return err
	}
	if err := m.adopt(call, "outgoing", target); err != nil {
		_ = call.Hangup()
		return err
	}
	m.startAudio(call, true)
	log.Printf("voip: placed video call %s to %s", call.ID(), call.Peer())
	return nil
}

// SetVideoEnabled mutes or unmutes the outbound camera flow.
func (m *Manager) SetVideoEnabled(enabled bool) error {
	call, err := m.current()
	if err != nil {
		return err
	}
	if err := call.SetVideoEnabled(enabled); err != nil {
		return err
	}
	m.mu.Lock()
	m.videoMuted = !enabled
	m.mu.Unlock()
	m.notify()
	return nil
}

// SendVideoFrame pushes one already-encoded Annex-B H.264 access unit.
func (m *Manager) SendVideoFrame(accessUnit []byte) error {
	if len(accessUnit) == 0 {
		return nil
	}
	call, err := m.current()
	if err != nil {
		return err
	}
	if !call.IsSendingVideo() {
		return errors.New("outbound video is not active yet")
	}
	return call.SendVideo(accessUnit)
}

// SetVideoOrientation announces the local camera rotation in quarter turns.
func (m *Manager) SetVideoOrientation(orientation int) error {
	call, err := m.current()
	if err != nil {
		return err
	}
	return call.SetVideoOrientation(orientation)
}
