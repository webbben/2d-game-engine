package dialog

import (
	"fmt"
	"image/color"
	"log"
	"os"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text"
	"github.com/webbben/2d-game-engine/config"
	"github.com/webbben/2d-game-engine/entity"
	"github.com/webbben/2d-game-engine/general_util"
	"github.com/webbben/2d-game-engine/tileset"
	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
)

var dialogBoxHeight = 255
var dialogX int = 0
var dialogY int = config.ScreenHeight - dialogBoxHeight
var textOffsetX int = 30
var textOffsetY int = 45
var textY int = dialogY + textOffsetY
var textX int = dialogX + textOffsetX
var charsPerSecond int = 40

type Conversation struct {
	Greeting          Dialog
	greetingDone      bool
	Bridge            Dialog
	Topics            map[string]Dialog
	topicKeys         []string
	visitedTopics     []bool // whether topic at index has been visited
	hoveredTopicIndex int
	currentDialog     *Dialog
	Character         entity.Entity
	End               bool

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
	OptionBoxCoords      BoxCoords
	optionBoxInit        bool
	boxInit              bool
}

type BoxCoords struct {
	X, Y, Width, Height int
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
	CurrentText         string       // current text showing in the dialog
	charIndex           int          // index of the char that is next to show in the dialog text
	frameCounter        int          // counts how many frames have passed on the current dialog step
	lastSpacePressFrame int          // for timing space bar presses during dialog
	End                 bool         // if this dialog has ended yet. signals to stop rendering dialog and send control back outside of the dialog
	ShowOptionsWindow   bool         // if the options window should show. options window is for choosing dialog options/topics
	AwaitSpaceKey       bool         // if the dialog should wait for a final space key press before ending
	spaceKeyFlash       bool         // used to flash the space key prompt every second or so
	lastFlash           time.Time    // last time the space key prompt was flashed
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
	if c.Topics != nil && c.optionBoxInit {
		if len(c.Topics) > 0 {
			DrawOptions(screen, c.Boxes.OptionBox, c.Font.fontFace, c.topicKeys, c.OptionBoxCoords.X, c.OptionBoxCoords.Y, c.hoveredTopicIndex, c.visitedTopics)
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

	text.Draw(screen, d.CurrentText, f.fontFace, textX, textY, color.White)
	if d.spaceKeyFlash {
		text.Draw(screen, ">", f.fontFace, dialogX+b.DialogBox.Bounds().Dx()-40, dialogY+b.DialogBox.Bounds().Dy()-25, color.White)
	}
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
	// once dialog is finished, wait for topic selection
	if c.currentDialog.End {
		if len(c.topicKeys) == 0 {
			// init topics, and if none exist end the conversation
			c.topicKeys = make([]string, 0, len(c.Topics))
			for k := range c.Topics {
				c.topicKeys = append(c.topicKeys, k)
			}
			c.visitedTopics = make([]bool, len(c.topicKeys))
			if len(c.topicKeys) == 0 {
				c.End = true
				return
			}
		}
		c.UpdateOptions()
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
		if time.Since(d.lastFlash) >= time.Second {
			d.spaceKeyFlash = !d.spaceKeyFlash
			d.lastFlash = time.Now()
		}
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
					d.spaceKeyFlash = false
				}
			} else {
				d.End = true
				d.spaceKeyFlash = false
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

func DrawOptions(screen, box *ebiten.Image, f font.Face, options []string, x, y int, hoverIndex int, visitedOptions []bool) {
	if len(options) == 0 {
		return
	}

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(x), float64(y))
	screen.DrawImage(box, op)

	for i, option := range options {
		var c color.Color = color.RGBA{200, 200, 150, 0}
		if visitedOptions[i] {
			// set text color to light purple if option has been visited
			c = color.RGBA{180, 50, 180, 0} // light purple color
		}
		if i == hoverIndex {
			c = color.RGBA{255, 255, 255, 255} // white
			text.Draw(screen, "*", f, x+textOffsetX-5, y+(20*i)+textOffsetY+5, c)
		}
		text.Draw(screen, option, f, x+textOffsetX+10, y+(20*i)+textOffsetY, c)
	}
}

func (c *Conversation) UpdateOptions() {
	if !c.boxInit {
		return
	}
	// initialize option box position
	if !c.optionBoxInit {
		boxHeight := 20 * len(c.topicKeys)
		boxY := dialogY
		if boxHeight > dialogBoxHeight {
			boxY = dialogY - (boxHeight - dialogBoxHeight)
		}
		c.OptionBoxCoords = BoxCoords{
			X:      dialogX + c.DialogBox.Bounds().Dx(),
			Y:      boxY,
			Width:  c.OptionBox.Bounds().Dx(),
			Height: c.OptionBox.Bounds().Dy(),
		}
		c.optionBoxInit = true
	}
	// topic selection uses the mouse pointer
	// detect which topic is being hovered over
	c.hoveredTopicIndex = -1
	dy := -10 // need to adjust y position a bit to get the hover detection to work correctly
	if general_util.IsHovering(c.OptionBoxCoords.X, c.OptionBoxCoords.Y, c.OptionBoxCoords.X+c.OptionBoxCoords.Width, c.OptionBoxCoords.Y+c.OptionBoxCoords.Height) {
		for i := range c.topicKeys {
			if general_util.IsHovering(c.OptionBoxCoords.X+textOffsetX, c.OptionBoxCoords.Y+(20*i)+textOffsetY+dy, c.OptionBoxCoords.X+c.OptionBoxCoords.Width, c.OptionBoxCoords.Y+(20*(i+1))+textOffsetY+dy) {
				c.hoveredTopicIndex = i
			}
		}
	}
	// detect if a topic is clicked
	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
		if c.hoveredTopicIndex >= 0 {
			topic, exists := c.Topics[c.topicKeys[c.hoveredTopicIndex]]
			if !exists {
				fmt.Println("**Warning! Topic not found in conversation")
				return
			}
			c.currentDialog = &topic
			c.visitedTopics[c.hoveredTopicIndex] = true
		}
	}
}
