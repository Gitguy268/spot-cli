package art

// sit is Spot sitting and looking at you: big head, floppy ears, red collar
// with a gold tag. Every other pose is this grid with patches on top.
var sit = Sprite{
	"............................",
	"..........OOOOOOOO..........",
	"........OOKHHHKKKKOO........",
	"......OOKHHKKKKKKKKKOO......",
	".....OKKHKKKKKKKKKKKKKO.....",
	"...OOOKKKKKKKKKKKKKKKKOOO...",
	"..OkkkOKKKKKKKKKKKKKKOkkkO..",
	".OkkkkOKKKKKKKKKKKKKKOkkkkO.",
	".OkkkkOKKEEKKKKKKEEKKOkkkkO.",
	".OkkkkOKEWEEKKKKEWEEKOkkkkO.",
	".OkkkkOKEEwEKKKKEEwEKOkkkkO.",
	".OkkkkOKKEEKMMMMKEEKKOkkkkO.",
	".OkkkkOBBKKMMMMMMKKBBOkkkkO.",
	".OkkkkOKKKMMNnNNMMKKKOkkkkO.",
	"..OkkkOKKKMMMNNMMMKKKOkkkO..",
	"...OOOOKKKMkMkkMkMKKKOOOO...",
	"......OKKKMMkMMkMMKKKO......",
	"......OOKKKMMMMMMKKKOO......",
	".......ORRRRRRRRRRRRO.......",
	".......OrrrrrGGrrrrrO.......",
	"......OKKKKKKggKKKKKKO......",
	".....OKKKKKKKKKKKKKKKKO.....",
	"....OKKKKkKKKKKKKKkKKKKO....",
	"...OKKKKKkKKKkkKKKkKKKKKO...",
	"...OkKKKKkKKKkkKKKkKKKKkO...",
	"...OkKKKKkKKKkkKKKkKKKKkO...",
	"..OHHHHKkHkHHkkHHkHkKHHHHO..",
	"..OOOOOOOOOOOOOOOOOOOOOOOO..",
}

// Eyes sit at x 8 and 16, rows 8 to 11 of the unpadded grid.
func eyes(rows ...string) []Patch {
	return []Patch{{X: 8, Y: 8, Rows: rows}, {X: 16, Y: 8, Rows: rows}}
}

var (
	eyesOpen   = eyes("KEEK", "EWEE", "EEwE", "KEEK")
	eyesBlink  = eyes("KKKK", "KKKK", "EKKE", "KEEK")
	eyesHappy  = eyes("KKKK", "KEEK", "EKKE", "KKKK")
	eyesLeft   = eyes("KEEK", "WEEE", "wEEE", "KEEK")
	eyesRight  = eyes("KEEK", "EEWE", "EEEw", "KEEK")
	eyesUp     = eyes("EWWE", "EWEE", "EEEE", "KEEK")
	eyesSad    = eyes("KKKK", "EWWE", "EWwE", "KEEK")
	eyesSleepy = eyes("KKKK", "KKKK", "EEEE", "KKKK")
	eyesStars  = eyes("KSSK", "SSSS", "KSSK", "KKKK")
	eyesHeart  = eyes("LKLK", "LLLK", "KLKK", "KKKK")
)

var (
	tongue     = []Patch{{X: 12, Y: 16, Rows: []string{".TT.", ".Tt."}}}
	tongueLong = []Patch{{X: 12, Y: 16, Rows: []string{".TT.", ".Tt.", ".TT."}}}
	mouthSad   = []Patch{{X: 11, Y: 15, Rows: []string{"MMkkMM", "MkMMkM"}}}
	mouthOpen  = []Patch{{X: 11, Y: 15, Rows: []string{"MkEEkM", "MMEtMM"}}}
	tear       = []Patch{{X: 8, Y: 12, Rows: []string{"X", "X"}}}
	blushBig   = []Patch{{X: 7, Y: 12, Rows: []string{"BBB"}}, {X: 18, Y: 12, Rows: []string{"BBB"}}}
)

// Paw trick: the front paw on your left goes up for a high five.
var pawUp = []Patch{
	{X: 9, Y: 26, Rows: []string{"KKKK"}},
	{X: 5, Y: 15, Rows: []string{".OOOO.", "OHHHHO", "OHkHkO", "OKKKKO", ".OKKO.", ".OKKO.", ".OKKO."}},
}

// Bowl drawn in front of Spot, over his paws.
var bowl = []Patch{{X: 7, Y: 21, Rows: []string{
	"....F.f..F.F.f....",
	"..fFFfFFfFFfFFFf..",
	".OOOOOOOOOOOOOOOO.",
	"OYYYYYYYYYYYYYYYYO",
	".OYYYYYYYYYYYYYYO.",
	"..OyyyyyyyyyyyyO..",
	"...OOOOOOOOOOOO...",
}}}

// Tails are drawn after the grid is padded to 32 wide. The body's right
// edge is at x 26 on rows 23 to 25.
var (
	tailUp   = []Patch{{X: 26, Y: 16, Rows: []string{"...OO.", "..OKKO", "..OKO.", ".OKKO.", ".OKO..", "OKKO..", "KKO...", "KO...."}}}
	tailMid  = []Patch{{X: 26, Y: 19, Rows: []string{".....O", "...OOK", "..OKKO", ".OKKO.", "KKKO..", "KOO..."}}}
	tailFlat = []Patch{{X: 26, Y: 23, Rows: []string{"OOOO..", "KKKKOO", "OOOOO."}}}
)

const padX = 2

func pad(s Sprite, n int) Sprite {
	out := make(Sprite, len(s))
	side := ""
	for range n {
		side += "."
	}
	for i, row := range s {
		out[i] = side + row + side
	}
	return out
}

// padTop adds empty rows above a sprite so it lines up with taller ones.
func padTop(s Sprite, n int) Sprite {
	blank := ""
	for range s.W() {
		blank += "."
	}
	out := make(Sprite, 0, len(s)+n)
	for range n {
		out = append(out, blank)
	}
	return append(out, s...)
}

func mirrorPatches(ps []Patch, w int) []Patch {
	out := make([]Patch, len(ps))
	for i, p := range ps {
		rows := Sprite(p.Rows).Mirror()
		out[i] = Patch{X: w - p.X - rows.W(), Y: p.Y, Rows: rows}
	}
	return out
}

// pose draws face patches on the unpadded grid, pads it, then draws the tail.
func pose(tail []Patch, face ...[]Patch) Sprite {
	var all []Patch
	for _, p := range face {
		all = append(all, p...)
	}
	s := pad(sit.With(all...), padX)
	return s.With(tail...)
}

func left(t []Patch) []Patch { return mirrorPatches(t, sit.W()+2*padX) }
