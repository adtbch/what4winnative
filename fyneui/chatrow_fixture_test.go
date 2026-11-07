package fyneui

import (
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"

	"github.com/lugvitc/whats4linux/api"
)

// TestChatRowFixtureGolden renders a row with the same fixture data the React
// harness uses, so the PNG can be compared side by side with
// fyneui/testdata/react/chatrow_dark_400x700.png.
func TestChatRowFixtureGolden(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()

	cases := []struct {
		name string
		chat api.ChatElement
	}{
		{"group", api.ChatElement{Contact: api.Contact{JID: "120363999999999999@g.us", FullName: "Vapers Semarangan 18+"}, LatestTS: 1762490000, Pinned: true}},
		{"contact", api.ChatElement{Contact: api.Contact{JID: "6282134544728@s.whatsapp.net", FullName: "cintakuuu"}, LatestTS: 1762500000}},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			s := &Sidebar{
				rowRefs:   make(map[fyne.CanvasObject]*chatRowRefs),
				unread:    map[string]int{},
				avatarRes: map[string]fyne.Resource{},
				visible:   []api.ChatElement{c.chat},
			}
			row := s.newRow()
			s.updateRow(0, row)

			img := renderPNG(t, "chatrow_fixture_"+c.name, row, fyne.NewSize(reactRowWidth, reactRowHeight))
			assertNotBlank(t, img, 0.03)
		})
	}
}
