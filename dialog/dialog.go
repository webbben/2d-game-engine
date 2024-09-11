package dialog

import (
	"fmt"
	"image/color"
	"log"
	"os"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/webbben/2d-game-engine/config"
	"github.com/webbben/2d-game-engine/tileset"
	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
)

var dialogBoxHeight = 255
var dialogX int = 10
var dialogY int = config.ScreenHeight - dialogBoxHeight
var textY int = dialogY + 45
var textX int = dialogX + 25
var charsPerSecond int = 22

type DialogStep struct {
	Text         string
	TakeInput    bool
	InputOptions []string
}

type DialogTiles struct {
	Top, TopLeft, TopRight, Left, Right, BottomLeft, Bottom, BottomRight, Fill *ebiten.Image
}

type Dialog struct {
	Type                int // the type of dialog screen
	Steps               []DialogStep
	CurrentStep         int
	SpeakerName         string // name of the person that the user is interacting with
	box                 *ebiten.Image
	boxInit             bool // box image has been created
	fontFace            font.Face
	fontInit            bool   // font has been loaded
	FontName            string // name of the font to use
	CurrentText         string
	charIndex           int
	frameCounter        int
	lastSpacePressFrame int
	End                 bool // if this dialog has ended yet. signals to stop rendering dialog and send control back outside of the dialog
	ShowOptionsWindow   bool // if the options window should show. options window is for choosing dialog options/topics
	DialogTiles
}

func (d *Dialog) SetDialogTiles(imagesDirectoryPath string) {
	tileset, err := tileset.LoadTilesetByPath(imagesDirectoryPath)
	if err != nil {
		fmt.Println("failed to set dialog tiles:", err)
		return
	}
	d.Top = tileset["T"]
	d.TopLeft = tileset["TL"]
	d.TopRight = tileset["TR"]
	d.Left = tileset["L"]
	d.Right = tileset["R"]
	d.BottomLeft = tileset["BL"]
	d.Bottom = tileset["B"]
	d.BottomRight = tileset["BR"]
	d.Fill = tileset["F"]

	// confirm all tiles loaded correctly
	if d.Top == nil || d.TopLeft == nil || d.TopRight == nil || d.Left == nil ||
		d.Right == nil || d.BottomLeft == nil || d.Bottom == nil ||
		d.BottomRight == nil || d.Fill == nil {
		fmt.Println("**Warning! Some dialog tiles are not loaded correctly.")
	}
}

func (d *Dialog) createDialogBox(numTilesWide, numTilesHigh, tileSize int) {
	d.box = ebiten.NewImage(numTilesWide*tileSize, numTilesHigh*tileSize)
	for x := 0; x < numTilesWide; x++ {
		for y := 0; y < numTilesHigh; y++ {
			// get the image we will place
			var img *ebiten.Image
			op := &ebiten.DrawImageOptions{}
			if x == 0 {
				if y == 0 {
					// top left
					img = d.TopLeft
				} else if y == numTilesHigh-1 {
					// bottom left
					img = d.BottomLeft
				} else {
					// left
					img = d.Left
				}
			} else if x == numTilesWide-1 {
				if y == 0 {
					// top right
					img = d.TopRight
				} else if y == numTilesHigh-1 {
					// bottom right
					img = d.BottomRight
				} else {
					// right
					img = d.Right
				}
			} else if y == 0 {
				img = d.Top
			} else if y == numTilesHigh-1 {
				img = d.Bottom
			} else {
				img = d.Fill
				op.ColorScale.ScaleAlpha(0.75)
			}
			// draw the tile
			op.GeoM.Translate(float64(x*tileSize), float64(y*tileSize))
			d.box.DrawImage(img, op)
		}
	}
}

