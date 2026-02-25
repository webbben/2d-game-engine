// Package dialogv2 is an improved dialog system compared to the original one
package dialogv2

import (
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/webbben/2d-game-engine/data/defs"
	"github.com/webbben/2d-game-engine/data/state"
	"github.com/webbben/2d-game-engine/definitions"
	"github.com/webbben/2d-game-engine/internal/config"
	"github.com/webbben/2d-game-engine/internal/display"
	"github.com/webbben/2d-game-engine/internal/logz"
	"github.com/webbben/2d-game-engine/internal/pubsub"
	"github.com/webbben/2d-game-engine/internal/text"
	"github.com/webbben/2d-game-engine/internal/ui/box"
	"github.com/webbben/2d-game-engine/internal/ui/button"
	"github.com/webbben/2d-game-engine/internal/ui/modal"
	"golang.org/x/image/font"
)

type dialogResponseStatus string

const (
	dialogResponseStarted      dialogResponseStatus = "started"    // text portion of response has been started, and waiting for it to finish writing to line writer
	dialogResponseTextDone     dialogResponseStatus = "text done"  // text portion of response is done; ready for player replies or next response
	dialogResponseUserReply    dialogResponseStatus = "user reply" // awaiting user reply
	dialogResponseFinished     dialogResponseStatus = "finished"   // this dialog response is completely finished, and there is nothing else to do here
	dialogResponseActionInProg dialogResponseStatus = "action"     // a dialog action was triggered and is waiting to finish
)

const (
	ActionTypeGetUserInput defs.DialogActionType        = "get_user_input"
	ActionScopePlayerName  defs.DialogActionResultScope = "player_name"
)

const (
	VarPlayerName string = "{player_name}"
)

type GetUserInputActionParams struct {
	ModalTitle        string
	ConfirmButtonText string
}

var quitTopic defs.DialogTopic = defs.DialogTopic{
	ID:     "QUIT",
	Prompt: "Goodbye",
}

type DialogSession struct {
	Exit         bool                      // when set, dialog should exit
	ProfileState *state.DialogProfileState // the persisted state for this dialog profile
	ProfileDef   *defs.DialogProfileDef    // the actual dialog profile definition
	Ctx          DialogContext             // the dialog state that things within the dialog can read/write on

	boxSrc                box.Box
	TextBoxImg            *ebiten.Image
	TopicBoxImg           *ebiten.Image
	topicBoxDefaultHeight int
	BigReplyBoxImg        *ebiten.Image

	LineWriter text.LineWriter

	eventBus *pubsub.EventBus
	defMgr   *definitions.DefinitionManager

	flashContinueIcon   bool
	iconFlashTimer      int
	ticksSinceLastClick int

	topicButtons []*button.Button
	topicList    []defs.TopicID
	replyButtons []*button.Button
	replyList    []defs.DialogReply

	f font.Face // font used for all text in dialog

	// topic flow

	currentTopic    *defs.DialogTopic    // the topic is the root node of the conversation progression we are at
	currentResponse *defs.DialogResponse // the "response" indicates what node we are in the convesation progression, which started from the topic.
	responseStatus  dialogResponseStatus

	// possible action modals

	userInputModal *modal.TextInputModal
}

func ConditionsMet(conditions []defs.DialogCondition, ctx defs.ConditionContext) bool {
	for _, cond := range conditions {
		if !cond.IsMet(ctx) {
			return false
		}
	}

	return true
}

type DialogSessionParams struct {
	NPCID         string
	ProfileID     defs.DialogProfileID
	BoxTilesetSrc string
	BoxOriginID   int
	TextFont      font.Face
}

