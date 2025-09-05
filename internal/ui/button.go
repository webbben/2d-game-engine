package ui

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/webbben/2d-game-engine/internal/config"
	"github.com/webbben/2d-game-engine/internal/logz"
	"github.com/webbben/2d-game-engine/internal/rendering"
	"github.com/webbben/2d-game-engine/internal/text"
	"golang.org/x/image/font"
)

type Button struct {
	init          bool
	ButtonText    string
	Width, Height int
	x, y          int    // position of this button. this is set during draw, and only needed here for checking mouse hovers/clicks
	OnClick       func() // callback function for when this button is clicked
	isClicked     bool   // flag indicates that the button has been clicked and is waiting for mouse release
	isHovered     bool   // flag indicates that the button is being hovered over
	fontFace      font.Face

	hoverBoxImg *ebiten.Image
	textImg     *ebiten.Image
}

func NewButton(buttonText string, fontFace font.Face, width, height int, onClick func()) *Button {
	// set defaults
	if fontFace == nil {
		fontFace = config.DefaultFont
		if fontFace == nil {
			panic("default font not loaded!")
		}
	}
	dx, dy, baselineY := text.GetStringSize(buttonText, fontFace)
	padding := 6
	dx += padding
	dy += padding
	baselineY += padding / 2
	baselineX := padding / 2

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
		OnClick:    onClick,
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

func (b *Button) Update() {
	if !b.init {
		panic("button not created yet: " + b.ButtonText)
	}
	if b.Width == 0 || b.Height == 0 {
		panic("button dimensions are 0!")
	}

	mouseX, mouseY := ebiten.CursorPosition()
	b.isHovered = mouseX > b.x && mouseX < b.x+b.Width && mouseY > b.y && mouseY < b.y+b.Height
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

	// for now we are just doing clear buttons that highlight when hovered
	dx, dy := rendering.CenterImageOnImage(b.hoverBoxImg, b.textImg)
	rendering.DrawImage(screen, b.textImg, float64(x+dx), float64(y+dy), 0)

	if b.isHovered {
		// show a highlight box
		rendering.DrawImage(screen, b.hoverBoxImg, float64(x), float64(y), 0)
	}
}
