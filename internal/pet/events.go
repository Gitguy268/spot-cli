package pet

import (
	"fmt"
	"strings"
	"time"
)

// Reaction is what Spot does in answer to something. Anim names an
// animation in package art.
type Reaction struct {
	Anim      string
	Text      string
	Treats    int // treats earned
	NewTricks []Trick
	NewBadges []Badge
	Comeback  bool
}

// pick chooses a line from a list, varying with the time.
func pick(now time.Time, lines ...string) string {
	return lines[int(now.UnixMilli()%int64(len(lines)))]
}

// key is how a command is remembered for comebacks: its words, squashed.
func key(cmd string) string { return strings.Join(strings.Fields(cmd), " ") }

// TestPassed records a green run.
func (s *State) TestPassed(cmd string, dur time.Duration, now time.Time) Reaction {
	s.Tick(now)
	k := key(cmd)
	prev, seen := s.LastResult[k]
	wasRed := seen && !prev
	s.record(k, cmd, true, dur, now)

	s.Passes++
	s.Streak++
	s.BestStreak = max(s.BestStreak, s.Streak)
	if s.Day != day(now) {
		s.Day, s.PassesDay = day(now), 0
	}
	s.PassesDay++
	switch s.GreenDay {
	case day(now):
	case day(now.AddDate(0, 0, -1)):
		s.GreenDays++
		s.GreenDay = day(now)
	default:
		s.GreenDays = 1
		s.GreenDay = day(now)
	}

	r := Reaction{Anim: "wag", Treats: 1}
	s.Food += 10
	s.Joy += 12
	s.Bond += 0.5
	if wasRed {
		r.Comeback = true
		r.Anim = "zoomies"
		r.Treats = 2
		s.Joy += 15
		r.Text = pick(now,
			"red to green! %s does zoomies around the room",
			"you fixed it! %s is running laps",
			"%s knew you could do it. ZOOMIES.",
		)
	} else {
		r.Text = pick(now,
			"%s wags his tail",
			"%s's whole back half is wagging",
			"%s does a happy little hop",
			"%s brings you his ball to celebrate",
			"%s thinks you're a very good developer",
			"%s spins in a circle",
		)
	}
	s.Treats += r.Treats
	s.Energy--
	s.clamp()
	r.Text = fmt.Sprintf(r.Text, s.Name)
	r.NewTricks = s.learn(now)
	r.NewBadges = s.award(now, dur, r.Comeback)
	s.note(now, "tests passed: "+short(cmd))
	return r
}

// TestFailed records a red run. Spot is sad for you, not with you.
func (s *State) TestFailed(cmd string, dur time.Duration, now time.Time) Reaction {
	s.Tick(now)
	s.record(key(cmd), cmd, false, dur, now)
	s.Fails++
	s.Streak = 0
	s.Joy -= 6
	s.clamp()
	r := Reaction{Anim: "sad"}
	r.Text = fmt.Sprintf(pick(now,
		"%s puts his head on your knee",
		"%s nudges your hand. you'll get it",
		"%s brings you a sock for comfort",
		"%s's ears droop. he believes in you",
		"%s sits close. red is just a colour",
	), s.Name)
	r.NewBadges = s.award(now, dur, false)
	s.note(now, "tests failed: "+short(cmd))
	return r
}

func (s *State) record(k, cmd string, ok bool, dur time.Duration, now time.Time) {
	if s.LastResult == nil {
		s.LastResult = map[string]bool{}
	}
	if _, seen := s.LastResult[k]; !seen && len(s.LastResult) >= 64 {
		for old := range s.LastResult { // forget one; any will do
			delete(s.LastResult, old)
			break
		}
	}
	s.LastResult[k] = ok
	s.Runs++
	s.Last = &Run{Cmd: short(cmd), OK: ok, At: now, Dur: dur}
}

func short(cmd string) string {
	cmd = key(cmd)
	if r := []rune(cmd); len(r) > 60 {
		return string(r[:59]) + "…"
	}
	return cmd
}

// FreeKibbleEvery is how often the bowl refills for free.
const FreeKibbleEvery = 4 * time.Hour

// Feed gives Spot a treat, or his free kibble when there are no treats.
func (s *State) Feed(now time.Time) Reaction {
	s.Tick(now)
	switch {
	case s.Food >= 95:
		return Reaction{Anim: "tilt", Text: s.Name + " sniffs the bowl and looks at you. he's full!"}
	case s.Treats > 0:
		s.Treats--
		s.Food += 35
		s.Joy += 6
		s.clamp()
		s.note(now, "had a treat")
		return Reaction{Anim: "eat", Text: fmt.Sprintf("nom nom nom · %d %s left", s.Treats, plural(s.Treats, "treat"))}
	case now.Sub(s.KibbleAt) >= FreeKibbleEvery:
		s.KibbleAt = now
		s.Food += 25
		s.Joy += 3
		s.clamp()
		s.note(now, "had his kibble")
		return Reaction{Anim: "eat", Text: "crunch crunch · pass some tests to earn treats"}
	}
	wait := FreeKibbleEvery - now.Sub(s.KibbleAt)
	return Reaction{Anim: "beg", Text: fmt.Sprintf("no treats left. free kibble in %s, or pass a test", roundDur(wait))}
}

// Pet scratches Spot behind the ears.
func (s *State) Pet(now time.Time) Reaction {
	s.Tick(now)
	s.Pets++
	if s.BondDay != day(now) {
		s.BondDay, s.BondGained = day(now), 0
	}
	if s.BondGained < 10 {
		s.Bond += 2
		s.BondGained += 2
	}
	s.Joy += 5
	s.clamp()
	r := Reaction{Anim: "love", Text: fmt.Sprintf(pick(now,
		"%s leans into the scratches",
		"%s's back leg thumps",
		"%s melts. 10/10 ear scratch",
		"%s gives you a slow blink",
		"%s rolls his head into your hand",
	), s.Name)}
	r.NewTricks = s.learn(now)
	r.NewBadges = s.award(now, 0, false)
	return r
}

// Fetched records a round of fetch with this many catches.
func (s *State) Fetched(catches int, now time.Time) Reaction {
	s.Tick(now)
	s.FetchBest = max(s.FetchBest, catches)
	s.Joy += 10
	s.Energy -= 10
	s.Bond++
	s.clamp()
	s.note(now, fmt.Sprintf("played fetch: %d %s", catches, plural(catches, "catch")))
	r := Reaction{Anim: "wag", Text: fmt.Sprintf("%d %s! %s drops the ball at your feet", catches, plural(catches, "catch"), s.Name)}
	r.NewTricks = s.learn(now)
	r.NewBadges = s.award(now, 0, false)
	return r
}

// Rename changes Spot's name (he answers to anything said kindly).
func (s *State) Rename(name string) {
	if n := strings.TrimSpace(name); n != "" {
		s.Name = n
	}
}

func plural(n int, word string) string {
	if n == 1 {
		return word
	}
	if strings.HasSuffix(word, "ch") {
		return word + "es"
	}
	return word + "s"
}

func roundDur(d time.Duration) string {
	if d < time.Minute {
		return "a minute"
	}
	h, m := int(d.Hours()), int(d.Minutes())%60
	if h == 0 {
		return fmt.Sprintf("%dm", m)
	}
	return fmt.Sprintf("%dh%02dm", h, m)
}
