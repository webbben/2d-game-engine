package stepper

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/webbben/2d-game-engine/config"
	"github.com/webbben/2d-game-engine/logz"
	"github.com/webbben/2d-game-engine/model"
	"github.com/webbben/2d-game-engine/ui/button"
	"github.com/webbben/2d-game-engine/ui/text"
	"golang.org/x/image/font"
)

type Stepper struct {
	decrementButton *button.Button
	incrementButton *button.Button

	counterMin, counterMax int
	counterVal             int
	counterFont            font.Face
	counterFg, counterBg   color.Color
	counterMaxWidth        int
}

type StepperParams struct {
	MinVal, MaxVal       int
	Font                 font.Face
	FontFg               color.Color
	FontBg               color.Color
	DecrementButtonImage *ebiten.Image
	IncrementButtonImage *ebiten.Image
}

func NewStepper(params StepperParams) Stepper {
	if params.Font == nil {
		params.Font = config.DefaultFont
	}
	if params.MinVal >= params.MaxVal {
		logz.Panicln("NewStepper", "invalid min/max values:", "min:", params.MinVal, "max:", params.MaxVal)
	}

	s := Stepper{
		decrementButton: button.NewImageButton("", config.DefaultFont, params.DecrementButtonImage),
		incrementButton: button.NewImageButton("", config.DefaultFont, params.IncrementButtonImage),
		counterFont:     params.Font,
		counterMin:      params.MinVal,
		counterMax:      params.MaxVal,
		counterVal:      params.MinVal,
		counterFg:       params.FontFg,
		counterBg:       params.FontBg,
	}

	maxWidth, _, _ := text.GetStringSize(fmt.Sprintf("%v", s.counterMax), s.counterFont)
	s.counterMaxWidth = maxWidth

	return s
}

func (s *Stepper) Update() {
	if s.decrementButton.Update().Clicked {
		s.Decrement()
	}
	if s.incrementButton.Update().Clicked {
		s.Increment()
	}
}

func (s Stepper) Dimensions() (dx, dy int) {
	tilesize := int(config.GetScaledTilesize())
	dx = s.decrementButton.Width + s.incrementButton.Width + (tilesize / 2) + s.counterMaxWidth
	dy = max(s.decrementButton.Height, s.incrementButton.Height)
	return dx, dy
}

func (s *Stepper) Draw(screen *ebiten.Image, x, y float64) {
	tilesize := int(config.TileSize * config.UIScale)
	drawX := int(x)
	drawY := int(y)

	s.decrementButton.Draw(screen, drawX, drawY)
	drawX += s.decrementButton.Width

	// draw the counter
	drawX += tilesize / 4
	count := fmt.Sprintf("%v", s.counterVal)
	sX, sY := text.CenterTextInRect(count, s.counterFont, model.Rect{X: float64(drawX), Y: float64(drawY), W: float64(s.counterMaxWidth), H: float64(tilesize)})
	text.DrawShadowText(screen, count, s.counterFont, sX, sY, s.counterFg, s.counterBg, 0, 0)
	drawX += (tilesize / 4) + s.counterMaxWidth

	s.incrementButton.Draw(screen, drawX, drawY)
}

func (s *Stepper) Decrement() {
	s.counterVal--
	if s.counterVal < s.counterMin {
		s.counterVal = s.counterMax
	}
}

func (s *Stepper) Increment() {
	s.counterVal++
	if s.counterVal > s.counterMax {
		s.counterVal = s.counterMin
	}
}

func (s *Stepper) SetValue(val int) {
	if val < s.counterMin || val > s.counterMax {
		logz.Panicf("%v is out of bounds [%v, %v]", val, s.counterMin, s.counterMax)
	}
	s.counterVal = val
}

func (s *Stepper) GetValue() int {
	return s.counterVal
}
