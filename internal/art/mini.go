package art

// mini is the small Spot that shows up in your shell after a test run.
// It is 22 pixels wide and 14 tall: 22 columns by 7 rows.
var mini = pad(Sprite{
	"......OOOOOO......",
	"....OOKKHKKKOO....",
	"..OOKHKKKKKKKKOO..",
	".OkkOKKKKKKKKOkkO.",
	".OkkOKEWKKEWKOkkO.",
	".OkkOKEEKKEEKOkkO.",
	".OkkOBKKMMKKBOkkO.",
	"..OkOKKMNNMKKOkO..",
	"...OOKKMkkMKKOO...",
	".....ORRRRRRO.....",
	".....OrrGGrrO.....",
	"....OKKKggKKKO....",
	"...OKKkKKKKkKKO...",
	"...OOHHOOOOHHOO...",
}, 2)

func miniEyes(l, r [2]string) []Patch {
	return []Patch{{X: 8, Y: 4, Rows: l[:]}, {X: 12, Y: 4, Rows: r[:]}}
}

var (
	miniOpen   = miniEyes([2]string{"EW", "EE"}, [2]string{"EW", "EE"})
	miniBlink  = miniEyes([2]string{"KK", "EE"}, [2]string{"KK", "EE"})
	miniHappy  = []Patch{{X: 7, Y: 4, Rows: []string{"KOK", "OKO"}}, {X: 12, Y: 4, Rows: []string{"KOK", "OKO"}}}
	miniSad    = miniEyes([2]string{"EW", "EE"}, [2]string{"EW", "EE"})
	miniSleepy = miniEyes([2]string{"KK", "EE"}, [2]string{"KK", "EE"})
	miniLove   = miniEyes([2]string{"LL", "KL"}, [2]string{"LL", "LK"})
	miniTongue = []Patch{{X: 10, Y: 8, Rows: []string{"TT", "Tt"}}}
	miniFrown  = []Patch{{X: 9, Y: 8, Rows: []string{"kMMk"}}}
	miniTear   = []Patch{{X: 8, Y: 6, Rows: []string{"X", "X"}}}

	miniTailUp  = []Patch{{X: 17, Y: 8, Rows: []string{"...OO", "..OKO", ".OKO.", "OKO..", "KO..."}}}
	miniTailMid = []Patch{{X: 17, Y: 10, Rows: []string{"..OOO", "OOKKO", "KKOO."}}}
)

func miniPose(tail []Patch, face ...[]Patch) Sprite {
	s := mini
	for _, f := range face {
		s = s.With(f...)
	}
	return s.With(tail...)
}

func miniLeft(t []Patch) []Patch { return mirrorPatches(t, mini.W()) }
