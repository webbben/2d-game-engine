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
	if d.Exit {
		// dialog has ended
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

	// draw subtopic buttons
	for _, subtopic := range d.currentTopic.SubTopics {
		subtopic.button.Draw(screen)
	}
}

func (d *Dialog) Update() {
	if d.Exit {
		// dialog has ended
		return
	}

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
	case text.LW_WRITING:
		d.flashDoneIcon = false
		d.flashContinueIcon = false
	case text.LW_AWAIT_PAGER:
		// TODO
	case text.LW_TEXT_DONE:
		// all text has been displayed

		// handle status transition
		switch d.currentTopic.status {
		case topic_status_showingMainText:
			d.currentTopic.status = topic_status_mainTextDone
		case topic_status_returned:
			// topic has been returned to from a subtopic
			// don't show main text and await the next logical step
			// return text should have been shown by the previous topic's transition
			d.currentTopic.status = topic_status_mainTextDone
		case topic_status_goingBack:
			// final text has finished; time to go back to parent topic for real
			d.returnToParentTopic()
			return
		case topic_status_awaitSubtopic:
			if len(d.currentTopic.SubTopics) == 0 {
				panic("waiting for subtopic selection even though no subtopics exist!")
			}
			for _, subtopic := range d.currentTopic.SubTopics {
				subtopic.button.Update()
				// check if current topic changed - if so, return since we need to restart update loop
				if d.currentTopic.status != topic_status_awaitSubtopic {
					return
				}
			}
		}

		if len(d.currentTopic.SubTopics) > 0 {
			d.currentTopic.status = topic_status_awaitSubtopic
			return
		} else {
			// there are no sub-topics, so wait for user to continue and go back to parent topic
			d.awaitDone()
			return
		}
	}
}

func (d *Dialog) awaitDone() {
	if d.currentTopic.status != topic_status_mainTextDone && d.currentTopic.status != topic_status_goingBack {
		// we shouldn't be waiting for a user continue unless we are in one of these topic statuses
		panic("invalid status for dialog.awaitDone: " + d.currentTopic.status)
	}
	if d.lineWriter.WritingStatus != text.LW_TEXT_DONE {
		panic("dialog.awaitDone: lineWriter status is expected to be done. invalid status found: " + d.lineWriter.WritingStatus)
	}

	// flash done icon
	d.iconFlashTimer++
	if d.iconFlashTimer > 30 {
		d.flashDoneIcon = !d.flashDoneIcon
		d.iconFlashTimer = 0
	}

	if ebiten.IsKeyPressed(ebiten.KeySpace) || ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
		// user has signaled to continue; end current topic.
		d.returnToParentTopic()
	}
}
