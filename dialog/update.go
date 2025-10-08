package dialog

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/webbben/2d-game-engine/internal/display"
	"github.com/webbben/2d-game-engine/internal/pubsub"
	"github.com/webbben/2d-game-engine/internal/rendering"
	"github.com/webbben/2d-game-engine/internal/text"
)

// since dialog's draw needs to set some positions (which are used also to detect mouse hovering) making this a pointer function
func (d *Dialog) Draw(screen *ebiten.Image) {
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

	// draw main dialog box content
	tileSize := d.BoxDef.TileSize()
	lineWriterLastY := d.lineWriter.Draw(screen, int(d.x)+tileSize, int(d.y)+(tileSize/2))
	// show text branch options if applicable
	if d.currentTopic.status == topic_status_awaitOption {
		nextOptionButtonY := lineWriterLastY + 10
		for i, textBranch := range d.currentTopic.currentTextBranch.Options {
			optionButtonX := int(d.x + 20)
			d.currentTopic.currentTextBranch.Options[i].button.Draw(screen, optionButtonX, nextOptionButtonY)
			nextOptionButtonY += textBranch.button.Height
		}
	}

	continueX := int(d.x) + d.boxImage.Bounds().Dx() - tileSize
	continueY := int(d.y) + d.boxImage.Bounds().Dy() - (tileSize / 2)
	if d.flashContinueIcon {
		text.DrawShadowText(screen, "…", d.TextFont.fontFace, continueX, continueY, nil, nil, 0, 0)
	} else if d.flashDoneIcon {
		text.DrawShadowText(screen, "", d.TextFont.fontFace, continueX, continueY, nil, nil, 0, 0)
	}

	// draw subtopic buttons
	for i, subtopic := range d.currentTopic.SubTopics {
		buttonX := int(d.topicBoxX) + (tileSize / 2)
		buttonY := display.SCREEN_HEIGHT - ((i + 1) * subtopic.button.Height) - 15
		subtopic.button.Draw(screen, buttonX, buttonY)
	}
}

func (d *Dialog) Update(eventBus *pubsub.EventBus) {
	if d.Exit {
		// dialog has ended
		return
	}

	if !d.init {
		// do initialization
		d.initialize(eventBus)

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
		// check if user is clicking to skip forward
		d.skipForward()
	case text.LW_AWAIT_PAGER:
		// LineWriter has finished a page, but has more to show.
		// wait for user input before continuing
		d.awaitContinue()
		return
	case text.LW_TEXT_DONE:
		// all text has been displayed

		// handle status transition
		// if status is changed, MUST return so that status logic branch can be reapplied
		switch d.currentTopic.status {
		case topic_status_showingMainText:
			d.currentTopic.status = topic_status_mainTextDone
			return
		case topic_status_returned:
			// topic has been returned to from a subtopic
			// don't show main text and await the next logical step
			// return text should have been shown by the previous topic's transition
			d.currentTopic.status = topic_status_mainTextDone
			return
		case topic_status_mainTextDone:
			// main text is done; now to see if user should select something

			// check for text branch options
			if len(d.currentTopic.currentTextBranch.Options) > 0 {
				// show options and wait for user to select one
				d.currentTopic.status = topic_status_awaitOption
				return
			}

			// check for subtopics
			if len(d.currentTopic.SubTopics) > 0 {
				d.currentTopic.status = topic_status_awaitSubtopic
				return
			}
		case topic_status_awaitOption:
			// awaiting topic selection
			if !d.currentTopic.currentTextBranch.init {
				panic("topic status says await option, but current text branch hasn't been initialized")
			}
			if len(d.currentTopic.currentTextBranch.Options) == 0 {
				panic("awaiting text branch option choice, but there are no text branch options")
			}
			for i := range d.currentTopic.currentTextBranch.Options {
				result := d.currentTopic.currentTextBranch.Options[i].button.Update()
				if result.Clicked {
					// when text branch option is clicked, switch to that option
					d.setCurrentTextBranch(d.currentTopic.currentTextBranch.Options[i])
					return
				}
			}
			return
		case topic_status_goingBack:
			// final text has finished; time to go back to parent topic for real
			d.returnToParentTopic(eventBus)
			return
		case topic_status_awaitSubtopic:
			if len(d.currentTopic.SubTopics) == 0 {
				panic("waiting for subtopic selection even though no subtopics exist!")
			}
			for i := range d.currentTopic.SubTopics {
				result := d.currentTopic.SubTopics[i].button.Update()
				if result.Clicked {
					if d.currentTopic.SubTopics[i].isEndDialogTopic {
						d.EndDialog(eventBus)
					} else {
						d.setTopic(d.currentTopic.SubTopics[i], false, eventBus)
					}
					return
				}
			}
			return
		}

		// All text has been shown and there are no user selection options waiting
		// the current topic has nowhere to go, so await user confirmation to end this topic and go back
		d.awaitDone(eventBus)
		return
	}
}

func (d *Dialog) awaitDone(eventBus *pubsub.EventBus) {
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

	if d.handleUserClick() {
		// user has signaled to continue; end current topic.
		d.returnToParentTopic(eventBus)
	}
}

func (d *Dialog) awaitContinue() {
	if d.currentTopic.status != topic_status_showingMainText {
		panic("awaiting user continue, but topic status isn't showingMainText")
	}

	// flash done icon
	d.iconFlashTimer++
	if d.iconFlashTimer > 30 {
		d.flashContinueIcon = !d.flashContinueIcon
		d.iconFlashTimer = 0
	}

	if d.handleUserClick() {
		// user has signaled to continue; page lineWriter
		d.lineWriter.NextPage()
	}
}

func (d *Dialog) skipForward() {
	if d.handleUserClick() {
		// user has signaled to continue; page lineWriter
		d.lineWriter.FastForward()
	}
}

func (d *Dialog) handleUserClick() bool {
	d.ticksSinceLastClick++
	if d.ticksSinceLastClick < 30 {
		return false
	}

	if ebiten.IsKeyPressed(ebiten.KeySpace) || ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
		d.ticksSinceLastClick = 0
		return true
	}
	return false
}
