package utils

import "github.com/adtbch/whats4linux/internal/eventbus"

// SubscribeEvent delivers fan-out bus messages published by the WhatsApp API
// layer (QR codes, status changes). Kept here so UI packages never import
// the shell or the API package directly.
func SubscribeEvent(topic string) <-chan any {
	return eventbus.Default.Subscribe(topic)
}
