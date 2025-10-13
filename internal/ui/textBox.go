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
	icon             *ebiten.Image
	highlightOnHover bool
	mouse.MouseBehavior
	x, y    int
	options TextBoxOptions
	f       font.Face
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
		icon:             icon,
		options:          *ops,
		f:                f,
	}

	textBox.SetText(s)

	return textBox
}

func (tb *TextBox) SetText(s string) {
	tileSize := int(config.TileSize * config.UIScale)

	height := tileSize * 2
	width, _, _ := text.GetStringSize(s, tb.f)
	if tb.icon != nil {
		width += int(float64(tb.icon.Bounds().Dx())*config.UIScale) / 2
	}
	width += int(2.5 * float64(tileSize))
	width -= width % tileSize

	if tb.options.SetWidthPx > 0 {
		width = tb.options.SetWidthPx
	}

	tb.boxImage = tb.box.BuildBoxImage(width, height)

	if tb.icon != nil {
		rendering.DrawImage(tb.boxImage, tb.icon, float64(tileSize/2), float64(tileSize/2), config.UIScale)
	}

	sx, _, _ := text.GetStringSize(s, tb.f)
	sy, _ := text.GetRealisticFontMetrics(tb.f)
	textX := (width / 2) - (sx / 2)
	textY := (height / 2) + (sy / 2)

	if tb.icon != nil {
		textX += (tb.icon.Bounds().Dx() * int(config.UIScale)) / 2
	}

	text.DrawShadowText(tb.boxImage, s, tb.f, textX, textY, nil, nil, 0, 0)
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
