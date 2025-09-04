package dialog

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/webbben/2d-game-engine/internal/rendering"
	"github.com/webbben/2d-game-engine/internal/text"
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

	d.lineWriter.Draw(screen, int(d.x+20), int(d.y+35))

	if d.flashContinueIcon {
		// TODO
	} else if d.flashDoneIcon {
		continueX := int(d.x) + d.boxImage.Bounds().Dx() - 25
		continueY := int(d.y) + d.boxImage.Bounds().Dy() - 8
		text.DrawShadowText(screen, "", d.TextFont.fontFace, continueX, continueY, nil, nil, 0, 0)
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
	d.lineWriter.Update()
	switch d.lineWriter.WritingStatus {
	case text.LW_AWAIT_PAGER:
		// TODO
	case text.LW_TEXT_DONE:
		// all text has been displayed. If there are no options to show and we are waiting to continue,
		// show a flashing icon on the bottom right
		d.iconFlashTimer++
		if d.iconFlashTimer > 30 {
			d.flashDoneIcon = !d.flashDoneIcon
			d.iconFlashTimer = 0
		}
	}
}
