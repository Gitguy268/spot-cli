// Package tui is Spot's room: the full-screen app.
//
// Spot sits on his rug with his bowl and ball. You can pet him (p or a
// click), feed him (f), play fetch (b) and ask for tricks (t). The room
// re-reads the state file every two seconds, so when tests pass or fail in
// another terminal Spot reacts here too.
package tui

import (
	"errors"
	"fmt"
	"math"
	"os"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/colorprofile"

	"github.com/Gitguy268/spot-cli/internal/art"
	"github.com/Gitguy268/spot-cli/internal/pet"
	"github.com/Gitguy268/spot-cli/internal/store"
)

const fps = 12

// Scene size in pixels; it renders as sceneW columns by sceneH/2 rows.
const (
	sceneW = 64
	sceneH = 36
	spotX  = 16
	spotY  = 4
)

type mode int

const (
	room mode = iota
	tricks
	fetching
)

type particle struct {
	x, y, vx, vy float64
	s            art.Sprite
	life         int
}

type model struct {
	store *store.Store
	cfg   store.Config
	st    *pet.State
	now   func() time.Time

	w, h    int
	pal     art.Palette
	sty     styles
	profile colorprofile.Profile
	tick    int

	anim      string // an action playing on top of the mood
	animUntil int    // tick when it ends
	runX      float64
	caption   string
	captUntil int
	seen      time.Time // last test run already reacted to
	parts     []particle
	mode      mode
	fetch     fetchGame
	err       string
	quitting  bool
}

type tickMsg struct{}

func tickCmd() tea.Cmd {
	return tea.Tick(time.Second/fps, func(time.Time) tea.Msg { return tickMsg{} })
}

// Run opens the room and returns an exit code.
func Run(s *store.Store, cfg store.Config) int {
	m, err := newModel(s, cfg, time.Now)
	if err != nil {
		fmt.Fprintln(os.Stderr, "spot:", err)
		return 1
	}
	if _, err := tea.NewProgram(m).Run(); err != nil {
		fmt.Fprintln(os.Stderr, "spot:", err)
		return 1
	}
	return 0
}

func newModel(s *store.Store, cfg store.Config, now func() time.Time) (*model, error) {
	st, err := s.Load(now())
	if st == nil {
		return nil, err
	}
	m := &model{store: s, cfg: cfg, st: st, now: now, w: 80, h: 24, pal: art.Dark, profile: colorprofile.TrueColor}
	if errors.Is(err, store.ErrCorrupt) {
		m.say("Spot's old memories were unreadable, so he starts over as a puppy", 60)
	}
	m.sty = newStyles(false)
	if cfg.Theme == "light" {
		m.pal, m.sty = art.Light, newStyles(true)
	}
	st.Tick(now())
	if st.Last != nil {
		m.seen = st.Last.At
	}
	m.greet()
	return m, nil
}

func (m *model) greet() {
	n := m.st.Name
	switch m.st.MoodAt(m.now()) {
	case pet.Asleep:
		m.say(n+" is asleep under his blanket. press p to wake him gently", 72)
	case pet.Hungry:
		m.say(n+" sits by his empty bowl and looks at you. f to feed", 72)
	case pet.Sad:
		m.say(n+" could use some company. p for pets", 60)
	default:
		m.play("wag", 18)
		m.say(n+" is happy to see you!", 48)
	}
}

func (m *model) Init() tea.Cmd {
	cmds := []tea.Cmd{tickCmd()}
	if m.cfg.Theme == "auto" {
		cmds = append(cmds, tea.RequestBackgroundColor)
	}
	return tea.Batch(cmds...)
}

func (m *model) say(s string, ticks int) { m.caption, m.captUntil = s, m.tick+ticks }

func (m *model) play(anim string, ticks int) {
	m.anim, m.animUntil = anim, m.tick+ticks
	if anim == "zoomies" {
		m.runX = -32
	}
}

func (m *model) busy() bool { return m.anim != "" && m.tick < m.animUntil }

// act runs a state change through the store so changes made by shells in
// the meantime are kept.
func (m *model) act(fn func(*pet.State) pet.Reaction) pet.Reaction {
	var r pet.Reaction
	st, err := m.store.Update(m.now(), func(st *pet.State) error { r = fn(st); return nil })
	if err != nil && !errors.Is(err, store.ErrCorrupt) {
		m.err = err.Error()
		return r
	}
	m.st = st
	if st.Last != nil {
		m.seen = st.Last.At
	}
	m.react(r)
	return r
}

func (m *model) react(r pet.Reaction) {
	length := map[string]int{"wag": 24, "love": 30, "eat": 36, "sad": 36, "zoomies": 40, "beg": 30, "tilt": 24}[r.Anim]
	if length == 0 {
		length = 24
	}
	m.play(r.Anim, length)
	text := r.Text
	for _, t := range r.NewTricks {
		text += " · ✦ learned " + t.Name + "!"
	}
	for _, b := range r.NewBadges {
		text += " · " + b.Icon + " " + b.Name
	}
	m.say(text, max(length, 60))
	switch r.Anim {
	case "love":
		for i := range 5 {
			m.parts = append(m.parts, particle{x: float64(spotX + 3 + i*6), y: float64(spotY - 1 - (i%2)*3), vx: (float64(i) - 2) * .15, vy: -.5 - float64(i%2)*.2, s: art.Heart, life: 30})
		}
	case "wag", "zoomies":
		for i := range 3 {
			m.parts = append(m.parts, particle{x: float64(spotX + 4 + i*10), y: spotY, vy: -.4, s: art.Sparkle, life: 16})
		}
	}
}

