package ui

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/webbben/2d-game-engine/internal/config"
	"github.com/webbben/2d-game-engine/internal/rendering"
	"github.com/webbben/2d-game-engine/internal/text"
	"golang.org/x/image/font"
)

type Button struct {
	init          bool
	ButtonText    string
	Width, Height int
	X, Y          int
	OnClick       func() // callback function for when this button is clicked
	isClicked     bool   // flag indicates that the button has been clicked and is waiting for mouse release
	isHovered     bool   // flag indicates that the button is being hovered over
	fontFace      font.Face

	hoverBoxImg *ebiten.Image
	textImg     *ebiten.Image
}

func NewButton(buttonText string, fontFace font.Face, width, height int, x, y int, onClick func()) Button {
	// set defaults
	if fontFace == nil {
		fontFace = config.DefaultFont
		if fontFace == nil {
			panic("default font not loaded!")
		}
	}
	dx, dy := text.GetStringSize(buttonText, fontFace)
	if width == 0 {
		// base on width of button text
		width = dx + 10
	}
	if height == 0 {
		height = dy + 10
	}

	b := Button{
		init:       true,
		ButtonText: buttonText,
		Width:      width,
		Height:     height,
		X:          x,
		Y:          y,
		OnClick:    onClick,
		fontFace:   fontFace,
	}

	// build images
	b.hoverBoxImg = ebiten.NewImage(b.Width, b.Height)
	b.hoverBoxImg.Fill(color.RGBA{30, 30, 30, 5})
	b.textImg = ebiten.NewImage(dx, dy)
	text.DrawShadowText(b.textImg, b.ButtonText, b.fontFace, 0, dy, nil, nil, 0, 0)

	return b
}

func (b *Button) Update() {
	if !b.init {
		panic("button not created yet!")
	}
	// check if button is hovering
	mouseX, mouseY := ebiten.CursorPosition()
	b.isHovered = mouseX > b.X && mouseX < b.X+b.Width && mouseY > b.Y && mouseY < b.Y+b.Height
	if b.isHovered {
		// handle button clicks
		if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
			// button has been clicked
			b.isClicked = true
			return
		} else {
			if b.isClicked {
				// button click has finished (user let go of mouse click), so trigger callback
				b.OnClick()
				b.isClicked = false
				return
			}
		}
	} else {
		if b.isClicked {
			b.isClicked = false
		}
	}
}

func (b Button) Draw(screen *ebiten.Image) {
	if !b.init {
		panic("button not created yet!")
	}
	if b.hoverBoxImg == nil {
		panic("tried to draw button before hover box image was created")
	}
	if b.fontFace == nil {
		panic("no font face set for button!")
	}
	// for now we are just doing clear buttons that highlight when hovered
	dx, dy := rendering.CenterImageOnImage(b.hoverBoxImg, b.textImg)
	rendering.DrawImage(screen, b.textImg, float64(b.X+dx), float64(b.Y+dy), 0)

	if b.isHovered {
		// show a highlight box
		rendering.DrawImage(screen, b.hoverBoxImg, float64(b.X), float64(b.Y), 0)
	}
}
