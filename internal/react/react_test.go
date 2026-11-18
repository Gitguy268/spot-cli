package react

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/colorprofile"
	"github.com/charmbracelet/x/ansi"

	"github.com/Gitguy268/spot-cli/internal/art"
	"github.com/Gitguy268/spot-cli/internal/pet"
)

func TestFaceEveryMood(t *testing.T) {
	for _, m := range []pet.Mood{pet.Asleep, pet.Hungry, pet.Sad, pet.Proud, pet.Happy, pet.Content, pet.Sleepy, "unknown"} {
		for _, style := range []string{"face", "emoji", "ascii"} {
			f := Face(m, style, "", colorprofile.TrueColor)
			if f == "" || strings.Contains(f, "\n") {
				t.Errorf("Face(%s, %s) = %q", m, style, f)
			}
		}
		if f := Face(m, "ascii", "", colorprofile.TrueColor); strings.ContainsAny(f, "\x1b") || !isASCII(f) {
			t.Errorf("ascii face %q is not plain ASCII", f)
		}
	}
}

func isASCII(s string) bool {
	for _, r := range s {
		if r > 127 {
			return false
		}
	}
	return true
}

func TestFaceWrapsEscapesForTheShell(t *testing.T) {
	z := Face(pet.Happy, "face", "zsh", colorprofile.TrueColor)
	if !strings.Contains(z, "%{\x1b[") || strings.Count(z, "%{") != strings.Count(z, "%}") {
		t.Fatalf("zsh face doesn't wrap escapes: %q", z)
	}
	b := Face(pet.Happy, "face", "bash", colorprofile.TrueColor)
	if !strings.Contains(b, "\x01\x1b[") || strings.Count(b, "\x01") != strings.Count(b, "\x02") {
		t.Fatalf("bash face doesn't wrap escapes: %q", b)
	}
	if p := Face(pet.Happy, "face", "", colorprofile.NoTTY); strings.Contains(p, "\x1b") {
		t.Fatalf("NoTTY face has colour: %q", p)
	}
}

func TestBeside(t *testing.T) {
	out := Beside([]string{"ab", "a", "abc"}, []string{"x"}, 2)
	if len(out) != 3 || out[1] != "a    x" {
		t.Fatalf("%q", out)
	}
	out = Beside([]string{"a"}, []string{"x", "y", "z"}, 1)
	if len(out) != 3 || out[0] != "a x" || out[2] != "  z" {
		t.Fatalf("%q", out)
	}
}

func newState() *pet.State { return pet.New("Spot", time.Now()) }

func TestPlayShortIsOneLine(t *testing.T) {
	var b bytes.Buffer
	st := newState()
	r := st.TestPassed("go test", 0, time.Now())
	Play(Opts{Out: &b, Profile: colorprofile.ASCII, Mode: "short"}, r, st)
	if n := strings.Count(b.String(), "\n"); n != 1 {
		t.Fatalf("short mode printed %d lines: %q", n, b.String())
	}
	if !strings.Contains(b.String(), "+1 🦴") {
		t.Fatalf("no treat shown: %q", b.String())
	}
}

func TestPlayFullLeavesSpotInScrollback(t *testing.T) {
	var b bytes.Buffer
	st := newState()
	st.TestFailed("go test", 0, time.Now())
	r := st.TestPassed("go test", 0, time.Now())
	if r.Anim != "zoomies" {
		t.Fatalf("want a comeback, got %s", r.Anim)
	}
	frames := 0
	Play(Opts{Out: &b, Profile: colorprofile.TrueColor, Pal: art.Dark, Width: 100, Mode: "full", Sleep: func(time.Duration) { frames++ }}, r, st)
	out := b.String()
	if frames < 20 {
		t.Fatalf("only %d frames for zoomies + wag", frames)
	}
	if !strings.HasPrefix(out, ansi.HideCursor) || !strings.HasSuffix(out, ansi.ShowCursor) {
		t.Fatal("cursor not hidden and restored")
	}
	last := out[strings.LastIndex(out, "\x1b[J")+3:]
	if !strings.Contains(last, "Comeback pup") || strings.Count(last, "\n") != 7 {
		t.Fatalf("final frame should be 7 rows with the badge:\n%s", last)
	}
}

func TestPlayOffPrintsNothing(t *testing.T) {
	var b bytes.Buffer
	st := newState()
	Play(Opts{Out: &b, Mode: "off"}, st.Pet(time.Now()), st)
	if b.Len() != 0 {
		t.Fatalf("off mode printed %q", b.String())
	}
}

func TestNarrowTerminalFallsBackToOneLine(t *testing.T) {
	var b bytes.Buffer
	st := newState()
	Play(Opts{Out: &b, Mode: "full", Width: 30, Profile: colorprofile.ANSI}, st.Pet(time.Now()), st)
	if strings.Count(b.String(), "\n") != 1 {
		t.Fatalf("narrow: %q", b.String())
	}
}
