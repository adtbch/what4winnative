package utils

import "strings"

func IsGroupJID(jid string) bool {
	return strings.HasSuffix(jid, "@s.whatsapp.net") || strings.HasSuffix(jid, "@g.us")
}

func ShortSender(jid string) string {
	if parts := strings.Split(jid, "@"); len(parts) == 2 {
		return parts[0]
	}
	return jid
}

func TruncateRunes(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
