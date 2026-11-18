package tui

import (
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/Gitguy268/spot-cli/internal/pet"
	"github.com/Gitguy268/spot-cli/internal/store"
)

var noon = time.Date(2026, 11, 18, 12, 0, 0, 0, time.Local)

func newTest(t *testing.T) (*model, *store.Store) {
	t.Helper()
	s := &store.Store{Dir: t.TempDir()}
	m, err := newModel(s, store.Config{Reactions: "full", Theme: "dark", Prompt: "face"}, func() time.Time { return noon })
	if err != nil {
		t.Fatal(err)
	}
	m.Update(tea.WindowSizeMsg{Width: 100, Height: 30})
	return m, s
}

func key(k string) tea.KeyPressMsg {
	switch k {
	case "space":
		return tea.KeyPressMsg{Code: tea.KeySpace, Text: " "}
	case "esc":
		return tea.KeyPressMsg{Code: tea.KeyEscape}
	}
	r := []rune(k)[0]
	return tea.KeyPressMsg{Code: r, Text: k}
}

func ticks(m *model, n int) {
	for range n {
		m.Update(tickMsg{})
	}
}

func TestPetFeedAndSave(t *testing.T) {
	m, s := newTest(t)
	bond := m.st.Bond
	m.Update(key("p"))
	if m.anim != "love" || m.st.Bond <= bond || len(m.parts) == 0 {
		t.Fatalf("pet: anim %q bond %.0f→%.0f, %d hearts", m.anim, bond, m.st.Bond, len(m.parts))
	}
	m.Update(key("f"))
	if m.anim != "eat" {
		t.Fatalf("feed: anim %q", m.anim)
	}
	saved, _ := s.Load(noon)
	if saved.Pets != 1 || saved.Treats != 2 {
		t.Fatalf("not saved: pets %d treats %d", saved.Pets, saved.Treats)
	}
}

func TestFetchRound(t *testing.T) {
	m, s := newTest(t)
	m.Update(key("b"))
	if m.mode != fetching {
		t.Fatal("b should start fetch")
	}
	for m.fetch.u < catchFrom+0.05 {
		ticks(m, 1)
	}
	m.Update(key("space"))
	if m.fetch.catches != 1 {
		t.Fatalf("catches %d", m.fetch.catches)
	}
	for i := 0; m.mode == fetching && i < 2000; i++ {
		ticks(m, 1) // let the rest fly past: three misses end the round
	}
	if m.mode != room {
		t.Fatal("round never ended")
	}
	saved, _ := s.Load(noon)
	if saved.FetchBest != 1 {
		t.Fatalf("fetch best %d", saved.FetchBest)
	}
}

func TestEarlyPressIsAMiss(t *testing.T) {
	m, _ := newTest(t)
	m.Update(key("b"))
	ticks(m, 1)
	m.Update(key("space"))
	if m.fetch.misses != 1 || m.fetch.phase != fumbled {
		t.Fatalf("early press: %+v", m.fetch)
	}
	m.Update(key("esc"))
	if m.mode != room {
		t.Fatal("esc should leave fetch")
	}
}

func TestUnknownTrickTilts(t *testing.T) {
	m, _ := newTest(t)
	m.Update(key("t"))
	m.Update(key("1"))
	if m.anim != "tilt" || !strings.Contains(m.caption, "doesn't know") {
		t.Fatalf("anim %q caption %q", m.anim, m.caption)
	}
}

func TestKnownTrickPlays(t *testing.T) {
	m, _ := newTest(t)
	m.st.Tricks = []string{"paw"}
	m.Update(key("t"))
	m.Update(key("1"))
	if m.anim != "paw" {
		t.Fatalf("anim %q", m.anim)
	}
}

func TestSeesTestsFromOtherTerminals(t *testing.T) {
	m, s := newTest(t)
	ticks(m, 40) // let the greeting finish
	s.Update(noon.Add(time.Second), func(st *pet.State) error {
		st.TestFailed("go test ./...", 0, noon.Add(time.Second))
		return nil
	})
	ticks(m, 2*fps)
	if m.anim != "sad" {
		t.Fatalf("after a red run elsewhere: anim %q", m.anim)
	}
	ticks(m, 60)
	s.Update(noon.Add(2*time.Second), func(st *pet.State) error {
		st.TestPassed("go test ./...", 0, noon.Add(2*time.Second))
		return nil
	})
	ticks(m, 2*fps)
	if m.anim != "zoomies" {
		t.Fatalf("red then green elsewhere should be zoomies, got %q", m.anim)
	}
}

func TestViewFits(t *testing.T) {
	for _, size := range [][2]int{{100, 30}, {80, 24}, {60, 20}, {40, 12}} {
		m, _ := newTest(t)
		m.Update(tea.WindowSizeMsg{Width: size[0], Height: size[1]})
		ticks(m, 3)
		for _, step := range []string{"", "p", "b"} {
			if step != "" {
				m.Update(key(step))
				ticks(m, 5)
			}
			lines := strings.Split(m.View().Content, "\n")
			if len(lines) > size[1] {
				t.Errorf("%dx%d after %q: %d lines", size[0], size[1], step, len(lines))
			}
			for i, l := range lines {
				if w := lipgloss.Width(l); w > size[0] {
					t.Errorf("%dx%d after %q: line %d is %d wide", size[0], size[1], step, i, w)
				}
			}
		}
	}
}

func TestQuitSaves(t *testing.T) {
	m, s := newTest(t)
	_, cmd := m.Update(key("q"))
	if cmd == nil {
		t.Fatal("q should quit")
	}
	if _, err := s.Load(noon); err != nil {
		t.Fatal(err)
	}
}

func TestClickOnSpotPets(t *testing.T) {
	m, _ := newTest(t)
	ox, oy := m.sceneOrigin()
	m.Update(tea.MouseClickMsg{X: ox + spotX + 16, Y: oy + (spotY+14)/2})
	if m.st.Pets != 1 {
		t.Fatal("click on Spot didn't pet him")
	}
	m.Update(tea.MouseClickMsg{X: 0, Y: 0})
	if m.st.Pets != 1 {
		t.Fatal("click in the corner petted him")
	}
}
