//go:build !windows

package utils

func ReadClipboardImagePNG() ([]byte, error) { return nil, nil }
