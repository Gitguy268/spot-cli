package tui

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/Gitguy268/spot-cli/internal/art"
	"github.com/Gitguy268/spot-cli/internal/pet"
)

var (
	nameStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#ff6b6b"))
	dimStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#8a90b8"))
	textStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#d8dcf5"))
	goldStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#ffc94a"))
	keyStyle  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#ffc94a"))
	barColour = map[string]string{"food": "#ffb35c", "joy": "#ff6f9c", "energy": "#7fb2ff", "bond": "#c4f05a"}
)

var moon = art.Sprite{".SS.", "SS..", "SS..", ".SS."}

var moodAnim = map[pet.Mood]string{
	pet.Asleep: "sleep", pet.Sleepy: "sleepy", pet.Hungry: "beg", pet.Sad: "sad",
	pet.Proud: "love", pet.Happy: "idle", pet.Content: "idle",
}

var moodWords = map[pet.Mood]string{
	pet.Asleep: "fast asleep", pet.Hungry: "hungry", pet.Sad: "a bit blue", pet.Proud: "proud of you",
	pet.Happy: "happy", pet.Content: "content", pet.Sleepy: "sleepy",
}

func (m *model) View() tea.View {
	v := tea.NewView(m.content())
	v.AltScreen = true
	v.MouseMode = tea.MouseModeCellMotion
	v.WindowTitle = m.st.Name + " 🐾"
	return v
}

func (m *model) content() string {
	if m.quitting {
		return ""
	}
	if m.w < sceneW+2 || m.h < 23 {
		return m.compact()
	}
	var scene art.Sprite
	if m.mode == fetching {
		scene = m.fetchScene()
	} else {
		scene = m.scene()
	}
	pic := strings.Join(art.Render(scene, m.pal, m.profile), "\n")
	lines := []string{m.header(), m.bars(), m.footer()}
	block := lipgloss.JoinVertical(lipgloss.Left, pic, strings.Join(lines, "\n"))
	return lipgloss.Place(m.w, m.h, lipgloss.Center, lipgloss.Center, block)
}

// sceneOrigin is where the scene's top-left cell lands on screen.
func (m *model) sceneOrigin() (int, int) {
	return max(0, (m.w-sceneW)/2), max(0, (m.h-(sceneH/2+5))/2)
}

func (m *model) current() string {
	if m.busy() {
		return m.anim
	}
	return moodAnim[m.mood()]
}

func (m *model) scene() art.Sprite {
	c := art.NewCanvas(sceneW, sceneH)
	night := pet.Night(m.now())
	if night {
		c.Draw(moon, 3, 1)
		for i, p := range [][2]int{{12, 2}, {56, 3}, {48, 1}, {22, 5}} {
			if (m.tick/6+i)%5 != 0 {
				c.Draw(art.Sprite{"S"}, p[0], p[1])
			}
		}
	}
	// rug
	c.Ellipse(32, 33.5, 26, 3.6, 'D')
	for x := 10; x < 55; x += 6 {
		c.Draw(art.Sprite{"d", "d"}, x, 32)
	}
	name := m.current()
	a := art.Get(name)
	if name != "eat" && name != "zoomies" {
		bowl := art.Bowl
		if m.st.Food >= 60 {
			bowl = art.BowlFull
		}
		c.Draw(bowl, 49, 27)
	}
	if m.mode != fetching {
		c.Draw(art.Ball, 6, 30)
	}
	f := a.Frame(m.tick, fps)
	switch name {
	case "zoomies":
		c.Draw(f, int(m.runX), spotY+8)
	case "run":
		c.Draw(f, spotX, spotY+8)
	default:
		c.Draw(f, spotX, spotY)
	}
	if name == "sleep" {
		c.Draw(art.Blanket, spotX+1, spotY+22)
	}
	for _, p := range m.parts {
		c.Draw(p.s, int(p.x), int(p.y))
	}
	return c.Sprite()
}

func (m *model) header() string {
	st := m.st
	mood := m.mood()
	left := nameStyle.Render(st.Name) + dimStyle.Render(" is ") + textStyle.Render(moodWords[mood])
	right := goldStyle.Render(fmt.Sprintf("🦴 %d", st.Treats)) + dimStyle.Render(fmt.Sprintf("   streak %d   best %d", st.Streak, st.BestStreak))
	gap := max(1, sceneW-lipgloss.Width(left)-lipgloss.Width(right))
	return left + strings.Repeat(" ", gap) + right
}