func NewDialogSession(params DialogSessionParams, eventBus *pubsub.EventBus, defMgr *definitions.DefinitionManager, gameState GameStateContext) DialogSession {
	if params.ProfileID == "" {
		panic("profile ID was empty")
	}
	if params.NPCID == "" {
		panic("npcID was empty")
	}
	if params.BoxTilesetSrc == "" {
		panic("box tileset src is empty")
	}
	if eventBus == nil {
		panic("event bus was nil")
	}
	if defMgr == nil {
		panic("defMgr was nil")
	}
	if params.TextFont == nil {
		panic("text font was nil")
	}
	if gameState == nil {
		panic("game state was nil")
	}

	// Note: we expect the profile state to already have been made. If we decide to have "ad hoc" dialog profiles, this should change.

	profileDef := defMgr.GetDialogProfile(params.ProfileID)
	profileState := defMgr.GetDialogProfileState(params.ProfileID)

	ctx := NewDialogContext(params.NPCID, profileState, gameState, eventBus)
	ds := DialogSession{
		ProfileState: profileState,
		ProfileDef:   profileDef,
		Ctx:          ctx,
		eventBus:     eventBus,
		defMgr:       defMgr,
		f:            params.TextFont,
	}

	tileSize := int(config.GetScaledTilesize())
	topicBoxWidth := tileSize * 8
	textBoxWidth := display.SCREEN_WIDTH - tileSize - topicBoxWidth
	textBoxWidth -= textBoxWidth % tileSize
	textBoxHeight := tileSize * 6
	ds.topicBoxDefaultHeight = textBoxHeight

	b := box.NewBox(params.BoxTilesetSrc, params.BoxOriginID)
	ds.boxSrc = b

	ds.TextBoxImg = b.BuildBoxImage(textBoxWidth, textBoxHeight)
	ds.buildTopicBox(textBoxHeight)

	ds.LineWriter = text.NewLineWriter(textBoxWidth-tileSize, textBoxHeight-tileSize, params.TextFont, nil, nil, true, false)

	// when starting a new dialog session, we start with the first greeting from the NPC/DialogProfile
	firstGreeting := GetGreeting(*ds.ProfileDef, ds.Ctx)
	ds.ApplyResponse(firstGreeting)

	ds.eventBus.Publish(defs.Event{
		Type: pubsub.EventDialogStarted,
		Data: map[string]any{
			"profileID": ds.ProfileDef.ProfileID,
		},
	})

	return ds
}

// builds the topic box to fit the given height. has a minimum height it defaults to.
func (ds *DialogSession) buildTopicBox(height int) {
	if ds.topicBoxDefaultHeight <= 0 {
		panic("topic box default height was <= 0")
	}

	tileSize := int(config.GetScaledTilesize())
	topicBoxWidth := tileSize * 8

	height -= height % tileSize

	textBoxHeight := max(height, ds.topicBoxDefaultHeight)

	ds.TopicBoxImg = ds.boxSrc.BuildBoxImage(int(topicBoxWidth), textBoxHeight)
}

func (ds *DialogSession) setupTopicOptions() {
	if len(ds.replyButtons) > 0 {
		panic(`
			Tried to setup topic buttons, but there are already reply buttons. You should erase the reply buttons at the point where they are no longer needed,
			before trying to setup topic buttons.
		`)
	}
	options := ds.GetTopicOptions()
	options = append(options, quitTopic)

	ds.topicButtons = []*button.Button{}
	ds.topicList = []defs.TopicID{}

	tileSize := config.GetScaledTilesize()

	h, _ := text.GetRealisticFontMetrics(ds.f)
	w := ds.TopicBoxImg.Bounds().Dx() - int(tileSize)

	// if we need to increase the size of the topic box, do so here.
	// if a previous set of topics caused the topic box to enlarge, but it's no longer necessary, shrink it back down to default size here too.
	topicTotalHeight := h * len(options)
	ds.calculateTopicBoxSize(topicTotalHeight)

	for _, topic := range options {
		dx, _, _ := text.GetStringSize(topic.Prompt, ds.f)
		if dx > w {
			// TODO: probably at some point we can just make this a warn instead of panic. just want this here for now to catch oversized topic prompts early on.
			logz.Panicln("setupTopicOptions", "topic prompt was too long for the topic box. prompt width:", dx, "boxWidth:", w)
		}
		ds.topicButtons = append(ds.topicButtons, button.NewButton(topic.Prompt, ds.f, w, h))
		ds.topicList = append(ds.topicList, topic.ID)
	}

	// sanity check: ensure topics actually exist
	if len(ds.topicButtons) == 0 || len(ds.topicList) == 0 {
		panic("no topics?")
	}
	if len(ds.topicButtons) != len(ds.topicList) {
		panic("number of topic buttons does not equal number of topics?")
	}
}

func (ds *DialogSession) calculateTopicBoxSize(optionsTotalHeight int) {
	tileSize := config.GetScaledTilesize()
	topicBoxCurrentHeight := ds.TopicBoxImg.Bounds().Bounds().Dy()
	optionsTotalHeight -= optionsTotalHeight % int(tileSize)
	optionsTotalHeight += int(tileSize)

	if optionsTotalHeight > topicBoxCurrentHeight-int(tileSize) {
		// need to expand
		ds.buildTopicBox(optionsTotalHeight)
	} else {
		// check if we should shrink?
		if optionsTotalHeight < topicBoxCurrentHeight {
			if topicBoxCurrentHeight > ds.topicBoxDefaultHeight {
				// we can shrink back down, since it's taller than the default.
				ds.buildTopicBox(optionsTotalHeight)
			}
		}
	}
}

