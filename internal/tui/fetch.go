package tui

import (
	"fmt"
	"math"

	tea "charm.land/bubbletea/v2"

	"github.com/Gitguy268/spot-cli/internal/art"
	"github.com/Gitguy268/spot-cli/internal/pet"
)

// Fetch: you throw, the ball arcs over to Spot, and you press space as it
// reaches his mouth. Three misses end the round. Every catch makes the next
// throw a little faster.

type fetchPhase int

const (
	flying fetchPhase = iota
	caught
	fumbled
	chasing
)

type fetchGame struct {
	phase    fetchPhase
	u        float64 // how far along the throw, 0 to 1
	catches  int
	misses   int
	until    int // tick when the current phase ends
	throwSec float64
}

const (
	fetchSpotX = 30
	catchFrom  = 0.8
	maxMisses  = 3
)

func (m *model) startFetch() {
	m.mode = fetching
	m.fetch = fetchGame{throwSec: 1.6}
	m.anim = ""
	m.say("", 0)
}

func (m *model) fetchStep() {
	g := &m.fetch
	switch g.phase {
	case flying:
		g.u += 1 / (g.throwSec * fps)
		if g.u > 1.1 {
			g.misses++
			g.phase, g.until = chasing, m.tick+30
		}
	default:
		if m.tick >= g.until {
			if g.misses >= maxMisses {
				m.endFetch()
				return
			}
			g.phase, g.u = flying, 0
		}
	}
}

func (m *model) fetchKey(k string) (tea.Model, tea.Cmd) {
	g := &m.fetch
	switch k {
	case "esc", "q":
		m.endFetch()
	case "space", " ", "enter":
		if g.phase != flying {
			break
		}
		if g.u >= catchFrom {
			g.catches++
			g.throwSec = math.Max(0.7, g.throwSec-0.08)
			g.phase, g.until = caught, m.tick+10
		} else {
			g.misses++
			g.phase, g.until = fumbled, m.tick+14
		}
	}
	return m, nil
}

func (m *model) endFetch() {
	m.mode = room
	if m.fetch.catches == 0 && m.fetch.misses == 0 {
		return
	}
	n := m.fetch.catches
	m.act(func(st *pet.State) pet.Reaction { return st.Fetched(n, m.now()) })
}

// ballAt is the ball's position at u along its arc, in scene pixels.
func ballAt(u float64) (int, int) {
	u = clamp01(u)
	x := 2 + u*float64(fetchSpotX+12)
	y := 30 - u*12 - 16*math.Sin(math.Pi*u)
	return int(x), int(y)
}

func (m *model) fetchScene() art.Sprite {
	c := art.NewCanvas(sceneW, sceneH)
	c.Ellipse(32, 33.5, 26, 3.6, 'D')
	g := m.fetch
	switch g.phase {
	case flying:
		c.Draw(art.Get("idle").Frames[0], fetchSpotX, spotY)
		if g.u >= catchFrom {
			c.Draw(art.Get("wow").Frames[0], fetchSpotX, spotY)
		}
		x, y := ballAt(g.u)
		c.Draw(art.Ball, x, y)
	case caught:
		c.Draw(art.Get("wag").Frame(m.tick, fps), fetchSpotX, spotY)
		c.Draw(art.Ball, fetchSpotX+14, spotY+15)
		c.Draw(art.Sparkle, fetchSpotX+4, spotY+2)
		c.Draw(art.Sparkle, fetchSpotX+26, spotY+4)
	case fumbled:
		c.Draw(art.Get("tilt").Frames[0], fetchSpotX, spotY)
		left := g.until - m.tick
		c.Draw(art.Ball, fetchSpotX+10-(14-left), 30-left/3)
	case chasing:
		left := g.until - m.tick // 30 → 0: run out to the right, then back
		run := art.Get("run").Frame(m.tick, fps)
		if left > 15 {
			c.Draw(run, fetchSpotX+(30-left)*3, spotY+8)
		} else {
			c.Draw(run.Mirror(), fetchSpotX+left*3, spotY+8)
			c.Draw(art.Ball, fetchSpotX+left*3+1, spotY+20)
		}
	}
	return c.Sprite()
}

func (m *model) fetchCaption() string {
	g := m.fetch
	score := goldStyle.Render(fmt.Sprintf("🎾 %d", g.catches)) + dimStyle.Render(fmt.Sprintf("   misses %d/%d", g.misses, maxMisses))
	var say string
	switch g.phase {
	case flying:
		say = "press space when the ball reaches " + m.st.Name
	case caught:
		say = "caught it!"
	case fumbled:
		say = "too early, it bonked his nose"
	case chasing:
		say = m.st.Name + " chases it down and brings it back"
	}
	return score + "   " + textStyle.Render(say)
}
