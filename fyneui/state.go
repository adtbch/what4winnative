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
