package art

import (
	"strings"
	"testing"

	"github.com/charmbracelet/colorprofile"
	"github.com/charmbracelet/x/ansi"
)

func TestEverySpriteIsWellFormed(t *testing.T) {
	for name, a := range Anims {
		if len(a.Frames) == 0 || a.FPS <= 0 {
			t.Errorf("%s: %d frames at %d fps", name, len(a.Frames), a.FPS)
		}
		w, h := a.Frames[0].W(), a.Frames[0].H()
		for i, f := range a.Frames {
			for _, p := range []Palette{Dark, Light} {
				if err := f.Check(p); err != nil {
					t.Errorf("%s frame %d: %v", name, i, err)
				}
			}
			if f.W() != w || f.H() != h {
				t.Errorf("%s frame %d is %dx%d, frame 0 is %dx%d", name, i, f.W(), f.H(), w, h)
			}
		}
	}
	for _, s := range []Sprite{Heart, HeartTiny, BigZ, SmallZ, Ball, Bone, Sparkle, Drop, Bowl, BowlFull, Blanket} {
		if err := s.Check(Dark); err != nil {
			t.Errorf("prop: %v", err)
		}
	}
}

func TestSizes(t *testing.T) {
	for name, want := range map[string][2]int{"idle": {32, 28}, "wag": {32, 28}, "mini-wag": {22, 14}, "run": {32, 20}, "spin": {32, 28}} {
		f := Anims[name].Frames[0]
		if f.W() != want[0] || f.H() != want[1] {
			t.Errorf("%s is %dx%d, want %dx%d", name, f.W(), f.H(), want[0], want[1])
		}
	}
}

// Spot is recognisable: every big sitting frame has his red collar and gold tag.
func TestCollarAndTag(t *testing.T) {
	for _, name := range []string{"idle", "wag", "sad", "love", "paw", "mini-idle", "mini-wag"} {
		for i, f := range Anims[name].Frames {
			s := strings.Join(f, "")
			if !strings.ContainsRune(s, 'R') || !strings.ContainsRune(s, 'G') {
				t.Errorf("%s frame %d has no red collar or no tag", name, i)
			}
		}
	}
}

func TestRenderWidths(t *testing.T) {
	f := Anims["wag"].Frames[0]
	for _, p := range []colorprofile.Profile{colorprofile.TrueColor, colorprofile.ANSI256, colorprofile.ANSI, colorprofile.ASCII} {
		lines := Render(f, Dark, p)
		if len(lines) != f.Rows() {
			t.Fatalf("%v: %d lines, want %d", p, len(lines), f.Rows())
		}
		for i, l := range lines {
			if w := ansi.StringWidth(l); w > f.W() {
				t.Errorf("%v line %d is %d wide", p, i, w)
			}
		}
	}
}

func TestSixteenColoursKeepFurGrey(t *testing.T) {
	out := strings.Join(Render(Anims["idle"].Frames[0], Dark, colorprofile.ANSI), "")
	if strings.Contains(out, "[34m") || strings.Contains(out, ";44m") {
		t.Fatal("fur came out blue in 16 colours")
	}
}

func TestPlainKeepsTheEyes(t *testing.T) {
	lines := Plain(Anims["idle"].Frames[0])
	if !strings.Contains(strings.Join(lines, "\n"), "▀  ▀") {
		t.Fatalf("eyes should show as holes:\n%s", strings.Join(lines, "\n"))
	}
}

func TestFrameTiming(t *testing.T) {
	a := Anim{FPS: 2, Frames: []Sprite{{"a"}, {"b"}}}
	if a.Frame(0, 12)[0] != "a" || a.Frame(6, 12)[0] != "b" || a.Frame(12, 12)[0] != "a" {
		t.Fatal("2 fps animation on a 12 fps clock changes every 6 ticks")
	}
}

func TestPatchHelpers(t *testing.T) {
	s := Sprite{"ab.", "..c"}
	if got := s.Mirror(); got[0] != ".ba" || got[1] != "c.." {
		t.Fatalf("mirror: %v", got)
	}
	if got := s.Flip(); got[0] != "c.." || got[1] != ".ba" {
		t.Fatalf("flip: %v", got)
	}
	if got := s.With(Patch{X: 2, Y: 0, Rows: []string{"x", "_"}}); got[0] != "abx" || got[1] != "..." {
		t.Fatalf("with: %v", got)
	}
	if got := s.Shift(1, 0); got[0] != ".ab" {
		t.Fatalf("shift: %v", got)
	}
}
