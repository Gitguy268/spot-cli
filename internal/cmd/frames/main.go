// Command frames exports Spot's sprites.
//
//	go run ./internal/cmd/frames            writes site/frames.json
//	go run ./internal/cmd/frames -png dir   also writes PNG sheets
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"

	"github.com/Gitguy268/spot-cli/internal/art"
)

func main() {
	out := flag.String("o", "site/frames.json", "where to write the JSON")
	pngDir := flag.String("png", "", "also write preview sheets to this folder")
	scale := flag.Int("scale", 12, "pixels per sprite pixel in the PNG sheets")
	only := flag.String("only", "", "comma-separated animations to put in the PNG sheets")
	flag.Parse()

	if err := writeJSON(*out); err != nil {
		fail(err)
	}
	if *pngDir == "" {
		return
	}
	for name, p := range map[string]struct {
		pal art.Palette
		bg  color.RGBA
	}{
		"dark":  {art.Dark, color.RGBA{13, 13, 18, 255}},
		"black": {art.Dark, color.RGBA{0, 0, 0, 255}},
		"light": {art.Light, color.RGBA{250, 250, 247, 255}},
	} {
		if err := writeSheet(filepath.Join(*pngDir, name+".png"), p.pal, p.bg, *scale, *only); err != nil {
			fail(err)
		}
	}
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, "frames:", err)
	os.Exit(1)
}

type export struct {
	Palettes map[string]map[string]string `json:"palettes"`
	Anims    map[string]exportAnim        `json:"anims"`
}

type exportAnim struct {
	FPS    int          `json:"fps"`
	Moves  bool         `json:"moves,omitempty"`
	Frames []art.Sprite `json:"frames"`
}

func writeJSON(path string) error {
	e := export{Palettes: map[string]map[string]string{}, Anims: map[string]exportAnim{}}
	for name, p := range map[string]art.Palette{"dark": art.Dark, "light": art.Light} {
		m := map[string]string{}
		for r, c := range p {
			m[string(r)] = c.Hex()
		}
		e.Palettes[name] = m
	}
	for name, a := range art.Anims {
		e.Anims[name] = exportAnim{FPS: a.FPS, Moves: a.Moves, Frames: a.Frames}
	}
	b, err := json.MarshalIndent(e, "", " ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, append(b, '\n'), 0o644)
}

// writeSheet draws every frame of every animation, one animation per row.
func writeSheet(path string, pal art.Palette, bg color.RGBA, scale int, only string) error {
	var names []string
	cols, cw, ch := 0, 0, 0
	for name, a := range art.Anims {
		if only != "" && !slices.Contains(strings.Split(only, ","), name) {
			continue
		}
		names = append(names, name)
		cols = max(cols, len(a.Frames))
		for _, f := range a.Frames {
			cw, ch = max(cw, f.W()), max(ch, f.H())
		}
	}
	sort.Strings(names)
	cw, ch = cw+2, ch+2
	img := image.NewRGBA(image.Rect(0, 0, cols*cw*scale, len(names)*ch*scale))
	for i := 0; i < len(img.Pix); i += 4 {
		img.Pix[i], img.Pix[i+1], img.Pix[i+2], img.Pix[i+3] = bg.R, bg.G, bg.B, bg.A
	}
	for row, name := range names {
		for col, f := range art.Anims[name].Frames {
			for y := 0; y < f.H(); y++ {
				for x := 0; x < f.W(); x++ {
					c, ok := pal[f.At(x, y)]
					if !ok {
						continue
					}
					for dy := 0; dy < scale; dy++ {
						for dx := 0; dx < scale; dx++ {
							img.Set((col*cw+x+1)*scale+dx, (row*ch+y+1)*scale+dy, color.RGBA{c.R, c.G, c.B, 255})
						}
					}
				}
			}
		}
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return png.Encode(f, img)
}
