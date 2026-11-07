package fyneui

import (
	"testing"

	"fyne.io/fyne/v2"
)

func TestChatInfoPanelGolden(t *testing.T) {
	p := newChatInfoPanel(func() {})
	p.Set(chatInfoData{Name: "Alice", Subtitle: "+62 812-3456-7890", About: "Hey there! I am using WhatsApp."})
	img := renderPNG(t, "chat_info", p, fyne.NewSize(chatInfoWidth, 520))
	_, _, _, _, n := dominantCardBBox(img, hexNRGBA("#0c0c0c"), 20)
	if n < 20000 {
		t.Fatalf("chat info panel bg weak n=%d", n)
	}
	assertNotBlank(t, img, 0.003)
}

func TestChatInfoPanelGroupTitle(t *testing.T) {
	p := newChatInfoPanel(func() {})
	p.Set(chatInfoData{Name: "Team", Subtitle: "Group · 3 participants", IsGroup: true})
	if p.title.Text != "Group Info" {
		t.Fatalf("group title=%q", p.title.Text)
	}
	p.Set(chatInfoData{Name: "Bob", IsGroup: false})
	if p.title.Text != "Contact Info" {
		t.Fatalf("contact title=%q", p.title.Text)
	}
}

func TestChatInfoPanelClose(t *testing.T) {
	hit := false
	p := newChatInfoPanel(func() { hit = true })
	if p.onClose == nil { t.Fatal("nil close") }
	p.onClose()
	if !hit { t.Fatal("close not called") }
}
