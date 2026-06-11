package dialogv2

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/webbben/2d-game-engine/config"
	"github.com/webbben/2d-game-engine/display"
	"github.com/webbben/2d-game-engine/imgutil/rendering"
	"github.com/webbben/2d-game-engine/logz"
	"github.com/webbben/2d-game-engine/model"
	"github.com/webbben/2d-game-engine/ui/button"
	"github.com/webbben/2d-game-engine/ui/text"
	"github.com/webbben/2d-game-engine/utils"
)

// updateDialogResponse handles the logic for what to do next from the current node in the current topic's conversation
func (ds *DialogSession) updateDialogResponse() {
	if ds.Ctx.Exit {
		ds.Exit = true
	}
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
	case dialogResponseActionInProg:
		// action is ongoing; allow it to update, but don't let the rest of the response progress
		switch ds.currentResponse.Action.Type {
		case ActionTypeShowScreen:
			if ds.screenViewer == nil {
				panic("screen viewer was nil during show screen action")
			}
			ds.screenViewer.Update()
			if ds.screenViewer.IsDone() {
				// we can disconnect the screen now
				ds.screenViewer = nil
				ds.refreshOpinion()
				ds.continueApplyResponse()
			}
			return
		default:
			logz.Panicln("updateDialogResponse", "action type not recognized:", ds.currentResponse.Action.Type)
		}
	case dialogResponseStarted:
		// since we are no longer writing, that means we should move on to the next status
		// 2026-06-09 Note: not sure exactly why we would end up here, but we do anyhow. Doesn't cause issues so leaving it like this.
		ds.responseStatus = dialogResponseTextDone
		return
	case dialogResponseTextDone:
		// response is done
		// if there are any topic links, create the buttons for them now
		if len(ds.topicLinks) > 0 && ds.topicLinks[0].linkButton == nil {
			linkPositions := ds.LineWriter.GetLinkPositions()
			if len(linkPositions) != len(ds.topicLinks) {
				logz.Panicln("Dialog", "number of link positions doesn't match number of topic links. link pos:", linkPositions, "topic links:", ds.topicLinks)
			}
			for i := range ds.topicLinks {
				pos := linkPositions[i]
				dx := int(pos.X2 - pos.X)
				dy := pos.LineHeight
				ds.topicLinks[i].linkButton = button.NewButton(
					"",
					config.DefaultFont,
					dx,
					dy,
					ds.audioman)
				ds.topicLinks[i].x = pos.X
				ds.topicLinks[i].y = pos.Y - float64(pos.LineHeight)
			}
		}
		// check if there are replies for the user, or a chained response to move into
		if ds.currentResponse.NextResponse != nil && len(ds.currentResponse.Replies) > 0 {
			panic("dialogResponse has conflicting options: next response is set, but there are also replies")
		}
		if ds.currentResponse.NextResponse != nil {
			// move to chained response
			// wait for the player to click to continue though, in case they are still reading.
			if ds.flashUntilContinue() {
				if config.ButtonClickSfx != "" {
					ds.audioman.PlaySFX(config.ButtonClickSfx, 0.3)
				}
				ds.ApplyResponse(*ds.currentResponse.NextResponse)
			}
			return
		}
		if len(ds.currentResponse.NextResponseOptions) > 0 {
			// find the correct next response to move on to
			if ds.flashUntilContinue() {
				for _, nr := range ds.currentResponse.NextResponseOptions {
					if ConditionsMet(nr.Conditions, ds.Ctx) {
						ds.ApplyResponse(nr)
						return
					}
				}
				logz.Panicln(string(ds.Ctx.Profile.ProfileID), "tried to get a next response option, but none had their conditions met; current response text:", ds.currentResponse.Text)
			}
			return
		}
		if ds.currentResponse.Goodbye || len(ds.currentResponse.Replies) > 0 {
			// show user reply options and wait for them to choose
			// goodbye is also handled by a generated reply option
			ds.responseStatus = dialogResponseUserReply
			return
		}
		// there were no replies and no chained response; go back to topic selection.
		ds.responseStatus = dialogResponseFinished
	case dialogResponseUserReply:
		if !ds.currentResponse.Goodbye && len(ds.currentResponse.Replies) == 0 {
			panic("no replies available")
		}
		if len(ds.replyButtons) == 0 && ds.replyBox == nil {
			ds.setupReplyOptions()
			if len(ds.replyButtons) == 0 && ds.replyBox == nil {
				ds.panicln("no reply buttons or replyBox has been created")
			}
		}

		if ds.replyBox != nil {
			ds.updateReplyBox()
		} else {
			for i, b := range ds.replyButtons {
				if b.Update().Clicked {
					r := ds.replyList[i]
					ds.ApplyReply(r)
					return
				}
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
		// also check if topic links are clicked
		for i, link := range ds.topicLinks {
			if ds.topicLinks[i].linkButton.Update().Clicked {
				ds.SetTopic(link.topicID)
				return
			}
		}
	}
}

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
	case text.AwaitText:
		// if we find a case for the lineWriter to sit waiting for text (and not showing previous text) then we can change this
		logz.Panicln("DialogSession", "LineWriter is at awaitText... this probably shouldn't be happening")
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

// flashUntilContinue flashes the continue icon until the player clicks, then returns true so you can optionally do something.
func (ds *DialogSession) flashUntilContinue() bool {
	if ds.LineWriter.IsWriting() {
		panic("flashUntilContinue called, but linewriter is still writing")
	}

	// flash done icon
	ds.iconFlashTimer++
	if ds.iconFlashTimer > 30 {
		ds.flashContinueIcon = !ds.flashContinueIcon
		ds.iconFlashTimer = 0
	}

	return ds.handleUserClick()
}

func (ds *DialogSession) awaitContinue() {
	if ds.flashUntilContinue() {
		if config.ButtonClickSfx != "" {
			ds.audioman.PlaySFX(config.ButtonClickSfx, 0.3)
		}
		ds.LineWriter.NextPage()
	}
}

func (ds *DialogSession) Draw(screen *ebiten.Image) {
	if ds.screenViewer != nil {
		ds.screenViewer.Draw(screen)
		return
	}

	tileSize := config.GetScaledTilesize()

	textBoxBounds := ds.TextBoxImg.Bounds()
	topicBoxWidth := ds.TopicBoxImg.Bounds().Dx()

	// center everything horizontally
	totalWidth := textBoxBounds.Dx() + topicBoxWidth
	startX, _ := utils.CenterInScreen(totalWidth, int(tileSize))

	textBoxY := display.SCREEN_HEIGHT - textBoxBounds.Dy()
	textBoxX := int(startX)
	rendering.DrawImage(screen, ds.TextBoxImg, float64(textBoxX), float64(textBoxY), 0)
	ds.nameTitle.Draw(screen, float64(textBoxX), float64(textBoxY)-tileSize)

	lwX := textBoxX + int(tileSize/2)
	lwY := textBoxY + int(tileSize*2/3) // moved a little further down, since the box title hands down a bit
	ds.LineWriter.Draw(screen, lwX, lwY)
	for i, link := range ds.topicLinks {
		if link.linkButton == nil {
			// buttons haven't been made yet
			break
		}
		ds.topicLinks[i].linkButton.Draw(screen, lwX+int(link.x), lwY+int(link.y))
	}

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

	// position the topic box
	if len(ds.replyButtons) > 0 {
		buttonHeight = ds.replyButtons[0].Height
	} else if len(ds.topicButtons) > 0 {
		buttonHeight = ds.topicButtons[0].Height
	}
	optionBoxY -= ds.TopicBoxImg.Bounds().Dy()
	if optionBoxY > textBoxY {
		optionBoxY = textBoxY // don't let it go lower than the text box
	}
	rendering.DrawImage(screen, ds.TopicBoxImg, float64(optionBoxX), float64(optionBoxY), 0)

	if ds.showCharInfo {
		infoBoxY := 0
		infoBoxX := optionBoxX
		rendering.DrawImage(screen, ds.charInfoBoxImg, float64(infoBoxX), float64(infoBoxY), 0)
		infoBoxY += int(tileSize / 2)
		titleDx, titleDy, _ := text.GetStringSize(ds.npcName, config.DefaultTitleFont)
		infoBoxY += titleDy
		infoBoxX += (topicBoxWidth / 2) - (titleDx / 2)
		text.DrawShadowText(screen, ds.npcName, config.DefaultTitleFont, infoBoxX, infoBoxY, nil, nil, 0, 0)
		infoBoxY += int(tileSize)
		text.DrawShadowText(screen, ds.Ctx.culture.DisplayName, config.DefaultFont, infoBoxX, infoBoxY, nil, nil, 0, 0)
		infoBoxY += int(tileSize)
		opinionString := fmt.Sprintf("%v", ds.Ctx.opinion)
		opinionStringDx, opinionStringDy, _ := text.GetStringSize(opinionString, config.DefaultInfoFont)
		text.DrawShadowText(screen, "Opinion:", config.DefaultFont, infoBoxX, infoBoxY, nil, nil, 0, 0)
		if ds.opinionHoverRect == nil {
			ds.opinionHoverRect = &model.Rect{
				X: float64(infoBoxX),
				Y: float64(infoBoxY - opinionStringDy),
				W: (tileSize * 2) + float64(opinionStringDx),
				H: float64(opinionStringDy),
			}
		}
		infoBoxX += int(tileSize * 2)
		c := color.RGBA{255, 255, 0, 0}
		if ds.Ctx.opinion > 0 {
			c = color.RGBA{0, 255, 0, 0}
		} else if ds.Ctx.opinion < 0 {
			c = color.RGBA{255, 0, 0, 0}
		}
		text.DrawText(screen, opinionString, config.DefaultInfoFont, infoBoxX, infoBoxY, c)

		if ds.opinionHoverRect != nil {
			mouseX, mouseY := ebiten.CursorPosition()
			if ds.opinionHoverRect.Within(mouseX, mouseY) {
				dx := ds.opinionHoverWindow.Bounds().Dx()
				dy := ds.opinionHoverWindow.Bounds().Dy()
				popupX, popupY := utils.GetPositionNearMouse(5, dx, dy)
				rendering.DrawImage(screen, ds.opinionHoverWindow, float64(popupX), float64(popupY), 0)
			}
		}
	}

	// handle drawing replies or topics
	if ds.replyBox != nil {
		// replies are using the larger reply box, since they are too big
		ds.replyBox.draw(screen, textBoxY)
	} else if len(ds.replyButtons) > 0 {
		// replies are drawn in the topic box, since they are all small enough to fit
		replyX := optionBoxX + int(tileSize/2)
		replyY := optionBoxY + int(tileSize/2)

		for i, b := range ds.replyButtons {
			b.Draw(screen, replyX, replyY+(i*buttonHeight))
		}
	} else if len(ds.topicButtons) > 0 {
		for i, b := range ds.topicButtons {
			b.Draw(screen, optionBoxX+int(tileSize/2), optionBoxY+(i*buttonHeight)+(int(tileSize/2)))
		}
	}
}
