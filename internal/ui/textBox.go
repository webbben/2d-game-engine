package ui

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/webbben/2d-game-engine/internal/config"
	"github.com/webbben/2d-game-engine/internal/rendering"
	"github.com/webbben/2d-game-engine/internal/text"
	"golang.org/x/image/font"
)

type TextBox struct {
	box      BoxDef
	boxImage *ebiten.Image
}

func NewTextBox(s string, tilesetSrc string, originIndex int, f font.Face, icon *ebiten.Image, setWidthPx int) TextBox {
	if s == "" {
		panic("text is empty")
	}

	textBox := TextBox{
		box: NewBox(tilesetSrc, originIndex),
	}

	tileSize := int(config.TileSize * config.UIScale)

	height := tileSize * 2
	width, _, _ := text.GetStringSize(s, f)
	if icon != nil {
		width += int(float64(icon.Bounds().Dx())*config.UIScale) / 2
	}
	width += int(2.5 * float64(tileSize))
	width -= width % tileSize

	if setWidthPx > 0 {
		width = setWidthPx
	}

	textBox.boxImage = textBox.box.BuildBoxImage(width, height)

	if icon != nil {
		rendering.DrawImage(textBox.boxImage, icon, float64(tileSize/2), float64(tileSize/2), config.UIScale)
	}

	sx, _, _ := text.GetStringSize(s, f)
	sy, _ := text.GetRealisticFontMetrics(f)
	textX := (width / 2) - (sx / 2)
	textY := (height / 2) + (sy / 2)

	if icon != nil {
		textX += (icon.Bounds().Dx() * int(config.UIScale)) / 2
	}

	text.DrawShadowText(textBox.boxImage, s, f, textX, textY, nil, nil, 0, 0)

	return textBox
}

func (tb *TextBox) GetImage() *ebiten.Image {
	if tb.boxImage == nil {
		panic("image not generated yet")
	}
	return tb.boxImage
}

func (tb *TextBox) Draw(screen *ebiten.Image, x, y float64) {
	rendering.DrawImage(screen, tb.boxImage, x, y, 0)
}
