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

type DialogVariable string

const (
	VarPlayerName DialogVariable = "{player_name}"
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

	TextBoxImg  *ebiten.Image
	TopicBoxImg *ebiten.Image

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

func ConditionsMet(conditions []defs.Condition, ctx defs.ConditionContext) bool {
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

	profileDef := defMgr.GetDialogProfile(params.ProfileID)
	profileState := defMgr.GetDialogProfileState(params.ProfileID)
	if profileDef == nil {
		panic("defMgr gave back nil profile def")
	}
	if profileState == nil {
		panic("defMgr gave back nil profile state")
	}
	ctx := NewDialogContext(params.NPCID, profileState, gameState)
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

	b := box.NewBox(params.BoxTilesetSrc, params.BoxOriginID)
	ds.TextBoxImg = b.BuildBoxImage(textBoxWidth, textBoxHeight)
	ds.TopicBoxImg = b.BuildBoxImage(int(topicBoxWidth), textBoxHeight)

	ds.LineWriter = text.NewLineWriter(textBoxWidth-tileSize, textBoxHeight-tileSize, params.TextFont, nil, nil, true, false)

	// when starting a new dialog session, we start with the first greeting from the NPC/DialogProfile
	firstGreeting := GetGreeting(*ds.ProfileDef, ds.Ctx)
	ds.ApplyResponse(firstGreeting)

	return ds
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

	for _, topic := range options {
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

	ds.replyButtons = make([]*button.Button, 0)
	ds.replyList = make([]defs.DialogReply, 0)

	tileSize := config.GetScaledTilesize()

	h, _ := text.GetRealisticFontMetrics(ds.f)
	w := ds.TopicBoxImg.Bounds().Dx() - int(tileSize)

	for _, reply := range ds.currentResponse.Replies {
		if ConditionsMet(reply.Conditions, ds.Ctx) {
			continue
		}
		ds.replyButtons = append(ds.replyButtons, button.NewButton(reply.Text, ds.f, w, h))
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

	// response has started, so we don't need topic buttons anymore
	ds.topicButtons = []*button.Button{}
	ds.topicList = []defs.TopicID{}

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

	// TODO: apply effects

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
		allVars := []DialogVariable{VarPlayerName}

		for _, v := range allVars {
			s = strings.ReplaceAll(s, string(v), ds.Ctx.GameState.GetPlayerInfo().PlayerName)
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

	// TODO: apply effects

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

func GetNPCResponse(dt defs.DialogTopic, ctx defs.ConditionContext) defs.DialogResponse {
	for _, resp := range dt.Responses {
		if ConditionsMet(resp.Conditions, ctx) {
			return resp
		}
	}
	panic("no valid response found")
}

func GetGreeting(def defs.DialogProfileDef, ctx defs.ConditionContext) defs.DialogResponse {
	for _, greeting := range def.Greeting {
		if ConditionsMet(greeting.Conditions, ctx) {
			return greeting
		}
	}
	panic("no valid greeting found")
}
