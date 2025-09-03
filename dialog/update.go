package dialog

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text"
	"github.com/webbben/2d-game-engine/internal/rendering"
)

func (d Dialog) Draw(screen *ebiten.Image) {
	if !d.init {
		return
	}
	if d.boxImage == nil {
		panic("tried to draw dialog before its box image was built")
	}
	rendering.DrawImage(screen, d.boxImage, d.x, d.y, 0)
	if d.TopicsEnabled {
		rendering.DrawImage(screen, d.topicBoxImage, d.topicBoxX, d.topicBoxY, 0)
	}

	for i, line := range d.lineWriter.writtenLines {
		text.Draw(screen, line, d.TextFont.fontFace, int(d.x+20), int(d.y+35)+(i*d.lineWriter.lineHeight), color.Black)
	}

	if d.lineWriter.showContinueSymbol {
		continueX := int(d.x) + d.boxImage.Bounds().Dx() - 25
		continueY := int(d.y) + d.boxImage.Bounds().Dy() - 8
		text.Draw(screen, "", d.TextFont.fontFace, continueX, continueY, color.Black)
	}
}

func (d *Dialog) Update() {
	if !d.init {
		// do initialization
		d.initialize()
		return
	}

	if d.currentTopic == nil {
		panic("dialog is running in update loop, but has no current topic!")
	}

	// handle text display
	if d.lineWriter.currentLineNumber < len(d.lineWriter.linesToWrite) {
		if d.lineWriter.currentLineIndex < len(d.lineWriter.linesToWrite[d.lineWriter.currentLineNumber]) {
			d.textUpdateTimer++
			if d.textUpdateTimer > 0 {
				// continue to write the current line
				nextChar := d.lineWriter.linesToWrite[d.lineWriter.currentLineNumber][d.lineWriter.currentLineIndex]
				d.lineWriter.writtenLines[d.lineWriter.currentLineNumber] += string(nextChar)
				d.lineWriter.currentLineIndex++
				d.textUpdateTimer = 0
			}
		} else {
			// go to the next line
			d.lineWriter.currentLineNumber++
			d.lineWriter.currentLineIndex = 0
			d.lineWriter.writtenLines = append(d.lineWriter.writtenLines, "") // start next output line
		}
	} else {
		// all text has been displayed. If there are no options to show and we are waiting to continue,
		// show a flashing icon on the bottom right
		d.textUpdateTimer++
		if d.textUpdateTimer > 30 {
			d.lineWriter.showContinueSymbol = !d.lineWriter.showContinueSymbol
			d.textUpdateTimer = 0
		}
	}
}
