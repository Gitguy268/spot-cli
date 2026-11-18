package art

// Anim is a looping animation.
type Anim struct {
	FPS    int
	Frames []Sprite
	// Moves means the picture runs across the screen while it plays.
	Moves bool
}

// Frame returns the frame to show at tick t when the caller ticks fps times
// a second.
func (a Anim) Frame(tick, fps int) Sprite {
	if len(a.Frames) == 0 {
		return nil
	}
	step := tick * a.FPS / max(fps, 1)
	return a.Frames[step%len(a.Frames)]
}

// Anims holds every animation by name. The big ones are 32×28 pixels
// (32 columns by 14 rows); the run cycle is 32×20 and the mini ones 22×14.
var Anims = map[string]Anim{
	"idle": {FPS: 2, Frames: []Sprite{
		pose(tailFlat, eyesOpen), pose(tailFlat, eyesOpen), pose(tailFlat, eyesOpen),
		pose(tailFlat, eyesBlink),
		pose(tailFlat, eyesOpen), pose(tailFlat, eyesOpen),
		pose(tailFlat, eyesLeft), pose(tailFlat, eyesLeft),
		pose(tailFlat, eyesOpen), pose(tailFlat, eyesBlink), pose(tailFlat, eyesOpen),
		pose(tailFlat, eyesRight), pose(tailFlat, eyesRight),
	}},
	"wag": {FPS: 10, Frames: []Sprite{
		pose(tailUp, eyesHappy, tongue),
		pose(tailMid, eyesHappy, tongue).Shift(0, -1),
		pose(nil, eyesHappy, tongue),
		pose(left(tailMid), eyesHappy, tongue).Shift(0, -1),
		pose(left(tailUp), eyesHappy, tongue),
		pose(left(tailMid), eyesHappy, tongue).Shift(0, -1),
		pose(nil, eyesHappy, tongue),
		pose(tailMid, eyesHappy, tongue).Shift(0, -1),
	}},
	"sad": {FPS: 2, Frames: []Sprite{
		pose(nil, eyesSad, mouthSad), pose(nil, eyesSad, mouthSad),
		pose(nil, eyesSad, mouthSad, tear),
	}},
	"love": {FPS: 4, Frames: []Sprite{
		pose(tailUp, eyesHeart, blushBig, tongueLong),
		pose(left(tailUp), eyesHeart, blushBig, tongueLong),
	}},
	"wow":    {FPS: 4, Frames: []Sprite{pose(tailMid, eyesStars, mouthOpen), pose(left(tailMid), eyesStars, mouthOpen)}},
	"beg":    {FPS: 2, Frames: []Sprite{pose(tailFlat, eyesUp), pose(tailFlat, eyesUp), pose(tailFlat, eyesBlink)}},
	"sleepy": {FPS: 1, Frames: []Sprite{pose(nil, eyesSleepy)}},
	"tilt":   {FPS: 2, Frames: []Sprite{pose(tailFlat, eyesRight).Shift(1, 0), pose(tailFlat, eyesUp).Shift(1, 0)}},
	"sleep":  {FPS: 1, Frames: []Sprite{lieAsleep, lieAsleep.Shift(0, 1)}},
	"lie":    {FPS: 1, Frames: []Sprite{lie}},
	"eat":    {FPS: 4, Frames: []Sprite{lieHappy.Shift(0, 2).With(bowl...), lieHappy.Shift(0, 1).With(bowl...)}},
	"paw":    {FPS: 2, Frames: []Sprite{pose(tailMid, eyesHappy, tongue, pawUp), pose(left(tailMid), eyesHappy, tongue, pawUp)}},
	"roll":   {FPS: 4, Frames: []Sprite{belly, belly.With(legsUp2...)}},
	"back":   {FPS: 4, Frames: []Sprite{back}},
	"run":    {FPS: 8, Frames: []Sprite{run1, run2}},
	"spin": {FPS: 6, Frames: []Sprite{
		pose(tailFlat, eyesHappy, tongue), padTop(run2, 8), back, padTop(run2.Mirror(), 8),
	}},
	"zoomies": {FPS: 12, Moves: true, Frames: []Sprite{run1, run2}},

	"mini-idle": {FPS: 2, Frames: []Sprite{
		miniPose(nil, miniOpen), miniPose(nil, miniOpen), miniPose(nil, miniOpen), miniPose(nil, miniBlink),
	}},
	"mini-wag": {FPS: 10, Frames: []Sprite{
		miniPose(miniTailUp, miniHappy, miniTongue),
		miniPose(miniTailMid, miniHappy, miniTongue).Shift(0, -1),
		miniPose(miniLeft(miniTailUp), miniHappy, miniTongue),
		miniPose(miniLeft(miniTailMid), miniHappy, miniTongue).Shift(0, -1),
	}},
	"mini-sad":   {FPS: 2, Frames: []Sprite{miniPose(nil, miniSad, miniFrown), miniPose(nil, miniSad, miniFrown, miniTear)}},
	"mini-love":  {FPS: 2, Frames: []Sprite{miniPose(miniTailUp, miniLove, miniTongue), miniPose(miniLeft(miniTailUp), miniLove, miniTongue)}},
	"mini-sleep": {FPS: 1, Frames: []Sprite{miniPose(nil, miniSleepy)}},
}

var (
	lieAsleep = lie.With(Patch{X: 10, Y: 15, Rows: []string{"KKKK", "KKKK", "EEEE"}}, Patch{X: 18, Y: 15, Rows: []string{"KKKK", "KKKK", "EEEE"}})
	lieHappy  = lie.With(Patch{X: 10, Y: 15, Rows: []string{"KKKK", "KEEK", "EKKE"}}, Patch{X: 18, Y: 15, Rows: []string{"KKKK", "KEEK", "EKKE"}})
)

// belly is Spot rolled over on his back, paws in the air.
var belly = pose(nil, eyesHappy, tongueLong).Flip().Shift(0, 3).With(legsUp1...)

var (
	legsUp1 = []Patch{
		{X: 6, Y: 0, Rows: []string{".OO.", "OHHO", "OKKO", "OKKO"}},
		{X: 11, Y: 1, Rows: []string{".OO.", "OHHO", "OKKO"}},
		{X: 17, Y: 1, Rows: []string{".OO.", "OHHO", "OKKO"}},
		{X: 22, Y: 0, Rows: []string{".OO.", "OHHO", "OKKO", "OKKO"}},
	}
	legsUp2 = []Patch{
		{X: 6, Y: 0, Rows: []string{"____", ".OO.", "OHHO"}},
		{X: 22, Y: 0, Rows: []string{"____", ".OO.", "OHHO"}},
	}
)

// Get returns the named animation, falling back to idle.
func Get(name string) Anim {
	if a, ok := Anims[name]; ok {
		return a
	}
	return Anims["idle"]
}
