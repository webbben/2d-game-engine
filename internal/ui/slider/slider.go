package slider

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/webbben/2d-game-engine/internal/config"
	"github.com/webbben/2d-game-engine/internal/logz"
	"github.com/webbben/2d-game-engine/internal/mouse"
	"github.com/webbben/2d-game-engine/internal/rendering"
	"github.com/webbben/2d-game-engine/internal/tiled"
)

type Slider struct {
	tiles     []*ebiten.Image
	sliderImg *ebiten.Image
	ballImg   *ebiten.Image

	minVal, maxVal int
	stepSize       int
	numSteps       int
	stepDistPx     int

	x, y         int
	ballX        int // ball x (offset from slider x)
	currentValue int

	clickStarted bool

	mouse.MouseBehavior
}

func (s Slider) GetValue() int {
	return s.currentValue
}

type SliderParams struct {
	TilesetSrc    string
	TilesetOrigin int
	TileWidth     int // number of tiles wide this slider should be
	MinVal        int
	MaxVal        int
	StepSize      int
	InitialValue  int
}

func NewSlider(params SliderParams) Slider {
	if params.MinVal > params.MaxVal {
		panic("min value is greater than max value")
	}
	if params.MinVal == params.MaxVal {
		panic("min value and max value are the same")
	}
	if params.TileWidth < 3 {
		panic("tile width is less than 3")
	}
	if params.TilesetSrc == "" {
		panic("tileset src is empty")
	}
	if params.InitialValue < params.MinVal || params.InitialValue > params.MaxVal {
		panic("initial value is out of bounds of min & max")
	}
	if params.StepSize <= 0 {
		panic("step size must be positive and non-zero")
	}

	tileset, err := tiled.LoadTileset(params.TilesetSrc)
	if err != nil {
		logz.Panicf("failed to load tileset for slider: %s", err)
	}

	slider := Slider{
		maxVal:   params.MaxVal,
		minVal:   params.MinVal,
		stepSize: params.StepSize,
	}

	tileSize := int(config.TileSize * config.UIScale)

	// load images
	left, err := tileset.GetTileImage(params.TilesetOrigin)
	if err != nil {
		logz.Panicf("failed to load left tile for slider: %s", err)
	}
	middle, err := tileset.GetTileImage(params.TilesetOrigin + 1)
	if err != nil {
		logz.Panicf("failed to load middle tile for slider: %s", err)
	}
	right, err := tileset.GetTileImage(params.TilesetOrigin + 2)
	if err != nil {
		logz.Panicf("failed to load right tile for slider: %s", err)
	}
	ball, err := tileset.GetTileImage(params.TilesetOrigin + 3)
	if err != nil {
		logz.Panicf("failed to load ball tile for slider: %s", err)
	}
	slider.tiles = append(slider.tiles, left, middle, right, ball)

	// build slider image
	slider.sliderImg = ebiten.NewImage(params.TileWidth*tileSize, tileSize)
	for i := range params.TileWidth {
		switch i {
		case 0:
			rendering.DrawImage(slider.sliderImg, slider.tiles[0], float64(tileSize*i), 0, config.UIScale)
		case params.TileWidth - 1:
			rendering.DrawImage(slider.sliderImg, slider.tiles[2], float64(tileSize*i), 0, config.UIScale)
		default:
			rendering.DrawImage(slider.sliderImg, slider.tiles[1], float64(tileSize*i), 0, config.UIScale)
		}
	}
	slider.ballImg = rendering.ScaleImage(slider.tiles[3], config.UIScale)

	// calculate slider movement distance
	slider.numSteps = (slider.maxVal - slider.minVal) / slider.stepSize
	stepDistancePx := ((params.TileWidth - 1) * tileSize) / slider.numSteps
	slider.stepDistPx = stepDistancePx

	slider.SetValue(params.InitialValue)

	return slider
}

func (s *Slider) Update() {
	tileSize := int(config.TileSize * config.UIScale)
	// ballBounds := s.ballImg.Bounds()
	sliderBounds := s.sliderImg.Bounds()
	s.MouseBehavior.Update(s.x, s.y, sliderBounds.Dx(), sliderBounds.Dy(), false)

	if s.MouseBehavior.LeftClick.ClickStart {
		s.clickStarted = true
	} else if (s.MouseBehavior.LeftClick.ClickHolding || s.MouseBehavior.LeftClickOutside.ClickHolding) && s.clickStarted {
		// follow mouse x, as long as its within slider's bounds
		mouseX, _ := ebiten.CursorPosition()

		newValue := (mouseX - s.x - (tileSize / 2)) / s.stepDistPx
		newValue += s.minVal
		s.SetValue(newValue)
	} else {
		s.clickStarted = false
	}
}

func (s *Slider) SetValue(val int) {
	if val > s.maxVal {
		val = s.maxVal
	}
	if val < s.minVal {
		val = s.minVal
	}

	val -= val % s.stepSize

	s.currentValue = val

	step := (val - s.minVal) / s.stepSize
	step = max(0, step)
	step = min(step, s.numSteps)

	// since the stepSize is an int, sometimes it can be slightly too short and you see a gap on the last position
	// so, we calculate the last step position here
	if step == s.numSteps {
		s.ballX = s.sliderImg.Bounds().Dx() - int(config.TileSize*config.UIScale)
		return
	}

	s.ballX = step * s.stepDistPx
	s.ballX -= s.ballX % s.stepDistPx
}

func (s *Slider) Draw(screen *ebiten.Image, x, y float64) {
	s.x = int(x)
	s.y = int(y)

	rendering.DrawImage(screen, s.sliderImg, x, y, 0)
	rendering.DrawImage(screen, s.ballImg, x+float64(s.ballX), y, 0)
}
