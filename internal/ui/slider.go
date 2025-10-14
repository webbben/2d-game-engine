package ui

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

	minVal, maxVal float64
	stepSize       float64
	stepDistPx     int

	x, y  int
	ballX int // ball x (offset from slider x)

	mouse.MouseBehavior
}

type SliderParams struct {
	TilesetSrc    string
	TilesetOrigin int
	TileWidth     int
	MinVal        float64
	MaxVal        float64
	StepSize      float64
	InitialValue  float64
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
	slider.ballImg = slider.tiles[3]

	// calculate slider movement distance
	numSteps := int((slider.maxVal - slider.minVal) / slider.stepSize)
	stepDistancePx := (params.TileWidth * tileSize) / numSteps
	slider.stepDistPx = stepDistancePx

	return slider
}

func (s *Slider) Update() {
	tileSize := int(config.TileSize * config.UIScale)
	bounds := s.ballImg.Bounds()
	s.MouseBehavior.Update(s.x+s.ballX, s.y, bounds.Dx(), bounds.Dy(), false)
	if s.MouseBehavior.LeftClick.ClickHolding {
		// follow mouse x, as long as its within slider's bounds
		mouseX, _ := ebiten.CursorPosition()
		if mouseX < s.x {
			s.ballX = 0
		} else if mouseX > s.x+s.sliderImg.Bounds().Dx() {
			s.ballX = s.sliderImg.Bounds().Dx() - tileSize
		} else {
			// mouse is somewhere inside the slider; calculate correct step position
			step := (mouseX - s.x) / s.stepDistPx
			s.ballX = step * s.stepDistPx
		}
	}
}

func (s *Slider) Draw(screen *ebiten.Image, x, y float64) {
	s.x = int(x)
	s.y = int(y)

	rendering.DrawImage(screen, s.sliderImg, x, y, 0)
	rendering.DrawImage(screen, s.ballImg, x+float64(s.ballX), y, 0)
}
