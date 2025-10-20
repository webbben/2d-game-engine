package button

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/webbben/2d-game-engine/internal/config"
	"github.com/webbben/2d-game-engine/internal/logz"
	"github.com/webbben/2d-game-engine/internal/mouse"
	"github.com/webbben/2d-game-engine/internal/rendering"
	"github.com/webbben/2d-game-engine/internal/text"
	"golang.org/x/image/font"
)

type Button struct {
	init          bool
	ButtonText    string
	Width, Height int
	x, y          int // position of this button. this is set during draw, and only needed here for checking mouse hovers/clicks
	fontFace      font.Face

	mouseBehavior mouse.MouseBehavior

	buttonImg   *ebiten.Image
	hoverBoxImg *ebiten.Image
	textImg     *ebiten.Image
}

func NewImageButton(buttonText string, fontFace font.Face, img *ebiten.Image) *Button {
	dx := img.Bounds().Dx()
	dy := img.Bounds().Dy()
	b := NewButton(buttonText, fontFace, dx, dy)
	b.buttonImg = img

	return b
}

func NewButton(buttonText string, fontFace font.Face, width, height int) *Button {
	// set defaults
	if fontFace == nil {
		fontFace = config.DefaultFont
		if fontFace == nil {
			panic("default font not loaded!")
		}
	}
	dx, _, _ := text.GetStringSize(buttonText, fontFace)
	dy, dsc := text.GetRealisticFontMetrics(fontFace)
	dsc += 3
	paddingX := 10
	paddingY := dsc * 2
	dx += paddingX
	dy += paddingY
	baselineY := dy - (paddingY / 2)
	baselineX := paddingX / 2

	// if width or height is 0, that means we should set it to be as small as possible (caller doesn't care about size)
	// if an actual size is specified but it's too small for the text & font, change it and log a warning
	if width == 0 {
		// base on width of button text
		width = dx
	} else if width < dx {
		logz.Println("Button "+buttonText, "given width was too small for button text. resizing...")
		width = dx
	}
	if height == 0 {
		height = dy
	} else if height < dy {
		logz.Println("Button "+buttonText, "given height was too small for button text. resizing...")
		height = dy
	}

	b := Button{
		init:       true,
		ButtonText: buttonText,
		Width:      width,
		Height:     height,
		fontFace:   fontFace,
	}

	// build images
	b.hoverBoxImg = ebiten.NewImage(b.Width, b.Height)
	b.hoverBoxImg.Fill(color.RGBA{30, 30, 30, 5})
	b.textImg = ebiten.NewImage(dx, dy)
	//b.textImg.Fill(color.RGBA{100, 0, 0, 50})
	text.DrawShadowText(b.textImg, b.ButtonText, b.fontFace, baselineX, baselineY, nil, nil, 0, 0)

	return &b
}

type ButtonUpdateResult struct {
	Clicked bool
}

func (b *Button) Update() ButtonUpdateResult {
	if !b.init {
		panic("button not created yet: " + b.ButtonText)
	}
	if b.Width == 0 || b.Height == 0 {
		panic("button dimensions are 0!")
	}

	result := ButtonUpdateResult{}

	b.mouseBehavior.Update(b.x, b.y, b.Width, b.Height, false)

	if b.mouseBehavior.LeftClick.ClickReleased {
		result.Clicked = true
	}

	return result
}

func (b *Button) Draw(screen *ebiten.Image, x, y int) {
	if !b.init {
		panic("tried to draw button before it was created!")
	}
	if b.hoverBoxImg == nil {
		panic("tried to draw button before hover box image was created")
	}
	if b.fontFace == nil {
		panic("no font face set for button!")
	}

	// update internal position
	b.x, b.y = x, y

	if b.buttonImg != nil {
		// draw button image instead of highlight hover box
		dx, dy := rendering.CenterImageOnImage(b.buttonImg, b.textImg)
		rendering.DrawImage(screen, b.textImg, float64(x+dx), float64(y+dy), 0)
		ops := ebiten.DrawImageOptions{}
		if b.mouseBehavior.IsHovering {
			ops.ColorScale.Scale(1.2, 1.2, 1.2, 1)
		}
		rendering.DrawImageWithOps(screen, b.buttonImg, float64(x), float64(y), 0, &ops)
	} else {
		// for clear buttons with no button image; we just show a highlight box
		dx, dy := rendering.CenterImageOnImage(b.hoverBoxImg, b.textImg)
		rendering.DrawImage(screen, b.textImg, float64(x+dx), float64(y+dy), 0)
		if b.mouseBehavior.IsHovering {
			rendering.DrawImage(screen, b.hoverBoxImg, float64(x), float64(y), 0)
		}
	}
}
