package pet

import (
	"testing"
	"time"
)

var noon = time.Date(2026, 11, 18, 12, 0, 0, 0, time.Local)

func TestTickDecaysAndClamps(t *testing.T) {
	s := New("Spot", noon)
	s.Food, s.Joy, s.Energy = 100, 100, 100
	s.Tick(noon.Add(8 * time.Hour)) // 12:00 → 20:00, all daytime
	if s.Food < 34 || s.Food > 38 {
		t.Fatalf("food after 8h = %.1f, want about 36", s.Food)
	}
	if s.Energy < 67 || s.Energy > 69 {
		t.Fatalf("energy after 8h of day = %.1f, want 68", s.Energy)
	}
	s.Tick(noon.Add(30 * 24 * time.Hour))
	if s.Food != 0 || s.Joy != 10 || s.Energy < 5 {
		t.Fatalf("after a month: food %.1f joy %.1f energy %.1f; want 0, 10 (floor), ≥5", s.Food, s.Joy, s.Energy)
	}
}

func TestNightIsRest(t *testing.T) {
	s := New("Spot", noon)
	start := time.Date(2026, 11, 18, 23, 0, 0, 0, time.Local)
	s.Updated, s.Energy = start, 20
	s.Tick(start.Add(6 * time.Hour))
	if s.Energy < 99 {
		t.Fatalf("energy after a night = %.1f, want 100", s.Energy)
	}
}

func TestTickNeverGoesBack(t *testing.T) {
	s := New("Spot", noon)
	food := s.Food
	s.Tick(noon.Add(-time.Hour))
	if s.Food != food || !s.Updated.Equal(noon) {
		t.Fatal("a clock that went backwards changed the stats")
	}
}

func TestPassStreakTreatsAndComeback(t *testing.T) {
	s := New("Spot", noon)
	treats := s.Treats
	r := s.TestPassed("go test ./...", time.Second, noon)
	if r.Anim != "wag" || r.Treats != 1 || s.Treats != treats+1 || s.Streak != 1 {
		t.Fatalf("first pass: %+v, treats %d streak %d", r, s.Treats, s.Streak)
	}
	if len(r.NewBadges) != 1 || r.NewBadges[0].ID != "first-green" {
		t.Fatalf("want first-green badge, got %+v", r.NewBadges)
	}
	s.TestPassed("go test ./...", time.Second, noon.Add(time.Minute))
	f := s.TestFailed("go   test ./...", time.Second, noon.Add(2*time.Minute))
	if f.Anim != "sad" || s.Streak != 0 {
		t.Fatalf("fail: %+v streak %d", f, s.Streak)
	}
	c := s.TestPassed("go test ./...", time.Second, noon.Add(3*time.Minute))
	if !c.Comeback || c.Anim != "zoomies" || c.Treats != 2 {
		t.Fatalf("comeback: %+v", c)
	}
	if _, ok := s.Badges["comeback"]; !ok {
		t.Fatal("no comeback badge")
	}
	// a different command failing first does not make this one a comeback
	s.TestFailed("cargo test", time.Second, noon.Add(4*time.Minute))
	if s.TestPassed("go test ./...", time.Second, noon.Add(5*time.Minute)).Comeback {
		t.Fatal("comeback across different commands")
	}
}

func TestTricksUnlockByStreak(t *testing.T) {
	s := New("Spot", noon)
	var learned []string
	for i := range 3 {
		for _, tr := range s.TestPassed("npm test", 0, noon.Add(time.Duration(i)*time.Minute)).NewTricks {
			learned = append(learned, tr.ID)
		}
	}
	if len(learned) != 1 || learned[0] != "paw" {
		t.Fatalf("after 3 passes learned %v, want [paw]", learned)
	}
}

