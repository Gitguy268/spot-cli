package react

import (
	"fmt"
	"io"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/charmbracelet/colorprofile"
	"github.com/charmbracelet/x/ansi"

	"github.com/Gitguy268/spot-cli/internal/art"
	"github.com/Gitguy268/spot-cli/internal/pet"
)

// Opts says where and how to draw.
type Opts struct {
	Out     io.Writer
	Profile colorprofile.Profile
	Pal     art.Palette
	Width   int    // terminal columns; 0 means 80
	Light   bool   // light terminal background
	Mode    string // "full", "short" or "off"
	FPS     int    // 0 means 12
	Sleep   func(time.Duration)
}

func (o Opts) fps() int {
	if o.FPS <= 0 {
		return 12
	}
	return o.FPS
}

func (o Opts) width() int {
	if o.Width <= 0 {
		return 80
	}
	return o.Width
}

func (o Opts) sleep(d time.Duration) {
	if o.Sleep != nil {
		o.Sleep(d)
		return
	}
	time.Sleep(d)
}

// miniFor maps a big reaction to the small animation shown in the shell.
var miniFor = map[string]string{
	"wag": "mini-wag", "zoomies": "mini-wag", "love": "mini-love", "eat": "mini-love",
	"sad": "mini-sad", "sleep": "mini-sleep", "sleepy": "mini-sleep",
}

// Lines is the text shown next to Spot for a reaction.
func Lines(r pet.Reaction, st *pet.State, p colorprofile.Profile, light bool) []string {
	c := Colours(light)
	name := paint(p, c.Name, true)
	dim := paint(p, c.Dim, false)
	gold := paint(p, c.Gold, false)
	text := r.Text
	if st != nil && strings.HasPrefix(text, st.Name) {
		text = name(st.Name) + text[len(st.Name):]
	}
	out := []string{text}
	var info []string
	if r.Treats > 0 {
		info = append(info, fmt.Sprintf("+%d 🦴", r.Treats))
	}
	if st != nil {
		if st.Streak > 1 {
			info = append(info, fmt.Sprintf("streak %d", st.Streak))
		}
		if r.Anim == "sad" && st.Treats > 0 {
			info = append(info, fmt.Sprintf("%d 🦴 saved up", st.Treats))
		}
	}
	if len(info) > 0 {
		out = append(out, dim(strings.Join(info, " · ")))
	}
	for _, t := range r.NewTricks {
		out = append(out, gold("✦ new trick: "+t.Name)+dim("  try: spot trick "+t.ID))
	}
	for _, b := range r.NewBadges {
		out = append(out, gold(b.Icon+" badge: "+b.Name))
	}
	return out
}

// Play shows a reaction in the shell. In full mode Spot animates for about a
// second (zoomies run across the screen first) and then stays in the
// scrollback with the text beside him; short mode is one line.
func Play(o Opts, r pet.Reaction, st *pet.State) {
	if o.Mode == "off" {
		return
	}
	lines := Lines(r, st, o.Profile, o.Light)
	mood := pet.Content
	if st != nil {
		mood = st.MoodAt(time.Now())
	}
	if o.Mode == "short" || o.width() < 48 {
		f := Face(moodFor(r, mood), "face", "", o.Profile, o.Light)
		fmt.Fprintln(o.Out, f+"  "+strings.Join(lines, "  "))
		return
	}
	stop := interrupted()
	defer stop()
	fmt.Fprint(o.Out, ansi.HideCursor)
	defer fmt.Fprint(o.Out, ansi.ShowCursor)

	d := drawer{o: o}
	if r.Anim == "zoomies" {
		d.zoomies()
	}
	a := art.Get(miniFor[r.Anim])
	if _, ok := miniFor[r.Anim]; !ok {
		a = art.Get("mini-idle")
	}
	frames := o.fps() * 6 / 5
	for t := 0; t < frames && !d.cancelled(); t++ {
		d.show(Beside(art.Render(a.Frame(t, o.fps()), o.Pal, o.Profile), lines, 2))
		o.sleep(time.Second / time.Duration(o.fps()))
	}
	d.show(Beside(art.Render(a.Frames[0], o.Pal, o.Profile), lines, 2))
}

