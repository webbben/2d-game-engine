package dialog

import (
	"fmt"
	"image/color"
	"log"
	"os"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text"
	"github.com/webbben/2d-game-engine/config"
	"github.com/webbben/2d-game-engine/entity"
	"github.com/webbben/2d-game-engine/tileset"
	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
)

var dialogBoxHeight = 255
var dialogX int = 0
var dialogY int = config.ScreenHeight - dialogBoxHeight
var textY int = dialogY + 45
var textX int = dialogX + 25
var charsPerSecond int = 22

type Conversation struct {
	Greeting      Dialog
	greetingDone  bool
	Bridge        Dialog
	Topics        map[string]Dialog
	topicKeys     []string
	topicIndex    int
	currentDialog *Dialog
	Character     entity.Entity
	End           bool

	Boxes
	Font
	DialogTiles // tiles used to make the dialog boxes
}

type Font struct {
	fontFace font.Face // font used in the dialog
	fontInit bool      // font has been loaded
	FontName string    // name of the font to use
}

type Boxes struct {
	DialogBox, OptionBox *ebiten.Image
	boxInit              bool
}

type DialogStep struct {
	Text         string
	TakeInput    bool
	InputOptions []string
}

type DialogTiles struct {
	Top, TopLeft, TopRight, Left, Right, BottomLeft, Bottom, BottomRight, Fill *ebiten.Image
}

type Dialog struct {
	Type                int          // the type of dialog screen
	Steps               []DialogStep // the different steps in the dialog
	CurrentStep         int          // current dialog step
	SpeakerName         string       // name of the person that the user is interacting with
	CurrentText         string       // current text showing in the dialog
	charIndex           int          // index of the char that is next to show in the dialog text
	frameCounter        int          // counts how many frames have passed on the current dialog step
	lastSpacePressFrame int          // for timing space bar presses during dialog
	End                 bool         // if this dialog has ended yet. signals to stop rendering dialog and send control back outside of the dialog
	ShowOptionsWindow   bool         // if the options window should show. options window is for choosing dialog options/topics
	AwaitSpaceKey       bool         // if the dialog should wait for a final space key press before ending
}

func (c *Conversation) SetDialogTiles(imagesDirectoryPath string) {
	tileset, err := tileset.LoadTilesetByPath(imagesDirectoryPath)
	if err != nil {
		fmt.Println("failed to set dialog tiles:", err)
		return
	}
	c.Top = tileset["T"]
	c.TopLeft = tileset["TL"]
	c.TopRight = tileset["TR"]
	c.Left = tileset["L"]
	c.Right = tileset["R"]
	c.BottomLeft = tileset["BL"]
	c.Bottom = tileset["B"]
	c.BottomRight = tileset["BR"]
	c.Fill = tileset["F"]

	// confirm all tiles loaded correctly
	if c.Top == nil || c.TopLeft == nil || c.TopRight == nil || c.Left == nil ||
		c.Right == nil || c.BottomLeft == nil || c.Bottom == nil ||
		c.BottomRight == nil || c.Fill == nil {
		fmt.Println("**Warning! Some dialog tiles are not loaded correctly.")
	}
}

func createDialogBox(numTilesWide, numTilesHigh, tileSize int, t DialogTiles) *ebiten.Image {
	box := ebiten.NewImage(numTilesWide*tileSize, numTilesHigh*tileSize)
	for x := 0; x < numTilesWide; x++ {
		for y := 0; y < numTilesHigh; y++ {
			// get the image we will place
			var img *ebiten.Image
			op := &ebiten.DrawImageOptions{}
			if x == 0 {
				if y == 0 {
					// top left
					img = t.TopLeft
				} else if y == numTilesHigh-1 {
					// bottom left
					img = t.BottomLeft
				} else {
					// left
					img = t.Left
				}
			} else if x == numTilesWide-1 {
				if y == 0 {
					// top right
					img = t.TopRight
				} else if y == numTilesHigh-1 {
					// bottom right
					img = t.BottomRight
				} else {
					// right
					img = t.Right
				}
			} else if y == 0 {
				img = t.Top
			} else if y == numTilesHigh-1 {
				img = t.Bottom
			} else {
				img = t.Fill
				op.ColorScale.ScaleAlpha(0.75)
			}
			// draw the tile
			op.GeoM.Translate(float64(x*tileSize), float64(y*tileSize))
			box.DrawImage(img, op)
		}
	}
	return box
}