func TestFeedUsesTreatsThenKibble(t *testing.T) {
	s := New("Spot", noon)
	s.Food, s.Treats = 20, 1
	if r := s.Feed(noon); r.Anim != "eat" || s.Treats != 0 {
		t.Fatalf("treat feed: %+v treats %d", r, s.Treats)
	}
	s.Food = 20
	if r := s.Feed(noon); r.Anim != "eat" || !s.KibbleAt.Equal(noon) {
		t.Fatalf("kibble feed: %+v", r)
	}
	s.Food = 20
	if r := s.Feed(noon.Add(time.Hour)); r.Anim != "beg" {
		t.Fatalf("empty bowl should beg, got %+v", r)
	}
	s.Food = 20
	if r := s.Feed(noon.Add(FreeKibbleEvery)); r.Anim != "eat" {
		t.Fatalf("kibble should refill after %s, got %+v", FreeKibbleEvery, r)
	}
	later := noon.Add(2 * FreeKibbleEvery)
	s.Tick(later)
	s.Food = 99
	if r := s.Feed(later); r.Anim != "tilt" {
		t.Fatalf("full Spot should tilt, got %+v", r)
	}
}

func TestPetBondIsCappedPerDay(t *testing.T) {
	s := New("Spot", noon)
	b := s.Bond
	for i := range 20 {
		s.Pet(noon.Add(time.Duration(i) * time.Second))
	}
	if s.Bond != b+10 {
		t.Fatalf("bond after 20 pets in one day = %.0f, want %.0f", s.Bond, b+10)
	}
	s.Pet(noon.Add(24 * time.Hour))
	if s.Bond != b+12 {
		t.Fatalf("bond next day = %.0f, want %.0f", s.Bond, b+12)
	}
}

func TestMoods(t *testing.T) {
	s := New("Spot", noon)
	if m := s.MoodAt(noon); m != Happy {
		t.Fatalf("new puppy mood %s", m)
	}
	s.TestFailed("pytest", 0, noon)
	if m := s.MoodAt(noon.Add(time.Minute)); m != Sad {
		t.Fatalf("just failed: %s", m)
	}
	s.TestPassed("pytest", 0, noon.Add(2*time.Minute))
	if m := s.MoodAt(noon.Add(3 * time.Minute)); m != Proud {
		t.Fatalf("just passed: %s", m)
	}
	s.Food = 10
	if m := s.MoodAt(noon.Add(time.Hour)); m != Hungry {
		t.Fatalf("food 10: %s", m)
	}
	late := time.Date(2026, 11, 19, 2, 0, 0, 0, time.Local)
	if m := s.MoodAt(late); m != Asleep {
		t.Fatalf("2am: %s", m)
	}
}

func TestStreakDaysAndWeekBadge(t *testing.T) {
	s := New("Spot", noon)
	for d := range 7 {
		s.TestPassed("go test", 0, noon.AddDate(0, 0, d))
	}
	if s.GreenDays != 7 {
		t.Fatalf("green days %d", s.GreenDays)
	}
	if _, ok := s.Badges["week"]; !ok {
		t.Fatal("no week badge")
	}
	s.TestPassed("go test", 0, noon.AddDate(0, 0, 9))
	if s.GreenDays != 1 {
		t.Fatalf("a gap should reset green days, got %d", s.GreenDays)
	}
}

func TestLastResultIsBounded(t *testing.T) {
	s := New("Spot", noon)
	for i := range 200 {
		s.TestPassed("go test ./pkg"+string(rune('a'+i%26))+string(rune('a'+i/26)), 0, noon)
	}
	if len(s.LastResult) > 64 {
		t.Fatalf("remembered %d commands", len(s.LastResult))
	}
	if len(s.Log) > 30 {
		t.Fatalf("log has %d lines", len(s.Log))
	}
}

func TestPlural(t *testing.T) {
	for _, c := range []struct {
		n    int
		w, o string
	}{{1, "treat", "treat"}, {0, "treat", "treats"}, {2, "catch", "catches"}} {
		if got := plural(c.n, c.w); got != c.o {
			t.Errorf("plural(%d, %q) = %q", c.n, c.w, got)
		}
	}
}
