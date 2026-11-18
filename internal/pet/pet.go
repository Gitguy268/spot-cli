// Package pet is Spot's brain: his stats, how they change over time, and what
// he does when your tests pass or fail. Nothing here touches the disk or the
// terminal, and every function takes the time as an argument.
package pet

import (
	"math"
	"strings"
	"time"
)

// Version is the state file format.
const Version = 1

// State is everything Spot remembers.
type State struct {
	Version int       `json:"version"`
	Name    string    `json:"name"`
	Born    time.Time `json:"born"`
	Updated time.Time `json:"updated"`

	// Stats run from 0 to 100.
	Food   float64 `json:"food"`
	Joy    float64 `json:"joy"`
	Energy float64 `json:"energy"`
	Bond   float64 `json:"bond"`

	Treats     int       `json:"treats"`
	KibbleAt   time.Time `json:"kibbleAt"` // last free bowl of kibble
	Streak     int       `json:"streak"`
	BestStreak int       `json:"bestStreak"`
	Runs       int       `json:"runs"`
	Passes     int       `json:"passes"`
	Fails      int       `json:"fails"`
	Pets       int       `json:"pets"`
	FetchBest  int       `json:"fetchBest"`

	Day        string `json:"day"` // local date of PassesDay
	PassesDay  int    `json:"passesDay"`
	GreenDay   string `json:"greenDay"` // last local date with a pass
	GreenDays  int    `json:"greenDays"`
	BondDay    string `json:"bondDay"`
	BondGained int    `json:"bondGained"` // bond from petting today

	// Last result per command, so red then green can be a comeback.
	LastResult map[string]bool `json:"lastResult,omitempty"`
	Last       *Run            `json:"last,omitempty"`

	Tricks []string             `json:"tricks,omitempty"`
	Badges map[string]time.Time `json:"badges,omitempty"`
	Log    []Moment             `json:"log,omitempty"`
}

// Run is one test run Spot saw.
type Run struct {
	Cmd string        `json:"cmd"`
	OK  bool          `json:"ok"`
	At  time.Time     `json:"at"`
	Dur time.Duration `json:"dur"`
}

// Moment is a line in Spot's diary.
type Moment struct {
	At   time.Time `json:"at"`
	Text string    `json:"text"`
}

// New returns a puppy called name, fed and happy.
func New(name string, now time.Time) *State {
	if strings.TrimSpace(name) == "" {
		name = "Spot"
	}
	return &State{
		Version: Version, Name: name, Born: now, Updated: now,
		Food: 80, Joy: 80, Energy: 90, Bond: 10, Treats: 3,
		LastResult: map[string]bool{}, Badges: map[string]time.Time{},
	}
}

// Rates per hour.
const (
	foodPerHour      = 8   // full to hungry in about 8 hours
	joyPerHour       = 2.5 // he misses you a little
	hungryJoyPerHour = 4   // and more when his tummy rumbles
	energyPerHour    = 4
	restPerHour      = 14 // recovery while it's night
)

// Night is when Spot sleeps if nothing is going on.
func Night(t time.Time) bool { h := t.Hour(); return h >= 23 || h < 7 }

// Tick moves Spot's stats forward to now. Time is stepped in 10-minute
// slices so a night in the middle counts as rest. Only the last two weeks
// count: after that long he just waits by the door.
func (s *State) Tick(now time.Time) {
	if s.Updated.IsZero() {
		s.Updated = now
	}
	if !now.After(s.Updated) {
		return
	}
	t := s.Updated
	if now.Sub(t) > 14*24*time.Hour {
		t = now.Add(-14 * 24 * time.Hour)
	}
	for t.Before(now) {
		step := min(10*time.Minute, now.Sub(t))
		h := step.Hours()
		s.Food -= foodPerHour * h
		if s.Food < 35 {
			s.Joy -= hungryJoyPerHour * h
		} else {
			s.Joy -= joyPerHour * h
		}
		if Night(t) {
			s.Energy += restPerHour * h
		} else {
			s.Energy -= energyPerHour * h
		}
		s.clamp()
		t = t.Add(step)
	}
	s.Updated = now
}

func (s *State) clamp() {
	c := func(v, lo float64) float64 { return math.Max(lo, math.Min(100, v)) }
	s.Food = c(s.Food, 0)
	s.Joy = c(s.Joy, 10) // never hopeless
	s.Energy = c(s.Energy, 5)
	s.Bond = c(s.Bond, 0)
}

// Mood is what Spot looks like right now.
type Mood string

const (
	Asleep  Mood = "asleep"
	Hungry  Mood = "hungry"
	Sad     Mood = "sad"
	Proud   Mood = "proud" // just saw green tests
	Happy   Mood = "happy"
	Content Mood = "content"
	Sleepy  Mood = "sleepy"
)

// MoodAt works out Spot's mood. Call Tick first.
func (s *State) MoodAt(now time.Time) Mood {
	recent := s.Last != nil && now.Sub(s.Last.At) < 5*time.Minute
	switch {
	case recent && s.Last.OK:
		return Proud
	case recent && !s.Last.OK:
		return Sad
	case Night(now) && (s.Last == nil || now.Sub(s.Last.At) > 20*time.Minute):
		return Asleep
	case s.Food < 35:
		return Hungry
	case s.Energy < 20:
		return Sleepy
	case s.Joy < 35:
		return Sad
	case s.Joy >= 70:
		return Happy
	}
	return Content
}

func (s *State) note(now time.Time, text string) {
	s.Log = append(s.Log, Moment{At: now, Text: text})
	if len(s.Log) > 30 {
		s.Log = s.Log[len(s.Log)-30:]
	}
}

func day(t time.Time) string { return t.Format("2006-01-02") }

// PassRate is passes over runs, 0 when nothing ran yet.
func (s *State) PassRate() float64 {
	if s.Runs == 0 {
		return 0
	}
	return float64(s.Passes) / float64(s.Runs)
}
