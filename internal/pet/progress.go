package pet

import (
	"slices"
	"sort"
	"time"
)

// Trick is something Spot can learn.
type Trick struct {
	ID, Name, Anim, How string
	ok                  func(*State) bool
}

// Tricks in the order Spot learns them.
var Tricks = []Trick{
	{"paw", "High five", "paw", "a pass streak of 3", func(s *State) bool { return s.BestStreak >= 3 }},
	{"speak", "Speak", "wow", "a bond of 30", func(s *State) bool { return s.Bond >= 30 }},
	{"roll", "Roll over", "roll", "a pass streak of 10", func(s *State) bool { return s.BestStreak >= 10 }},
	{"spin", "Spin", "spin", "a pass streak of 25", func(s *State) bool { return s.BestStreak >= 25 }},
	{"sleep", "Play sleepy", "sleep", "a bond of 60", func(s *State) bool { return s.Bond >= 60 }},
}

// TrickByID finds a trick.
func TrickByID(id string) (Trick, bool) {
	for _, t := range Tricks {
		if t.ID == id {
			return t, true
		}
	}
	return Trick{}, false
}

// Knows reports whether Spot has learned a trick.
func (s *State) Knows(id string) bool { return slices.Contains(s.Tricks, id) }

func (s *State) learn(now time.Time) []Trick {
	var out []Trick
	for _, t := range Tricks {
		if !s.Knows(t.ID) && t.ok(s) {
			s.Tricks = append(s.Tricks, t.ID)
			s.note(now, "learned "+t.Name)
			out = append(out, t)
		}
	}
	return out
}

// Badge is a little achievement.
type Badge struct {
	ID, Icon, Name, How string
}

// Badges in display order.
var Badges = []Badge{
	{"first-green", "🌱", "First green", "pass a test run"},
	{"comeback", "🔁", "Comeback pup", "fix a red run"},
	{"ten-a-day", "🔟", "Busy day", "10 green runs in one day"},
	{"streak-25", "🔥", "On fire", "25 green runs in a row"},
	{"week", "📅", "Good week", "green runs 7 days in a row"},
	{"century", "💯", "Century", "100 green runs"},
	{"night-owl", "🦉", "Night owl", "a green run between midnight and 5"},
	{"early-bird", "🐦", "Early bird", "a green run between 5 and 7 in the morning"},
	{"patient", "⏳", "Patient pup", "a green run that took over 5 minutes"},
	{"best-friend", "💞", "Best friends", "a bond of 100"},
	{"belly-rubs", "🫶", "Belly rubs", "pet Spot 50 times"},
	{"fetch-champ", "🎾", "Fetch champ", "10 catches in one round of fetch"},
}

func (s *State) award(now time.Time, dur time.Duration, comeback bool) []Badge {
	if s.Badges == nil {
		s.Badges = map[string]time.Time{}
	}
	green := s.Last != nil && s.Last.OK && s.Last.At.Equal(now)
	h := now.Hour()
	earned := map[string]bool{
		"first-green": s.Passes >= 1,
		"comeback":    comeback,
		"ten-a-day":   s.PassesDay >= 10,
		"streak-25":   s.BestStreak >= 25,
		"week":        s.GreenDays >= 7,
		"century":     s.Passes >= 100,
		"night-owl":   green && h < 5,
		"early-bird":  green && h >= 5 && h < 7,
		"patient":     green && dur >= 5*time.Minute,
		"best-friend": s.Bond >= 100,
		"belly-rubs":  s.Pets >= 50,
		"fetch-champ": s.FetchBest >= 10,
	}
	var out []Badge
	for _, b := range Badges {
		if _, has := s.Badges[b.ID]; !has && earned[b.ID] {
			s.Badges[b.ID] = now
			s.note(now, "earned "+b.Name)
			out = append(out, b)
		}
	}
	return out
}

// EarnedBadges lists badges in the order they were earned.
func (s *State) EarnedBadges() []Badge {
	var out []Badge
	for _, b := range Badges {
		if _, ok := s.Badges[b.ID]; ok {
			out = append(out, b)
		}
	}
	sort.SliceStable(out, func(i, j int) bool { return s.Badges[out[i].ID].Before(s.Badges[out[j].ID]) })
	return out
}
