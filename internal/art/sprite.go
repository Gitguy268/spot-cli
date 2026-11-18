package art

import (
	"fmt"
	"strings"
)

// Sprite is a grid of palette runes. Rows all have the same width.
type Sprite []string

// W is the width in pixels.
func (s Sprite) W() int {
	if len(s) == 0 {
		return 0
	}
	return len([]rune(s[0]))
}

// H is the height in pixels.
func (s Sprite) H() int { return len(s) }

// At returns the rune at x, y, or '.' outside the grid.
func (s Sprite) At(x, y int) rune {
	if y < 0 || y >= len(s) || x < 0 {
		return '.'
	}
	r := []rune(s[y])
	if x >= len(r) {
		return '.'
	}
	return r[x]
}

// Patch is drawn on top of a sprite at X, Y. In a patch, '.' keeps the pixel
// underneath and '_' clears it.
type Patch struct {
	X, Y int
	Rows []string
}

// With returns a copy of s with the patches drawn over it in order.
func (s Sprite) With(ps ...Patch) Sprite {
	grid := make([][]rune, len(s))
	for i, row := range s {
		grid[i] = []rune(row)
	}
	for _, p := range ps {
		for dy, row := range p.Rows {
			y := p.Y + dy
			if y < 0 || y >= len(grid) {
				continue
			}
			for dx, c := range []rune(row) {
				x := p.X + dx
				if x < 0 || x >= len(grid[y]) || c == '.' {
					continue
				}
				if c == '_' {
					c = '.'
				}
				grid[y][x] = c
			}
		}
	}
	out := make(Sprite, len(grid))
	for i, row := range grid {
		out[i] = string(row)
	}
	return out
}

// Shift moves the whole picture by dx, dy, keeping its size.
func (s Sprite) Shift(dx, dy int) Sprite {
	out := make(Sprite, len(s))
	w := s.W()
	for y := range s {
		var b strings.Builder
		for x := 0; x < w; x++ {
			b.WriteRune(s.At(x-dx, y-dy))
		}
		out[y] = b.String()
	}
	return out
}

// Mirror flips the sprite left to right.
func (s Sprite) Mirror() Sprite {
	out := make(Sprite, len(s))
	for i, row := range s {
		r := []rune(row)
		for a, b := 0, len(r)-1; a < b; a, b = a+1, b-1 {
			r[a], r[b] = r[b], r[a]
		}
		out[i] = string(r)
	}
	return out
}

// Flip turns the sprite upside down (Spot rolled over on his back).
func (s Sprite) Flip() Sprite {
	out := make(Sprite, len(s))
	for i, row := range s.Mirror() {
		out[len(s)-1-i] = row
	}
	return out
}

// Check reports rows of unequal width and runes missing from the palette.
func (s Sprite) Check(p Palette) error {
	w := s.W()
	for y, row := range s {
		r := []rune(row)
		if len(r) != w {
			return fmt.Errorf("row %d is %d wide, want %d", y, len(r), w)
		}
		for x, c := range r {
			if c == '.' {
				continue
			}
			if _, ok := p[c]; !ok {
				return fmt.Errorf("pixel %d,%d: %q is not in the palette", x, y, c)
			}
		}
	}
	return nil
}

// Rows is the number of terminal rows a rendered sprite takes.
func (s Sprite) Rows() int { return (s.H() + 1) / 2 }
