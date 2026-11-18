package art

// Small props that get placed into scenes.
var (
	Heart     = Sprite{".L.L.", "LlLLL", ".LLL.", "..L.."}
	HeartTiny = Sprite{"L.L", ".L."}
	BigZ      = Sprite{"ZZZZ", "..Z.", ".Z..", "ZZZZ"}
	SmallZ    = Sprite{"ZZZ", ".Z.", "ZZZ"}
	Ball      = Sprite{".AA.", "AaAA", "AAaA", ".AA."}
	Bone      = Sprite{"FF...FF", ".FFFFF.", "FF...FF"}
	Sparkle   = Sprite{".S.", "SSS", ".S."}
	Drop      = Sprite{".X.", "XXX", ".X."}
	Bowl      = Sprite{
		"..............",
		"..............",
		".OOOOOOOOOOOO.",
		"OYYYYYYYYYYYYO",
		".OYYYYYYYYYYO.",
		"..OyyyyyyyyO..",
		"...OOOOOOOO...",
	}
	BowlFull = Sprite{
		"...F.f..F.f...",
		".fFFfFFfFFfFf.",
		".OOOOOOOOOOOO.",
		"OYYYYYYYYYYYYO",
		".OYYYYYYYYYYO.",
		"..OyyyyyyyyO..",
		"...OOOOOOOO...",
	}
	Blanket = Sprite{
		"..OOOOOOOOOOOOOOOOOOOOOOOOOO..",
		".OCcCcCcCcCcCcCcCcCcCcCcCcCcO.",
		"OCcCcCcCcCcCcCcCcCcCcCcCcCcCcO",
		"OcCcCcCcCcCcCcCcCcCcCcCcCcCcCO",
		".OOOOOOOOOOOOOOOOOOOOOOOOOOOO.",
	}
)

// Canvas is a sprite you draw into.
type Canvas struct {
	w, h int
	px   [][]rune
}

// NewCanvas returns an empty canvas.
func NewCanvas(w, h int) *Canvas {
	c := &Canvas{w: w, h: h, px: make([][]rune, h)}
	for y := range c.px {
		c.px[y] = make([]rune, w)
		for x := range c.px[y] {
			c.px[y][x] = '.'
		}
	}
	return c
}

// Draw puts s at x, y; see-through pixels are skipped.
func (c *Canvas) Draw(s Sprite, x, y int) {
	for dy := 0; dy < s.H(); dy++ {
		for dx := 0; dx < s.W(); dx++ {
			r := s.At(dx, dy)
			px, py := x+dx, y+dy
			if r == '.' || px < 0 || py < 0 || px >= c.w || py >= c.h {
				continue
			}
			c.px[py][px] = r
		}
	}
}

// Ellipse fills an ellipse with r.
func (c *Canvas) Ellipse(cx, cy, rx, ry float64, r rune) {
	for y := 0; y < c.h; y++ {
		for x := 0; x < c.w; x++ {
			dx, dy := (float64(x)+.5-cx)/rx, (float64(y)+.5-cy)/ry
			if dx*dx+dy*dy <= 1 {
				c.px[y][x] = r
			}
		}
	}
}

// Sprite returns what has been drawn.
func (c *Canvas) Sprite() Sprite {
	out := make(Sprite, c.h)
	for y, row := range c.px {
		out[y] = string(row)
	}
	return out
}
