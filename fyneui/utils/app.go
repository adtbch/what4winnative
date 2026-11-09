package utils

import (
	"sync"

	"github.com/adtbch/whats4linux/api"
)

// AppState holds the session state shared between the shell and its UI
// packages. The API client is injected by fyneui.New; UI navigation drives
// the rest. Exposed via FyneApp.State() so sub-packages never import the
// shell package.
type AppState struct {
	api            *api.Api
	mu             sync.Mutex
	Screen         string
	SelectedChatID string
}

func NewAppState(a *api.Api) *AppState {
	return &AppState{api: a, Screen: "login"}
}

func (s *AppState) SetScreen(screen string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Screen = screen
}

func (s *AppState) SetSelectedChatID(jid string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.SelectedChatID = jid
}

func (s *AppState) GetSelectedChatID() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.SelectedChatID
}

// RLock/RUnlock expose the state mutex so the shell can coordinate with
// sub-package UI without reaching into unexported fields.
func (a *AppState) Lock()   { a.mu.Lock() }
func (a *AppState) Unlock() { a.mu.Unlock() }
