package dialog

import (
	"image/color"
	"log"
	"os"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text"
	"github.com/webbben/2d-game-engine/config"
	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
)

var fontFace font.Face = nil
var dialogBoxHeight = 255
var dialogX int = 10
var dialogY int = config.ScreenHeight - dialogBoxHeight
var textY int = dialogY + 45
var textX int = dialogX + 25

type DialogStep struct {
	Text         string
	TakeInput    bool
	InputOptions []string
}

type Dialog struct {
	Type        int // the type of dialog screen
	Steps       []DialogStep
	CurrentStep int
	SpeakerName string // name of the person that the user is interacting with
	Box         *ebiten.Image
}

func (d *Dialog) DrawDialog(screen *ebiten.Image) {
	if d.CurrentStep >= len(d.Steps) {
		return
	}
	if fontFace == nil {
		fontFace = loadFont("Papyrus")
	}

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(dialogX), float64(dialogY))
	screen.DrawImage(d.Box, op)

	// draw the speaker name and dialog text
	speakerText := d.SpeakerName + ": " + d.Steps[d.CurrentStep].Text
	text.Draw(screen, speakerText, fontFace, textX, textY, color.White)
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
