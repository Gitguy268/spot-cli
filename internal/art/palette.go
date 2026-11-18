// Package art holds Spot's pixel art and draws it with half blocks.
//
// Every sprite is a grid of runes, one rune per pixel. Each rune names a
// colour in the palette; '.' is see-through. Two pixel rows share one
// terminal row: the top pixel is the foreground of '▀', the bottom one the
// background.
package art

// RGB is a 24-bit colour.
type RGB struct{ R, G, B uint8 }

// Hex returns the colour as #rrggbb.
func (c RGB) Hex() string {
	const digits = "0123456789abcdef"
	b := []byte{'#', 0, 0, 0, 0, 0, 0}
	for i, v := range []uint8{c.R, c.G, c.B} {
		b[1+2*i] = digits[v>>4]
		b[2+2*i] = digits[v&15]
	}
	return string(b)
}

func hex(s string) RGB {
	v := func(i int) uint8 {
		n := func(c byte) uint8 {
			switch {
			case c >= '0' && c <= '9':
				return c - '0'
			case c >= 'a' && c <= 'f':
				return c - 'a' + 10
			case c >= 'A' && c <= 'F':
				return c - 'A' + 10
			}
			panic("art: bad hex colour " + s)
		}
		return n(s[i])<<4 | n(s[i+1])
	}
	return RGB{v(1), v(3), v(5)}
}

// Palette maps sprite runes to colours.
type Palette map[rune]RGB

// Dark is the palette for dark terminals. Spot is a black labrador, so his
// fur is a cool charcoal with a lighter rim; pure black would vanish.
var Dark = Palette{
	'O': hex("#5d6284"), // rim light around the silhouette
	'K': hex("#2c2e3c"), // fur
	'k': hex("#20212c"), // fur in shadow (ears, belly)
	'H': hex("#474b66"), // sheen on the head and back
	'M': hex("#383a4c"), // muzzle
	'E': hex("#0c0c12"), // eye
	'W': hex("#ffffff"), // eye shine
	'w': hex("#9fb4ff"), // second, smaller eye shine
	'N': hex("#5b3f38"), // nose
	'n': hex("#8a6a61"), // nose shine
	'T': hex("#ff7fae"), // tongue
	't': hex("#e2588d"), // tongue shadow
	'R': hex("#e8383d"), // collar
	'r': hex("#a8242a"), // collar shadow
	'G': hex("#ffc94a"), // tag
	'g': hex("#c9912a"), // tag shadow
	'B': hex("#b75a80"), // blush
	'L': hex("#ff6f9c"), // heart
	'l': hex("#ffb3cb"), // heart shine
	'Z': hex("#cdd6ff"), // sleep z, sparkles
	'S': hex("#ffe27a"), // star
	'Y': hex("#4f8dff"), // bowl
	'y': hex("#2f5fc4"), // bowl shadow
	'F': hex("#c98b4f"), // kibble, bone
	'f': hex("#8e5a2e"), // kibble shadow
	'A': hex("#c4f05a"), // tennis ball
	'a': hex("#8fbf2a"), // tennis ball seam
	'D': hex("#3a3354"), // rug
	'd': hex("#4f4672"), // rug stripe
	'C': hex("#e9eefc"), // blanket
	'c': hex("#9aa6d8"), // blanket check
	'X': hex("#7fb2ff"), // tear
}

// Light is the palette for light terminals: the rim is dropped and the fur
// stays properly black, which reads well on white.
var Light = func() Palette {
	p := Palette{}
	for k, v := range Dark {
		p[k] = v
	}
	p['O'] = hex("#15161d")
	p['K'] = hex("#22232d")
	p['k'] = hex("#15161d")
	p['H'] = hex("#3d4057")
	p['M'] = hex("#30323f")
	p['Z'] = hex("#5a6bb8")
	p['C'] = hex("#cfd7f5")
	p['D'] = hex("#e4ddf3")
	p['d'] = hex("#cdc2e8")
	return p
}()