func (ds *DialogSession) setupReplyOptions() {
	if len(ds.topicButtons) > 0 {
		panic(`
			Tried to setup reply buttons, but there are already topic buttons. You should erase the topic buttons at the point where they are no longer needed,
			before trying to setup reply buttons. (Hint: if this is happening, that means whenever a topic was started and the NPC response was set, you never
			unset the topic buttons.
			`)
	}
	if ds.responseStatus != dialogResponseUserReply {
		panic("tried to setup reply buttons, but dialog status is incorrect")
	}
	if len(ds.currentResponse.Replies) == 0 {
		panic("tried to setup reply buttons, but there are no reply options")
	}

	// reply options can show in two possible places:
	//
	// 1) in the topic box
	// - when all replies are short enough to fit, like simple "yes/no" or answers with only a few words.
	//
	// 2) a larger reply box that appears above the dialog box
	// - when (at least one of) the replies are too long to fit in the topic box.
	// - e.g. when the player needs to give a lengthier reply that is an entire sentence.

	ds.replyButtons = make([]*button.Button, 0)
	ds.replyList = make([]defs.DialogReply, 0)

	tileSize := config.GetScaledTilesize()

	h, _ := text.GetRealisticFontMetrics(ds.f)
	w := ds.TopicBoxImg.Bounds().Dx() - int(tileSize)
	maxReplyWidth := w

	// first, find out if all replies can fit in the topic box, and get the maximum reply width.
	replies := []defs.DialogReply{}
	for _, reply := range ds.currentResponse.Replies {
		if ConditionsMet(reply.Conditions, ds.Ctx) {
			replies = append(replies, reply)
			dx, _, _ := text.GetStringSize(reply.Text, ds.f)
			if dx > maxReplyWidth {
				maxReplyWidth = dx
			}
		}
	}

	totalHeight := max(len(replies)*h, int(tileSize)*4)
	// now that we have all valid replies and the max width, create the buttons, and if needed, the larger replies box.
	if maxReplyWidth > w {
		maxReplyWidth -= maxReplyWidth % int(tileSize)
		maxReplyWidth += int(tileSize)
		totalHeight -= totalHeight % int(tileSize)
		totalHeight += int(tileSize)
		// create the big replies box
		// we add the extra padding here since we want the reply box to be wider, but not the buttons themselves
		ds.BigReplyBoxImg = ds.boxSrc.BuildBoxImage(maxReplyWidth+int(tileSize), totalHeight)
	} else {
		// using regular topic box; make sure it is the correct height
		ds.BigReplyBoxImg = nil // if set to nil, we know not to use it at draw time
		ds.calculateTopicBoxSize(totalHeight)
	}

	// build all the buttons, now that we know the width to use
	for _, reply := range replies {
		ds.replyButtons = append(ds.replyButtons, button.NewButton(reply.Text, ds.f, maxReplyWidth, h))
		ds.replyList = append(ds.replyList, reply)
	}

	if len(ds.replyButtons) == 0 {
		panic("setting up reply buttons, but no valid replies were found")
	}
}

func (ds *DialogSession) GetTopicOptions() []defs.DialogTopic {
	// use a map at first to ensure de-duplication
	topics := make(map[defs.TopicID]defs.DialogTopic)
	topicOptions := []defs.DialogTopic{}

	// first, get them from the profile
	for _, topicID := range ds.ProfileDef.TopicsIDs {
		topic := ds.defMgr.GetDialogTopic(topicID)
		if ConditionsMet(topic.Conditions, ds.Ctx) {
			topics[topicID] = *topic
		}
	}

	// next, go through unlocked topics
	for _, topicID := range ds.Ctx.GetUnlockedTopics() {
		topic := ds.defMgr.GetDialogTopic(topicID)
		if ConditionsMet(topic.Conditions, ds.Ctx) {
			topics[topicID] = *topic
		}
	}

	// convert map to slice
	for _, topic := range topics {
		topicOptions = append(topicOptions, topic)
	}

	return topicOptions
}

func (ds *DialogSession) ApplyResponse(dr defs.DialogResponse) {
	if ds.LineWriter.IsWriting() {
		panic("tried to apply response, but linewriter is already writing")
	}
	if dr.ID != "" && dr.Once && ds.Ctx.HasSeenResponse(dr.ID) {
		logz.Panicln("Dialog", "tried to apply response that is only supposed to be seen once")
	}

	// response has started, so we don't need topic buttons anymore
	ds.topicButtons = []*button.Button{}
	ds.topicList = []defs.TopicID{}

	// if the response has an ID, mark it as seen
	if dr.ID != "" {
		ds.Ctx.RecordResponseSeen(dr.ID)
	}

	ds.currentResponse = &dr

	// check for actions; if so, then we fire the action first and wait for it to finish
	if dr.Action != nil {
		ds.startAction()
		return
	}

	ds.continueApplyResponse()
}

