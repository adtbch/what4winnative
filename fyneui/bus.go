package fyneui

import "github.com/lugvitc/whats4linux/internal/eventbus"

func SubscribeEvent(name string) <-chan any {
	return eventbus.Default.Subscribe(name)
}
