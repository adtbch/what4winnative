package eventbus

import (
	"sync"
)

type Bus struct {
	mu   sync.RWMutex
	subs map[string][]chan any
}

var Default = &Bus{
	subs: make(map[string][]chan any),
}

func (b *Bus) Subscribe(name string) <-chan any {
	b.mu.Lock()
	defer b.mu.Unlock()
	ch := make(chan any, 64)
	b.subs[name] = append(b.subs[name], ch)
	return ch
}

func (b *Bus) Publish(name string, data any) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	for _, ch := range b.subs[name] {
		select {
		case ch <- data:
		default:
		}
	}
}
