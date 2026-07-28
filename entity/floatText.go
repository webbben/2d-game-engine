package entity

import (
	"image/color"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/webbben/2d-game-engine/config"
	"github.com/webbben/2d-game-engine/ui/text"
	"github.com/webbben/2d-game-engine/utils"
	"golang.org/x/image/font"
)

type FloatText struct {
	f     font.Face
	c     color.Color
	txt   string
	dur   time.Duration
	start time.Time
	y     float64 // how high (above the entity) the float text has floated
}

type FloatTextParams struct {
	Font     font.Face
	Color    color.Color
	Duration time.Duration
}

func NewFloatText(txt string, params FloatTextParams) *FloatText {
	return &FloatText{
		f:     params.Font,
		c:     params.Color,
		txt:   txt,
		dur:   params.Duration,
		start: time.Now(),
	}
}

func (ft FloatText) Done() bool {
	return time.Since(ft.start) >= ft.dur
}

func (ft *FloatText) Draw(screen *ebiten.Image, entDrawX, entDrawY float64) {
	if ft.Done() {
		return
	}

	params := text.DrawTextParams{
		Font: ft.f,
		Fg:   ft.c,
	}

	// fade the text out as it's duration is expiring
	elapsed := time.Since(ft.start)
	alpha := 1 - float32(elapsed)/float32(ft.dur)

	ops := &ebiten.DrawImageOptions{}
	ops.ColorScale.ScaleAlpha(alpha)

	xOffset := config.TileSize
	yOffset := config.TileSize * 2

	x := int(entDrawX) + xOffset
	y := int(entDrawY-ft.y) - yOffset

	text.DrawTextWithOptions(screen, ft.txt, x, y, params, ops)
}

func (ft *FloatText) Update() {
	if ft.Done() {
		return
	}

	ft.y += 1
}

type FloatMGMT struct {
	FloatTexts []*FloatText
}

func NewFloatMGMT() *FloatMGMT {
	return &FloatMGMT{}
}

func (mgmt *FloatMGMT) Update() {
	for _, ft := range mgmt.FloatTexts {
		ft.Update()
	}

	// now, check for float texts that are done and remove them
	i := 0
	for i < len(mgmt.FloatTexts) {
		if mgmt.FloatTexts[i].Done() {
			mgmt.FloatTexts = utils.RemoveIndexUnordered(mgmt.FloatTexts, i)
		} else {
			i++
		}
	}
}

func (mgmt *FloatMGMT) AddFloatText(ft *FloatText) {
	if ft == nil {
		panic("float text was nil")
	}
	mgmt.FloatTexts = append(mgmt.FloatTexts, ft)
}

func (mgmt *FloatMGMT) Draw(screen *ebiten.Image, entDrawX, entDrawY float64) {
	for _, ft := range mgmt.FloatTexts {
		if !ft.Done() {
			ft.Draw(screen, entDrawX, entDrawY)
		}
	}
}
