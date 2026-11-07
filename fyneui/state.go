package fyneui

import (
	"sync"

	"github.com/lugvitc/whats4linux/api"
)

type AppState struct {
	mu             sync.RWMutex
	api            *api.Api
	Screen         string // login|chats|settings
	SelectedChatID string
}

func (s *AppState) SetScreen(screen string) {
	s.mu.Lock()
	s.Screen = screen
	s.mu.Unlock()
}

func (s *AppState) SetSelectedChatID(jid string) {
	s.mu.Lock()
	s.SelectedChatID = jid
	s.mu.Unlock()
}

func (s *AppState) GetSelectedChatID() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.SelectedChatID
}