// PlayBig shows a big animation inline for dur, with a caption under it.
func PlayBig(o Opts, name string, dur time.Duration, caption string) {
	if o.Mode == "off" {
		if caption != "" {
			fmt.Fprintln(o.Out, caption)
		}
		return
	}
	stop := interrupted()
	defer stop()
	fmt.Fprint(o.Out, ansi.HideCursor)
	defer fmt.Fprint(o.Out, ansi.ShowCursor)
	d := drawer{o: o}
	a := art.Get(name)
	n := int(dur.Seconds() * float64(o.fps()))
	if o.Mode == "short" {
		n = 1
	}
	for t := 0; t < max(n, 1) && !d.cancelled(); t++ {
		f := art.Render(a.Frame(t, o.fps()), o.Pal, o.Profile)
		if caption != "" {
			f = append(f, "", caption)
		}
		d.show(f)
		if n > 1 {
			o.sleep(time.Second / time.Duration(o.fps()))
		}
	}
}

func moodFor(r pet.Reaction, fallback pet.Mood) pet.Mood {
	switch r.Anim {
	case "wag", "zoomies", "love", "eat":
		return pet.Proud
	case "sad":
		return pet.Sad
	case "beg":
		return pet.Hungry
	}
	return fallback
}

type drawer struct {
	o    Opts
	rows int
}

var stopped = make(chan os.Signal, 1)

func (d *drawer) cancelled() bool {
	select {
	case <-stopped:
		return true
	default:
		return false
	}
}

// show replaces what was drawn last with lines.
func (d *drawer) show(lines []string) {
	var b strings.Builder
	if d.rows > 0 {
		fmt.Fprintf(&b, "\r\x1b[%dA", d.rows)
	}
	b.WriteString("\r\x1b[J")
	for _, l := range lines {
		b.WriteString(l)
		b.WriteString("\n")
	}
	io.WriteString(d.o.Out, b.String())
	d.rows = len(lines)
}

// zoomies runs Spot across the terminal.
func (d *drawer) zoomies() {
	w := min(d.o.width()-1, 100)
	run := art.Get("zoomies")
	for x, t := -32, 0; x < w && !d.cancelled(); x, t = x+3, t+1 {
		c := art.NewCanvas(w, 20)
		c.Draw(run.Frame(t, 20), x, 0)
		if t%4 < 2 {
			c.Draw(art.Sparkle, x-4, 4+t%3)
		}
		d.show(art.Render(c.Sprite(), d.o.Pal, d.o.Profile))
		d.o.sleep(time.Second / 20)
	}
}

// Beside puts text lines to the right of a picture, centred on it.
func Beside(pic, text []string, gap int) []string {
	w := 0
	for _, l := range pic {
		w = max(w, ansi.StringWidth(l))
	}
	top := max(0, (len(pic)-len(text))/2)
	out := make([]string, max(len(pic), len(text)+top))
	for i := range out {
		l := ""
		if i < len(pic) {
			l = pic[i]
		}
		if j := i - top; j >= 0 && j < len(text) {
			l += strings.Repeat(" ", w-ansi.StringWidth(l)+gap) + text[j]
		}
		out[i] = l
	}
	return out
}

func paint(p colorprofile.Profile, hex string, bold bool) func(string) string {
	if p <= colorprofile.ASCII {
		return func(s string) string { return s }
	}
	st := ansi.Style{}.ForegroundColor(p.Convert(ansi.XParseColor(hex)))
	if bold {
		st = st.Bold()
	}
	return func(s string) string { return st.String() + s + ansi.ResetStyle }
}

// interrupted catches Ctrl-C while Spot is on screen, so the cursor comes
// back; the caller exits normally after.
func interrupted() func() {
	signal.Notify(stopped, os.Interrupt)
	return func() { signal.Stop(stopped) }
}
