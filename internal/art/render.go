package art

import (
	"image/color"
	"strings"

	"github.com/charmbracelet/colorprofile"
	"github.com/charmbracelet/x/ansi"
)

// Render draws a sprite as terminal lines, two pixel rows per line.
// Colours are lowered to what the profile supports. With ASCII or NoTTY it
// falls back to Plain.
func Render(s Sprite, pal Palette, p colorprofile.Profile) []string {
	if p <= colorprofile.ASCII {
		return Plain(s)
	}
	lines := make([]string, 0, s.Rows())
	for y := 0; y < s.H(); y += 2 {
		var b strings.Builder
		var last ansi.Style
		set := func(st ansi.Style) {
			if !equal(st, last) {
				b.WriteString(ansi.ResetStyle)
				if len(st) > 0 {
					b.WriteString(st.String())
				}
				last = st
			}
		}
		for x := 0; x < s.W(); x++ {
			top, topOK := colour(p, pal, s.At(x, y))
			bot, botOK := colour(p, pal, s.At(x, y+1))
			switch {
			case !topOK && !botOK:
				set(nil)
				b.WriteByte(' ')
			case topOK && !botOK:
				set(ansi.Style{}.ForegroundColor(top))
				b.WriteString("▀")
			case !topOK && botOK:
				set(ansi.Style{}.ForegroundColor(bot))
				b.WriteString("▄")
			case top == bot:
				set(ansi.Style{}.ForegroundColor(top))
				b.WriteString("█")
			default:
				set(ansi.Style{}.ForegroundColor(top).BackgroundColor(bot))
				b.WriteString("▀")
			}
		}
		set(nil)
		lines = append(lines, strings.TrimRight(b.String(), " "))
	}
	return lines
}

// Plain draws the sprite without colour: fur is ink, and the eyes, nose and
// tongue are left as holes so the face still reads.
func Plain(s Sprite) []string {
	ink := func(c rune) bool { return c != '.' && !strings.ContainsRune("EWwNnTtXGgLlSZ", c) }
	lines := make([]string, 0, s.Rows())
	for y := 0; y < s.H(); y += 2 {
		var b strings.Builder
		for x := 0; x < s.W(); x++ {
			t, u := ink(s.At(x, y)), ink(s.At(x, y+1))
			switch {
			case t && u:
				b.WriteString("█")
			case t:
				b.WriteString("▀")
			case u:
				b.WriteString("▄")
			default:
				b.WriteByte(' ')
			}
		}
		lines = append(lines, strings.TrimRight(b.String(), " "))
	}
	return lines
}

// colour returns the terminal colour for a sprite rune. Its values are
// comparable, so equal pixels can share a full block.
func colour(p colorprofile.Profile, pal Palette, r rune) (ansi.Color, bool) {
	c, ok := pal[r]
	if !ok {
		return nil, false
	}
	if p == colorprofile.ANSI {
		if b, ok := basic[r]; ok {
			return b, true
		}
	}
	return p.Convert(color.RGBA{c.R, c.G, c.B, 255}), true
}

// basic picks 16-colour stand-ins by hand: nearest-colour matching turns
// charcoal fur blue.
var basic = map[rune]ansi.BasicColor{
	'O': ansi.White, 'K': ansi.BrightBlack, 'k': ansi.Black, 'H': ansi.BrightBlack, 'M': ansi.BrightBlack,
	'E': ansi.Black, 'W': ansi.BrightWhite, 'w': ansi.BrightBlue, 'N': ansi.Yellow, 'n': ansi.Yellow,
	'T': ansi.BrightMagenta, 't': ansi.Magenta, 'R': ansi.BrightRed, 'r': ansi.Red, 'G': ansi.BrightYellow,
	'g': ansi.Yellow, 'B': ansi.Magenta, 'L': ansi.BrightMagenta, 'l': ansi.BrightWhite, 'Z': ansi.BrightCyan,
	'S': ansi.BrightYellow, 'Y': ansi.BrightBlue, 'y': ansi.Blue, 'F': ansi.Yellow, 'f': ansi.Red,
	'A': ansi.BrightGreen, 'a': ansi.Green, 'D': ansi.Magenta, 'd': ansi.BrightMagenta, 'C': ansi.BrightWhite,
	'c': ansi.BrightBlue, 'X': ansi.BrightCyan,
}

func equal(a, b ansi.Style) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