func (d *Dialog) DrawDialog(screen *ebiten.Image) {
	if d.End || d.CurrentStep >= len(d.Steps) {
		return
	}
	if !d.fontInit {
		d.fontFace = loadFont("Planewalker")
		d.fontInit = true
	}
	if !d.boxInit {
		boxWidth := (config.ScreenWidth / 17) - 2
		boxHeight := (config.ScreenHeight / 3) / 17
		d.createDialogBox(boxWidth, boxHeight, 17)
		d.boxInit = true
	}

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(dialogX), float64(dialogY))
	screen.DrawImage(d.box, op)

	// draw the speaker name and dialog text
	speakerText := d.SpeakerName + ": " + d.CurrentText
	text.Draw(screen, speakerText, d.fontFace, textX, textY, color.White)
}

func loadFont(fontName string) font.Face {
	fontFile, err := os.ReadFile("/Users/benwebb/dev/fun/ancient-rome/dialog/fonts/" + fontName + ".ttf")
	if err != nil {
		log.Fatal(err)
	}

	// parse font file
	ttf, err := opentype.Parse(fontFile)
	if err != nil {
		log.Fatal(err)
	}

	// create font face
	const dpi = 72
	customFont, err := opentype.NewFace(ttf, &opentype.FaceOptions{
		Size:    20,
		DPI:     dpi,
		Hinting: font.HintingFull,
	})
	if err != nil {
		log.Fatal(err)
	}
	return customFont
}

func (d *Dialog) DrawInputBox(screen *ebiten.Image) {
	// Set the dimensions for the input box
	boxWidth := 200
	boxHeight := 150
	boxX := config.ScreenWidth - boxWidth - 20 // Positioned on the right side
	boxY := config.ScreenHeight - boxHeight - 20

	// Draw the black box background
	boxColor := color.RGBA{0, 0, 0, 255} // Solid black
	vector.DrawFilledRect(screen, float32(boxX), float32(boxY), float32(boxWidth), float32(boxHeight), boxColor, false)

	// Draw the border (a simple, interesting border design)
	borderColor := color.RGBA{255, 255, 255, 255} // White border for contrast

	// Top and bottom borders
	for i := 0; i < boxWidth; i += 10 {
		vector.DrawFilledRect(screen, float32(boxX+i), float32(boxY), 8, 2, borderColor, false)             // Top border
		vector.DrawFilledRect(screen, float32(boxX+i), float32(boxY+boxHeight-2), 8, 2, borderColor, false) // Bottom border
	}

	// Left and right borders
	for i := 0; i < boxHeight; i += 10 {
		vector.DrawFilledRect(screen, float32(boxX), float32(boxY+i), 2, 8, borderColor, false)            // Left border
		vector.DrawFilledRect(screen, float32(boxX+boxWidth-2), float32(boxY+i), 2, 8, borderColor, false) // Right border
	}
}

func (d *Dialog) UpdateDialog() {
	if d.End {
		return
	}
	// full text of the step we are on
	stepText := d.Steps[d.CurrentStep].Text

	// If all characters are revealed, wait for spacebar to advance to the next step
	if d.charIndex >= len(stepText) && d.frameCounter-d.lastSpacePressFrame > 15 {
		if ebiten.IsKeyPressed(ebiten.KeySpace) {
			if d.CurrentStep < len(d.Steps)-1 {
				d.CurrentStep++
				d.resetTextProgress()
			} else {
				d.End = true
			}
		}
		return
	}

	if ebiten.IsKeyPressed(ebiten.KeySpace) && d.charIndex < len(stepText) {
		d.CurrentText = stepText
		d.charIndex = len(stepText)
		d.lastSpacePressFrame = d.frameCounter
		return
	}

	d.frameCounter++
	// show text one character at a time
	if d.frameCounter >= 60/charsPerSecond && d.charIndex < len(stepText) {
		d.CurrentText += string(stepText[d.charIndex])
		d.charIndex++
		d.frameCounter = 0
	}
}

func (d *Dialog) resetTextProgress() {
	d.CurrentText = ""
	d.charIndex = 0
	d.frameCounter = 0
	d.lastSpacePressFrame = 0
}