// sync picks up test runs recorded by the shell hook in other terminals.
func (m *model) sync() {
	st, err := m.store.Load(m.now())
	if err != nil || st == nil {
		return
	}
	st.Tick(m.now())
	m.st = st
	if st.Last == nil || !st.Last.At.After(m.seen) {
		return
	}
	m.seen = st.Last.At
	if st.Last.OK {
		if comeback(st) {
			m.react(pet.Reaction{Anim: "zoomies", Text: fmt.Sprintf("red to green! %s · %s does zoomies", st.Last.Cmd, st.Name)})
			return
		}
		m.react(pet.Reaction{Anim: "wag", Text: fmt.Sprintf("green! %s · %s wags · 🦴 %d", st.Last.Cmd, st.Name, st.Treats)})
	} else {
		m.react(pet.Reaction{Anim: "sad", Text: fmt.Sprintf("red: %s · %s puts his head on your knee", st.Last.Cmd, st.Name)})
	}
}

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.w, m.h = msg.Width, msg.Height
	case tea.ColorProfileMsg:
		m.profile = msg.Profile
	case tea.BackgroundColorMsg:
		if !msg.IsDark() {
			m.pal, m.sty = art.Light, newStyles(true)
		}
	case tickMsg:
		m.step()
		return m, tickCmd()
	case tea.MouseClickMsg:
		if m.mode == room && m.onSpot(msg.X, msg.Y) {
			m.act(func(st *pet.State) pet.Reaction { return st.Pet(m.now()) })
		}
	case tea.KeyPressMsg:
		return m.key(msg.String())
	}
	return m, nil
}

func (m *model) step() {
	m.tick++
	if m.tick%(2*fps) == 0 && m.mode == room && !m.busy() {
		m.sync()
	}
	if m.anim == "zoomies" && m.busy() {
		m.runX += 3
		if m.runX > sceneW {
			m.runX = -32
		}
	}
	if !m.busy() {
		m.anim = ""
		m.ambient()
	}
	alive := m.parts[:0]
	for _, p := range m.parts {
		p.x += p.vx
		p.y += p.vy
		p.life--
		if p.life > 0 && p.y > -6 {
			alive = append(alive, p)
		}
	}
	m.parts = alive
	if m.mode == fetching {
		m.fetchStep()
	}
}

// ambient adds the little things: Z's while he sleeps, a wag now and then.
func (m *model) ambient() {
	switch m.mood() {
	case pet.Asleep, pet.Sleepy:
		if m.tick%20 == 0 {
			s := art.SmallZ
			if m.tick%40 == 0 {
				s = art.BigZ
			}
			m.parts = append(m.parts, particle{x: spotX + 24, y: spotY + 10, vx: .25, vy: -.3, s: s, life: 40})
		}
	case pet.Happy, pet.Proud:
		if m.tick%(9*fps) == 0 {
			m.play("wag", 14)
		}
	}
}

func (m *model) mood() pet.Mood { return m.st.MoodAt(m.now()) }

func (m *model) key(k string) (tea.Model, tea.Cmd) {
	if k == "ctrl+c" {
		return m.quit()
	}
	switch m.mode {
	case tricks:
		return m.trickKey(k)
	case fetching:
		return m.fetchKey(k)
	}
	switch k {
	case "q", "esc":
		return m.quit()
	case "p", "enter":
		m.act(func(st *pet.State) pet.Reaction { return st.Pet(m.now()) })
	case "f":
		m.act(func(st *pet.State) pet.Reaction { return st.Feed(m.now()) })
	case "b", "space":
		if m.st.Energy < 15 {
			m.play("sleepy", 24)
			m.say(m.st.Name+" is too tired for fetch. let him nap a bit", 48)
			break
		}
		m.startFetch()
	case "t":
		m.mode = tricks
	case "?":
		m.play("tilt", 30)
		m.say("p pet · f feed · b fetch · t tricks · q quit · click Spot to pet him", 90)
	}
	return m, nil
}

func (m *model) trickKey(k string) (tea.Model, tea.Cmd) {
	m.mode = room
	for i, t := range pet.Tricks {
		if k == fmt.Sprint(i+1) || k == t.ID[:1] {
			if !m.st.Knows(t.ID) {
				m.play("tilt", 24)
				m.say(fmt.Sprintf("%s doesn't know %s yet. he learns it with %s", m.st.Name, t.Name, t.How), 60)
				return m, nil
			}
			m.play(t.Anim, 36)
			m.say(m.st.Name+": "+t.Name+"!", 48)
			if t.ID == "speak" {
				m.say("WOOF! "+m.st.Name+" is very proud of that", 48)
			}
			return m, nil
		}
	}
	return m, nil
}

func (m *model) quit() (tea.Model, tea.Cmd) {
	m.quitting = true
	_, _ = m.store.Update(m.now(), func(st *pet.State) error { st.Tick(m.now()); return nil })
	return m, tea.Quit
}

// onSpot reports whether a click at terminal x, y lands on Spot.
func (m *model) onSpot(x, y int) bool {
	ox, oy := m.sceneOrigin()
	px, py := x-ox, (y-oy)*2
	return px >= spotX+2 && px < spotX+30 && py >= spotY && py < spotY+28
}

// comeback is true when the diary says this command was red just before.
func comeback(st *pet.State) bool {
	var runs []string
	for _, l := range st.Log {
		if len(l.Text) > 13 && (l.Text[:13] == "tests passed:" || l.Text[:13] == "tests failed:") {
			runs = append(runs, l.Text)
		}
	}
	n := len(runs)
	return n >= 2 && runs[n-1] == "tests passed: "+st.Last.Cmd && runs[n-2] == "tests failed: "+st.Last.Cmd
}

func clamp01(v float64) float64 { return math.Max(0, math.Min(1, v)) }
