package slider

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/webbben/2d-game-engine/config"
	"github.com/webbben/2d-game-engine/internal/logz"
	"github.com/webbben/2d-game-engine/model"
	"github.com/webbben/2d-game-engine/ui/text"
	"golang.org/x/image/font"
)

type SliderGroup struct {
	sliders          map[string]*Slider
	sliderNames      []string
	labelFont        font.Face
	labelFg, labelBg color.Color
}

func (sg SliderGroup) Dimensions() (dx, dy int) {
	tilesize := config.TileSize * config.UIScale
	sliderDx, sliderDy := sg.sliders[sg.sliderNames[0]].Dimensions()
	width := int(tilesize*2) + sliderDx
	height := sliderDy * len(sg.sliderNames)
	return width, height
}

type SliderGroupParams struct {
	LabelFont    font.Face
	LabelColorFg color.Color
	LabelColorBg color.Color
}

type SliderDef struct {
	Label  string
	Params SliderParams
}

func NewSliderGroup(params SliderGroupParams, sliderDefs []SliderDef) SliderGroup {
	sg := SliderGroup{
		sliders:     make(map[string]*Slider),
		sliderNames: make([]string, 0),
		labelFont:   params.LabelFont,
		labelFg:     params.LabelColorFg,
		labelBg:     params.LabelColorBg,
	}

	for _, def := range sliderDefs {
		sg.sliderNames = append(sg.sliderNames, def.Label)
		slider := NewSlider(def.Params)
		sg.sliders[def.Label] = &slider
	}

	return sg
}

func (sg *SliderGroup) Update() {
	for _, slider := range sg.sliders {
		slider.Update()
	}
}

func (sg *SliderGroup) Draw(screen *ebiten.Image, x, y float64) {
	tileSize := config.TileSize * config.UIScale
	drawX := x
	drawY := y
	for _, name := range sg.sliderNames {
		slider, exists := sg.sliders[name]
		if !exists {
			panic("slider not found: " + name)
		}
		sx, sy := text.CenterTextInRect(name, sg.labelFont, model.Rect{X: drawX, Y: drawY, W: tileSize, H: tileSize})
		text.DrawShadowText(screen, name, sg.labelFont, sx, sy, sg.labelFg, sg.labelBg, 0, 0)
		slider.Draw(screen, drawX+tileSize, drawY)

		sliderWidth, _ := slider.Dimensions()
		sx, sy = text.CenterTextInRect(name, sg.labelFont, model.Rect{X: drawX + tileSize + float64(sliderWidth), Y: drawY, W: tileSize, H: tileSize})
		text.DrawShadowText(screen, fmt.Sprintf("%v", slider.GetValue()), sg.labelFont, sx, sy, sg.labelFg, sg.labelBg, 0, 0)

		drawY += tileSize
	}
}

func (sg SliderGroup) GetValue(sliderKey string) (val int) {
	slider, found := sg.sliders[sliderKey]
	if !found {
		logz.Panicf("tried to get value of nonexistent slider: %s", sliderKey)
	}

	return slider.GetValue()
}
