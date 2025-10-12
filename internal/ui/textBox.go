package ui

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/webbben/2d-game-engine/internal/config"
	"github.com/webbben/2d-game-engine/internal/mouse"
	"github.com/webbben/2d-game-engine/internal/rendering"
	"github.com/webbben/2d-game-engine/internal/text"
	"golang.org/x/image/font"
)

type TextBox struct {
	box              BoxDef
	boxImage         *ebiten.Image
	highlightOnHover bool
	mouse.MouseBehavior
	x, y int
}

func (tb TextBox) Dimensions() (dx, dy int) {
	bounds := tb.boxImage.Bounds()
	return bounds.Dx(), bounds.Dy()
}

type TextBoxOptions struct {
	HighlightOnHover bool // if set, box will highlight when mouse is hovering over it
	SetWidthPx       int  // if set, this specific width will be used instead of auto-calculating based on content

}

func NewTextBox(s string, tilesetSrc string, originIndex int, f font.Face, icon *ebiten.Image, ops *TextBoxOptions) TextBox {
	if s == "" {
		panic("text is empty")
	}
	if ops == nil {
		ops = &TextBoxOptions{}
	}

	textBox := TextBox{
		box:              NewBox(tilesetSrc, originIndex),
		highlightOnHover: ops.HighlightOnHover,
	}

	tileSize := int(config.TileSize * config.UIScale)

	height := tileSize * 2
	width, _, _ := text.GetStringSize(s, f)
	if icon != nil {
		width += int(float64(icon.Bounds().Dx())*config.UIScale) / 2
	}
	width += int(2.5 * float64(tileSize))
	width -= width % tileSize

	if ops.SetWidthPx > 0 {
		width = ops.SetWidthPx
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
	tb.x = int(x)
	tb.y = int(y)
	ops := ebiten.DrawImageOptions{}
	if tb.highlightOnHover && tb.MouseBehavior.IsHovering {
		ops.ColorScale.Scale(1.1, 1.1, 1.1, 1)
	}
	rendering.DrawImageWithOps(screen, tb.boxImage, x, y, 0, &ops)
}

func (tb *TextBox) Update() {
	if tb.highlightOnHover {
		w, h := tb.Dimensions()
		tb.MouseBehavior.Update(tb.x, tb.y, w, h, false)
	}
}
