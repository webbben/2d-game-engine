package dialogv2

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/webbben/2d-game-engine/internal/config"
	"github.com/webbben/2d-game-engine/internal/display"
	"github.com/webbben/2d-game-engine/internal/rendering"
	"github.com/webbben/2d-game-engine/internal/text"
)

// updateDialogResponse handles the logic for what to do next from the current node in the current topic's conversation
func (ds *DialogSession) updateDialogResponse() {
	if ds.Exit {
		return
	}

	if ds.currentResponse == nil {
		panic("current response was nil")
	}
	if ds.LineWriter.IsWriting() {
		panic("called updateDialogResponse while linewriter is still writing")
	}

	switch ds.responseStatus {
	case dialogResponseStarted:
		// since we are no longer writing, that means we should move on to the next status
		ds.responseStatus = dialogResponseTextDone
		return
	case dialogResponseTextDone:
		// response is done: check if there are replies for the user, or a chained response to move into
		if ds.currentResponse.NextResponse != nil && len(ds.currentResponse.Replies) > 0 {
			panic("dialogResponse has conflicting options: next response is set, but there are also replies")
		}
		if ds.currentResponse.NextResponse != nil {
			// move to chained response
			ds.ApplyResponse(*ds.currentResponse.NextResponse)
			return
		}
		if len(ds.currentResponse.Replies) > 0 {
			// show user reply options and wait for them to choose
			ds.responseStatus = dialogResponseUserReply
			return
		}
		// there were no replies and no chained response; go back to topic selection.
		ds.responseStatus = dialogResponseFinished
	case dialogResponseUserReply:
		if len(ds.currentResponse.Replies) == 0 {
			panic("no replies available")
		}
		if len(ds.replyButtons) == 0 {
			ds.setupReplyOptions()
		}
		for i, b := range ds.replyButtons {
			if b.Update().Clicked {
				r := ds.replyList[i]
				ds.ApplyReply(r)
				return
			}
		}
	case dialogResponseFinished:
		// await user input for topic selection
		if len(ds.topicButtons) == 0 {
			ds.setupTopicOptions()
		}
		for i, b := range ds.topicButtons {
			if b.Update().Clicked {
				// set topic
				t := ds.topicList[i]
				ds.SetTopic(t)
				return
			}
		}
	}
}

/*
*
* Overview of DialogSession logic:
*
* 1. Dialog starts
* - first greeting is shown (DialogResponse)
*
* 2. Await Topic Selection
*
* 3. Topic Selected:
* - response chain plays
* - topic's content ends -> back to (2)
*
 */

func (ds *DialogSession) Update() {
	// handle text display
	ds.LineWriter.Update()

	switch ds.LineWriter.WritingStatus {
	case text.Writing:
		ds.flashContinueIcon = false
		// check if user is clicking to skip forward
		ds.skipForward()
		return
	case text.AwaitPager:
		// LineWriter has finished a page, but has more to show.
		// wait for user input before continuing
		ds.awaitContinue()
		return
	case text.TextDone:
		// all text has been displayed
		ds.updateDialogResponse()
	}
}

func (ds *DialogSession) skipForward() {
	if !ds.LineWriter.IsWriting() {
		panic("tried to skip forward, but linewriter is writing anything")
	}
	if ds.handleUserClick() {
		// user has signaled to continue; page lineWriter
		ds.LineWriter.FastForward()
	}
}

func (ds *DialogSession) handleUserClick() bool {
	ds.ticksSinceLastClick++
	if ds.ticksSinceLastClick < 30 {
		return false
	}

	if ebiten.IsKeyPressed(ebiten.KeySpace) || ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
		ds.ticksSinceLastClick = 0
		return true
	}
	return false
}

func (ds *DialogSession) awaitContinue() {
	if ds.LineWriter.IsWriting() {
		panic("awaitContinue called, but linewriter is still writing")
	}

	// flash done icon
	ds.iconFlashTimer++
	if ds.iconFlashTimer > 30 {
		ds.flashContinueIcon = !ds.flashContinueIcon
		ds.iconFlashTimer = 0
	}

	if ds.handleUserClick() {
		// user has signaled to continue; page lineWriter
		ds.LineWriter.NextPage()
	}
}

func (ds *DialogSession) Draw(screen *ebiten.Image) {
	tileSize := config.GetScaledTilesize()

	textBoxBounds := ds.TextBoxImg.Bounds()
	textBoxY := display.SCREEN_HEIGHT - textBoxBounds.Dy()
	textBoxX := 0
	rendering.DrawImage(screen, ds.TextBoxImg, float64(textBoxX), float64(textBoxY), 0)

	ds.LineWriter.Draw(screen, textBoxX+int(tileSize/2), textBoxY+int(tileSize/2))

	// if linewriter is waiting to continue, show flashing continue icon
	if ds.flashContinueIcon {
		text.DrawShadowText(screen, "", ds.f, textBoxX+textBoxBounds.Dx()-int(tileSize), textBoxY+textBoxBounds.Dy()-int(tileSize/2), nil, nil, 0, 0)
	}

	if len(ds.replyButtons) > 0 && len(ds.topicButtons) > 0 {
		panic("can't have reply buttons and topic buttons at the same time")
	}

	optionBoxX := textBoxX + textBoxBounds.Dx()
	optionBoxY := display.SCREEN_HEIGHT // subtract the height of the option buttons from this
	buttonHeight := 0
	if len(ds.replyButtons) > 0 {
		buttonHeight = ds.replyButtons[0].Height
	} else if len(ds.topicButtons) > 0 {
		buttonHeight = ds.topicButtons[0].Height
	}
	numOptions := max(len(ds.replyButtons), len(ds.topicButtons))
	optionBoxY -= buttonHeight * numOptions
	if optionBoxY > textBoxY {
		optionBoxY = textBoxY // don't let it go lower than the text box
	}
	rendering.DrawImage(screen, ds.TopicBoxImg, float64(optionBoxX), float64(optionBoxY), 0)

	if len(ds.replyButtons) > 0 {
		// show reply buttons where topic buttons normally go
		for i, b := range ds.replyButtons {
			b.Draw(screen, optionBoxX+int(tileSize/2), optionBoxY+(i*buttonHeight)+(int(tileSize/2)))
		}
	} else if len(ds.topicButtons) > 0 {
		for i, b := range ds.topicButtons {
			b.Draw(screen, optionBoxX+int(tileSize/2), optionBoxY+(i*buttonHeight)+(int(tileSize/2)))
		}
	}
}
