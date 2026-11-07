package fyneui

import (
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"

	"github.com/lugvitc/whats4linux/api"
)

// reactFixture mirrors FIXTURE in frontend/src/harness.tsx exactly, so the Fyne
// PNG and the React PNG render the same three rows and can be diffed.
func reactFixture() ([]api.ChatElement, map[string]int) {
	chats := []api.ChatElement{
		{
			Contact:       api.Contact{JID: "6282134544728@s.whatsapp.net", FullName: "cintakuuu"},
			LatestMessage: "iyaa aku juga sayang kamu",
			LatestTS:      1762500000,
		},
		{
			Contact:       api.Contact{JID: "120363999999999999@g.us", FullName: "Vapers Semarangan 18+"},
			LatestMessage: "besok kumpul jam 8",
			LatestTS:      1762490000,
			Sender:        "Budi",
			Pinned:        true,
		},
		{
			Contact:       api.Contact{JID: "628999888777@s.whatsapp.net", FullName: "+62 899-9888-777"},
			LatestMessage: "📷 Photo",
			LatestTS:      1762400000,
		},
	}
	return chats, map[string]int{
		"6282134544728@s.whatsapp.net": 2,
		"120363999999999999@g.us":      128,
	}
}

// TestChatRowFixtureGolden renders every React fixture row at the exact React
// row size so the output can be diffed against
// fyneui/testdata/react/chatrow_dark_400x700.png.
func TestChatRowFixtureGolden(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()

	chats, unread := reactFixture()
	s := &Sidebar{
		rowRefs:   make(map[fyne.CanvasObject]*chatRowRefs),
		unread:    unread,
		avatarRes: map[string]fyne.Resource{},
		visible:   chats,
	}

	for i := range chats {
		row := s.newRow()
		s.updateRow(i, row)
		slug := "chatrow_fixture_" + string(rune('a'+i))
		img := renderPNG(t, slug, row, fyne.NewSize(reactRowWidth, reactRowHeight))
		assertNotBlank(t, img, 0.03)
	}
}