func (ds *DialogSession) startAction() {
	if ds.LineWriter.IsWriting() {
		panic("tried to start action, but linewriter is writing")
	}
	if ds.currentResponse.Action == nil {
		panic("starting action, but no action found on current response")
	}
	ds.responseStatus = dialogResponseActionInProg

	action := ds.currentResponse.Action

	switch action.Type {
	case ActionTypeGetUserInput:
		params, ok := action.Params.(GetUserInputActionParams)
		if !ok {
			panic("unable to cast params as GetUserInputActionParams... was the wrong params type chosen?")
		}
		m := modal.NewTextInputModal(modal.TextInputModalParams{
			BoxTilesetSrc:     config.DefaultUIBox.TilesetSrc,
			BoxOriginIndex:    config.DefaultUIBox.OriginIndex,
			TitleText:         params.ModalTitle,
			ConfirmButtonText: params.ConfirmButtonText,
		})
		ds.userInputModal = &m
	default:
		logz.Panicln("startAction", "action type not recognized:", action.Type)
	}
}

func (ds *DialogSession) continueApplyResponse() {
	if ds.LineWriter.IsWriting() {
		panic("tried to continue apply response, but linewriter is writing")
	}
	if ds.currentResponse == nil {
		panic("continueApplyResponse: currentResponse is nil")
	}

	ds.setResponseText(ds.currentResponse.Text)

	for _, topicID := range ds.currentResponse.NextTopics {
		ds.Ctx.RecordTopicUnlocked(topicID)
	}

	for _, effect := range ds.currentResponse.Effects {
		effect.Apply(&ds.Ctx)
	}

	// reply handling is done once text is finished

	ds.responseStatus = dialogResponseStarted
}

func (ds *DialogSession) setResponseText(s string) {
	if s == "" {
		panic("tried to set an empty string")
	}

	ds.LineWriter.Clear()

	// detect if there are any variables to fill in
	if strings.Contains(s, "{") {
		allVars := []string{VarPlayerName}

		for _, v := range allVars {
			s = strings.ReplaceAll(s, v, ds.Ctx.GameState.GetPlayerInfo().PlayerName)
		}
	}

	ds.LineWriter.SetSourceText(s)
}

func (ds *DialogSession) ApplyReply(dr defs.DialogReply) {
	if ds.responseStatus != dialogResponseUserReply {
		panic("applying reply while status is incorrect")
	}
	if ds.LineWriter.IsWriting() {
		panic("applying reply while linewriter is still writing")
	}

	// the player has chosen a reply, so we no longer need the reply buttons.
	ds.replyButtons = []*button.Button{}
	ds.replyList = []defs.DialogReply{}

	for _, effect := range dr.Effects {
		effect.Apply(&ds.Ctx)
	}

	if dr.NextResponse == nil {
		// no response to reply... so, time to go back to topic selection
		ds.responseStatus = dialogResponseFinished
		return
	}

	// setup next response
	ds.ApplyResponse(*dr.NextResponse)
}

func (ds *DialogSession) SetTopic(topicID defs.TopicID) {
	// first, check that we are ready to accept new topics
	if ds.responseStatus != dialogResponseFinished {
		panic("setting topic when dialogresponse is not finished")
	}
	if ds.LineWriter.IsWriting() {
		panic("tried to set topic while line writer is already writing")
	}
	if topicID == "QUIT" {
		// end dialog
		ds.Exit = true
		ds.eventBus.Publish(defs.Event{
			Type: pubsub.EventDialogEnded,
			Data: map[string]any{
				"profileID": ds.ProfileDef.ProfileID,
			},
		})
		return
	}
	topic := ds.defMgr.GetDialogTopic(topicID)
	ds.currentTopic = topic

	ds.Ctx.RecordTopicSeen(topicID)

	// get the linewriter started on writing
	resp := GetNPCResponse(*ds.currentTopic, ds.Ctx)
	ds.ApplyResponse(resp)

	if ds.responseStatus != dialogResponseStarted {
		panic("why isn't status response started?")
	}
}

func GetNPCResponse(dt defs.DialogTopic, ctx DialogContext) defs.DialogResponse {
	return chooseResponse(dt.Responses, ctx)
}

func GetGreeting(def defs.DialogProfileDef, ctx DialogContext) defs.DialogResponse {
	return chooseResponse(def.Greeting, ctx)
}

func chooseResponse(responses []defs.DialogResponse, ctx DialogContext) defs.DialogResponse {
	for _, response := range responses {
		if response.ID != "" && response.Once && ctx.HasSeenResponse(response.ID) {
			continue
		}
		if ConditionsMet(response.Conditions, ctx) {
			return response
		}
	}
	panic("no valid response found")
}