func (m *model) bar(label string, v float64) string {
	n := int(v/10 + .5)
	fill := lipgloss.NewStyle().Foreground(lipgloss.Color(barColour[label])).Render(strings.Repeat("█", n))
	return dimStyle.Render(fmt.Sprintf("%-7s", label)) + fill + dimStyle.Render(strings.Repeat("░", 10-n)) + "   "
}

func (m *model) bars() string {
	st := m.st
	return m.bar("food", st.Food) + m.bar("joy", st.Joy) + "\n" + m.bar("energy", st.Energy) + m.bar("bond", st.Bond)
}

func (m *model) footer() string {
	var caption string
	switch {
	case m.err != "":
		caption = lipgloss.NewStyle().Foreground(lipgloss.Color("#ff6b6b")).Render(m.err)
	case m.mode == tricks:
		caption = m.trickMenu()
	case m.mode == fetching:
		caption = m.fetchCaption()
	case m.tick < m.captUntil:
		caption = textStyle.Render(m.caption)
	}
	caption = truncate(caption, sceneW)
	keys := []string{"p", "pet", "f", "feed", "b", "fetch", "t", "tricks", "?", "help", "q", "quit"}
	if m.mode == fetching {
		keys = []string{"space", "catch", "esc", "stop"}
	}
	var k []string
	for i := 0; i < len(keys); i += 2 {
		k = append(k, keyStyle.Render(keys[i])+dimStyle.Render(" "+keys[i+1]))
	}
	return caption + "\n" + strings.Join(k, dimStyle.Render("  ·  "))
}

func (m *model) trickMenu() string {
	var parts []string
	for i, t := range pet.Tricks {
		label := fmt.Sprintf("%d %s", i+1, t.Name)
		if m.st.Knows(t.ID) {
			parts = append(parts, goldStyle.Render(label))
		} else {
			parts = append(parts, dimStyle.Render(label))
		}
	}
	return strings.Join(parts, dimStyle.Render(" · "))
}

// compact is the room for small terminals: little Spot and the stats.
func (m *model) compact() string {
	a := art.Get("mini-idle")
	switch m.current() {
	case "wag", "zoomies", "paw", "spin", "roll", "wow":
		a = art.Get("mini-wag")
	case "love", "eat":
		a = art.Get("mini-love")
	case "sad", "beg":
		a = art.Get("mini-sad")
	case "sleep", "sleepy":
		a = art.Get("mini-sleep")
	}
	pic := art.Render(a.Frame(m.tick, fps), m.pal, m.profile)
	st := m.st
	text := []string{
		nameStyle.Render(st.Name) + dimStyle.Render(" is ") + textStyle.Render(moodWords[m.mood()]),
		dimStyle.Render(fmt.Sprintf("food %3.0f  joy %3.0f", st.Food, st.Joy)),
		dimStyle.Render(fmt.Sprintf("energy %3.0f  bond %3.0f", st.Energy, st.Bond)),
		goldStyle.Render(fmt.Sprintf("🦴 %d", st.Treats)) + dimStyle.Render(fmt.Sprintf("  streak %d", st.Streak)),
	}
	room := m.w - 26 // the picture and the gap next to it
	if m.w < 50 {
		room = m.w - 2
	}
	if m.tick < m.captUntil {
		text = append(text, textStyle.Render(truncate(m.caption, max(10, room))))
	}
	text = append(text, keyStyle.Render("p")+dimStyle.Render(" pet ")+keyStyle.Render("f")+dimStyle.Render(" feed ")+keyStyle.Render("q")+dimStyle.Render(" quit"))
	block := strings.Join(text, "\n")
	if m.w >= 50 && m.h >= 8 {
		block = lipgloss.JoinHorizontal(lipgloss.Center, strings.Join(pic, "\n"), "  ", block)
	}
	return lipgloss.Place(m.w, m.h, lipgloss.Center, lipgloss.Center, block)
}

func truncate(s string, w int) string {
	if lipgloss.Width(s) <= w {
		return s
	}
	r := []rune(s)
	for len(r) > 0 && lipgloss.Width(string(r)) > w-1 {
		r = r[:len(r)-1]
	}
	return string(r) + "…"
}
