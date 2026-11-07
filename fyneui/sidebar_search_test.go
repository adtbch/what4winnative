package fyneui

import "testing"

// TestSidebarSearchMatchesReact pins the filtering rule from
// useChatStore.ts:150-163 (useFilteredChatIds): lowercase the RAW term, match
// the chat name only. The old Fyne code trimmed the term and also matched the
// last message preview, which React never does.
func TestSidebarSearchMatchesReact(t *testing.T) {
	chats, _ := reactFixture()
	s := &Sidebar{allChats: chats, view: "chats"}

	names := func(term string) []string {
		s.search = term
		out := s.filterAndSortLocked()
		got := make([]string, 0, len(out))
		for _, c := range out {
			got = append(got, chatDisplayName(c))
		}
		return got
	}

	if got := names(""); len(got) != 3 {
		t.Fatalf("empty term returned %v, want all 3 chats", got)
	}
	if got := names("CINTA"); len(got) != 1 || got[0] != "cintakuuu" {
		t.Fatalf("case-insensitive name match = %v, want [cintakuuu]", got)
	}
	// Preview text must NOT match — React only looks at chat.name.
	if got := names("sayang"); len(got) != 0 {
		t.Fatalf("preview matched %v; React only matches names", got)
	}
	// React does not trim, so a trailing space is part of the term.
	if got := names("cintakuuu "); len(got) != 0 {
		t.Fatalf("trimmed term matched %v; React keeps the space", got)
	}
}