func (c *Conversation) DrawConversation(screen *ebiten.Image) {
	if c.End || !c.fontInit || !c.boxInit {
		return
	}
	if c.currentDialog != nil {
		c.currentDialog.DrawDialog(screen, c.Font, c.Boxes, c.DialogTiles)
		// wait until dialog is finished to show options
		if !c.currentDialog.End {
			return
		}
	}
	// show topics
	if c.Topics != nil {
		if len(c.Topics) > 0 {
			DrawOptions(screen, c.Boxes.OptionBox, c.Font.fontFace, c.topicKeys, dialogX+c.Boxes.DialogBox.Bounds().Dx())
		}
	}
}

func (d *Dialog) DrawDialog(screen *ebiten.Image, f Font, b Boxes, tiles DialogTiles) {
	if !f.fontInit {
		fmt.Println("**Warning! Font not loaded for dialog")
		return
	}
	if !b.boxInit {
		fmt.Println("**Warning! Dialog boxes not created")
		return
	}

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(dialogX), float64(dialogY))
	screen.DrawImage(b.DialogBox, op)
	/*
		if d.ShowOptionsWindow {
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(float64(b.DialogBox.Bounds().Dx()), float64(dialogY))
			screen.DrawImage(b.OptionBox, op)
		}*/

	// draw the speaker name and dialog text
	speakerText := d.SpeakerName + ": " + d.CurrentText
	text.Draw(screen, speakerText, f.fontFace, textX, textY, color.White)
}

func loadFont(fontName string) font.Face {
	fontFile, err := os.ReadFile("dialog/fonts/" + fontName + ".ttf")
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

func (c *Conversation) UpdateConversation() {
	if c.End {
		return
	}
	if !c.fontInit {
		fmt.Println("Loading font")
		c.fontFace = loadFont("Planewalker")
		c.fontInit = true
	}
	if !c.boxInit {
		fmt.Println("Creating dialog box")
		boxWidth := (config.ScreenWidth / 17) / 4 * 3
		boxHeight := (config.ScreenHeight / 17) / 3
		c.DialogBox = createDialogBox(boxWidth, boxHeight, 17, c.DialogTiles)
		optionBoxWidth := (config.ScreenWidth / 17) / 4
		optionBoxHeight := (config.ScreenHeight / 17) / 3
		c.OptionBox = createDialogBox(optionBoxWidth, optionBoxHeight, 17, c.DialogTiles)
		c.boxInit = true
	}

	if !c.greetingDone {
		c.currentDialog = &c.Greeting
		c.greetingDone = true
	}
	if c.currentDialog == nil {
		return
	}
	// if there's a dialog, update it
	c.currentDialog.UpdateDialog()
	// once dialog is finished, wait for topic selection, or end if none are available
	if c.currentDialog.End {
		if len(c.topicKeys) == 0 {
			c.topicKeys = make([]string, 0, len(c.Topics))
			for k := range c.Topics {
				c.topicKeys = append(c.topicKeys, k)
			}
		}
		if len(c.topicKeys) == 0 {
			c.End = true
			return
		}
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
		if d.CurrentStep < len(d.Steps)-1 {
			// if there are more steps, wait for space key to advance
			if ebiten.IsKeyPressed(ebiten.KeySpace) {
				d.CurrentStep++
				d.resetTextProgress()
			}
		} else {
			// if this is the last step, wait for a space key if applicable, or end dialog right away
			if d.AwaitSpaceKey {
				if ebiten.IsKeyPressed(ebiten.KeySpace) {
					d.End = true
				}
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

func DrawOptions(screen, box *ebiten.Image, f font.Face, options []string, x int) {
	if len(options) == 0 {
		return
	}
	// calculate height of box based on number of options
	boxHeight := 20 * len(options)
	// calculate box y position based on screen height and box height
	// box is either the same y position as the dialog, or higher up if there are too many options
	boxY := dialogY
	if boxHeight > dialogBoxHeight {
		boxY = dialogY - (boxHeight - dialogBoxHeight)
	}

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(x), float64(boxY))
	screen.DrawImage(box, op)

	for i, option := range options {
		text.Draw(screen, option, f, x+25, boxY+(20*i)+45, color.White)
	}
}
