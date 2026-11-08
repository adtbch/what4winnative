//go:build !windows

package fyneui

func readClipboardImagePNG() ([]byte, error) { return nil, nil }
