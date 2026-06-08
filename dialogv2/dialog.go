// Package dialogv2 is an improved dialog system compared to the original one
package dialogv2

import (
	"fmt"
	"image/color"
	"regexp"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/webbben/2d-game-engine/audio"
	"github.com/webbben/2d-game-engine/config"
	"github.com/webbben/2d-game-engine/data/datamanager"
	"github.com/webbben/2d-game-engine/data/defs"
	"github.com/webbben/2d-game-engine/data/id"
	"github.com/webbben/2d-game-engine/data/state"
	"github.com/webbben/2d-game-engine/display"
	characterstate "github.com/webbben/2d-game-engine/entity/characterState"
	"github.com/webbben/2d-game-engine/logz"
	"github.com/webbben/2d-game-engine/model"
	"github.com/webbben/2d-game-engine/pubsub"
	"github.com/webbben/2d-game-engine/quest"
	"github.com/webbben/2d-game-engine/screen"
	"github.com/webbben/2d-game-engine/ui/box"
	"github.com/webbben/2d-game-engine/ui/button"
	"github.com/webbben/2d-game-engine/ui/text"
	"github.com/webbben/2d-game-engine/utils"
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
	Goodbye string = "Goodbye"
)

const (
	ActionTypeShowScreen defs.DialogActionType = "show_screen"
)

const (
	VarPlayerName    string = "{player_name}"
	VarPlayerCulture string = "{player_culture}"
)

var AllDialogVariables = []string{
	VarPlayerName, VarPlayerCulture,
}

type GetUserInputActionParams struct {
	ModalTitle        string
	ConfirmButtonText string
}

type ShowScreenActionParams struct {
	ScreenID     defs.ScreenID
	ScreenParams any
}

var quitTopic defs.DialogTopic = defs.DialogTopic{
	ID:     "QUIT",
	Prompt: Goodbye,
}

type DialogSession struct {
	Exit         bool                      // when set, dialog should exit
	ProfileState *state.DialogProfileState // the persisted state for this dialog profile
	ProfileDef   *defs.DialogProfileDef    // the actual dialog profile definition
	Ctx          DialogContext             // the dialog state that things within the dialog can read/write on

	showCharInfo bool

	nameTitle             box.BoxTitle
	boxSrc                box.Box
	TextBoxImg            *ebiten.Image
	TopicBoxImg           *ebiten.Image
	topicBoxDefaultHeight int

	charInfoBoxImg     *ebiten.Image
	npcName            string
	opinionHoverWindow *ebiten.Image // the hover window that shows to show opinion mods. if nil, no mods to show.
	opinionHoverRect   *model.Rect   // the area where, if the mouse is hovering in it, the opinion hover window should start showing
	opinionMods        []defs.OpinionModifier

	LineWriter text.LineWriter

	eventBus *pubsub.EventBus
	dataman  *datamanager.DataManager
	scrMgr   *screen.ScreenManager
	audioman *audio.AudioManager

	flashContinueIcon   bool
	iconFlashTimer      int
	ticksSinceLastClick int

	topicButtons []*button.Button
	topicList    []defs.TopicID
	replyButtons []*button.Button // reply buttons that can appear in the topic box (for small replies only)
	replyBox     *replyBox        // a box to show bigger replies, when a reply needs more space for an entire sentence (or more).
	replyList    []defs.DialogReply

	topicLinks []topicLink // links in the current response's text

	f font.Face // font used for all text in dialog

	// topic flow

	currentTopic    *defs.DialogTopic    // the topic is the root node of the conversation progression we are at
	currentResponse *defs.DialogResponse // the "response" indicates what node we are in the convesation progression, which started from the topic.
	responseStatus  dialogResponseStatus

	// possible action modals

	screenViewer *screen.ScreenViewer
	ctxForScreen defs.GameContext
}

func (ds DialogSession) panicln(args ...any) {
	if ds.currentTopic != nil {
		args = append(args, fmt.Sprintf("CurrentTopic: %s", ds.currentTopic.ID))
	}
	if ds.currentResponse != nil {
		args = append(args, fmt.Sprintf("CurrentResp: \"%s\"", ds.currentResponse.Text))
	}
	logz.Panicln("DialogSession", args...)
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
	ShowCharInfo  bool
	NPCID         string
	ProfileID     defs.DialogProfileID
	BoxTilesetSrc string
	BoxOriginID   int
	TextFont      font.Face
}

func NewDialogSession(
	params DialogSessionParams,
	eventBus *pubsub.EventBus,
	dataman *datamanager.DataManager,
	scrMgr *screen.ScreenManager,
	gameCtx defs.GameContext,
	questman *quest.QuestManager,
	audioman *audio.AudioManager,
) DialogSession {
	validateParams(params, eventBus, dataman, gameCtx, scrMgr)

	// Note: we expect the profile state to already have been made. If we decide to have "ad hoc" dialog profiles, this should change.

	profileDef := dataman.GetDialogProfile(params.ProfileID)
	profileState := dataman.GetDialogProfileState(params.ProfileID)

	ctx := NewDialogContext(params.NPCID, profileState, *profileDef, gameCtx, eventBus, dataman, questman)
	ds := DialogSession{
		showCharInfo: params.ShowCharInfo,
		scrMgr:       scrMgr,
		ctxForScreen: gameCtx,
		ProfileState: profileState,
		ProfileDef:   profileDef,
		Ctx:          ctx,
		eventBus:     eventBus,
		dataman:      dataman,
		audioman:     audioman,
		f:            params.TextFont,
	}

	ds.dialogSetup(params)

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

func validateParams(params DialogSessionParams, eventBus *pubsub.EventBus, dataman *datamanager.DataManager, gameCtx defs.GameContext, scrMgr *screen.ScreenManager) {
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
	if dataman == nil {
		panic("dataman was nil")
	}
	if params.TextFont == nil {
		panic("text font was nil")
	}
	if gameCtx == nil {
		panic("game state was nil")
	}
	if scrMgr == nil {
		panic("scrmgr was nil")
	}
}

func (ds *DialogSession) dialogSetup(params DialogSessionParams) {
	tileSize := int(config.GetScaledTilesize())
	topicBoxWidth := tileSize * 8
	screenWidth := utils.RoundDownToTile(display.SCREEN_WIDTH, tileSize)
	textBoxWidth := screenWidth - topicBoxWidth
	textBoxWidth -= textBoxWidth % tileSize
	textBoxHeight := tileSize * 6
	ds.topicBoxDefaultHeight = textBoxHeight

	b := box.NewBox(params.BoxTilesetSrc, params.BoxOriginID)
	ds.boxSrc = b

	npcState := ds.dataman.GetCharacterState(id.CharacterStateID(params.NPCID))
	npcName := npcState.DisplayName

	ds.nameTitle = box.NewBoxTitle(params.BoxTilesetSrc, 111, npcName, config.DefaultTitleFont)

	ds.TextBoxImg = b.BuildBoxImage(textBoxWidth, textBoxHeight, config.UIScale)
	ds.buildTopicBox(textBoxHeight)

	ds.charInfoBoxImg = b.BuildBoxImage(tileSize*8, textBoxHeight, config.UIScale)
	ds.npcName = npcName
	npcDef := ds.dataman.GetCharacterDef(npcState.DefID)
	cultureDef := ds.dataman.GetCultureDef(npcDef.CultureID)
	ds.Ctx.culture = cultureDef
	playerState := ds.dataman.GetCharacterState(id.CharacterStateID(defs.PlayerID))
	currentTime := ds.Ctx.GetCurrentGameTime()
	// TODO: add some kind of hover window to show the specific opinion modifiers
	// Also, if anything during dialog can cause a new opinion modifier, we need to recalculate this
	ds.opinionMods, ds.Ctx.opinion = characterstate.CalculateOpinion(npcState, playerState, currentTime, ds.dataman)
	ds.buildOpinionHoverBox()

	lwParams := config.LineWriterParams{
		LineWidthPx:           textBoxWidth - tileSize,
		MaxHeightPx:           textBoxHeight - tileSize,
		FontFace:              params.TextFont,
		UseShadow:             true,
		TextBlipSfx:           config.DefaultTextBlipSfx,
		TextBlipTickInterval:  5,
		SupportSpecialSymbols: true,
	}
	ds.LineWriter = text.NewLineWriter(ds.audioman, lwParams)
}

func (ds *DialogSession) buildOpinionHoverBox() {
	if len(ds.opinionMods) == 0 {
		ds.opinionHoverWindow = nil
		return
	}

	b := box.NewBox("boxes/boxes.tsj", 231)

	// figure out how big the box should be
	marginY := 5
	maxDx := 0
	totalDy := 0
	for _, mod := range ds.opinionMods {
		dx, dy, _ := text.GetStringSize(mod.String(), config.DefaultInfoFont)
		maxDx = max(maxDx, dx)
		totalDy += dy + marginY
	}

	lineDy, _ := text.GetRealisticFontMetrics(config.DefaultInfoFont)
	tilesize := config.GetScaledTilesize()
	maxDx = utils.RoundUpToTile(maxDx, int(tilesize))
	totalDy = utils.RoundUpToTile(totalDy, int(tilesize))

	// for the box's requirements
	totalDy += int(tilesize * 2) //  the box is pretty thin on the border tiles
	totalDy = max(totalDy, int(tilesize*3))
	maxDx += int(tilesize * 2)

	ds.opinionHoverWindow = b.BuildBoxImage(maxDx, totalDy, config.UIScale)

	// now, draw the opinion mod text
	drawX := int(tilesize)
	drawY := int(tilesize + float64(lineDy))
	for _, mod := range ds.opinionMods {
		c := color.RGBA{0, 255, 0, 0}
		if mod.Mod < 0 {
			c = color.RGBA{255, 0, 0, 0}
		}
		text.DrawText(ds.opinionHoverWindow, mod.String(), config.DefaultInfoFont, drawX, drawY, c)
		drawY += lineDy + marginY
	}
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

	ds.TopicBoxImg = ds.boxSrc.BuildBoxImage(int(topicBoxWidth), textBoxHeight, config.UIScale)
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
		dx, _, _ := text.GetStringSize(topic.Prompt, ds.f)
		if dx > w {
			// TODO: probably at some point we can just make this a warn instead of panic. just want this here for now to catch oversized topic prompts early on.
			logz.Panicln("setupTopicOptions", "topic prompt was too long for the topic box. prompt width:", dx, "boxWidth:", w)
		}
		ds.topicButtons = append(ds.topicButtons, button.NewButton(topic.Prompt, ds.f, w, h, ds.audioman))
		ds.topicList = append(ds.topicList, topic.ID)
	}

	// if we need to increase the size of the topic box, do so here.
	// if a previous set of topics caused the topic box to enlarge, but it's no longer necessary, shrink it back down to default size here too.
	topicTotalHeight := len(options) * ds.topicButtons[0].Height
	ds.calculateTopicBoxSize(topicTotalHeight)

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
	optionsTotalHeight += int(tileSize / 2) // add extra space above and below the options, for the edge box tiles
	optionsTotalHeight = utils.RoundUpToTile(optionsTotalHeight, int(tileSize))

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

	tileSize := config.GetScaledTilesize()

	h, _ := text.GetRealisticFontMetrics(ds.f)
	w := ds.TopicBoxImg.Bounds().Dx() - int(tileSize)
	maxReplyWidth := w

	if ds.currentResponse.Goodbye {
		// setup the 'Goodbye' reply
		ds.replyButtons = make([]*button.Button, 0)
		ds.replyList = make([]defs.DialogReply, 0)
		ds.replyList = append(ds.replyList, defs.DialogReply{
			Text:    Goodbye,
			Goodbye: true,
		})
		ds.replyButtons = append(ds.replyButtons, button.NewButton(Goodbye, ds.f, maxReplyWidth, h, ds.audioman))
		return
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

	hasFancyReplies := false

	// first, find out if all replies can fit in the topic box, and get the maximum reply width.
	replies := []defs.DialogReply{}
	for _, reply := range ds.currentResponse.Replies {
		if ConditionsMet(reply.Conditions, ds.Ctx) {
			replies = append(replies, reply)
			dx, _, _ := text.GetStringSize(reply.Text, ds.f)
			if dx > maxReplyWidth {
				maxReplyWidth = dx
			}
			if reply.InfoText != nil || reply.Decoration != "" {
				hasFancyReplies = true
			}
		}
	}

	if len(replies) == 0 {
		ds.panicln("trying to setup replies, but no valid replies were found (none met their conditions)")
	}

	ds.replyList = replies

	// now that we have all valid replies and the max width, create the buttons, and if needed, the larger replies box.
	// if any of the replies have decorations, info, etc then use the large reply box too
	if maxReplyWidth > w || hasFancyReplies {
		ds.setupReplyBox(maxReplyWidth, replies)
		return
	}

	totalHeight := max(len(replies)*h, int(tileSize)*4)

	// using regular topic box; make sure it is the correct height
	ds.replyBox = nil // if set to nil, we know not to use it at draw time
	ds.calculateTopicBoxSize(totalHeight)

	// build all the buttons, now that we know the width to use
	for _, reply := range replies {
		ds.replyButtons = append(ds.replyButtons, button.NewButton(reply.Text, ds.f, maxReplyWidth, h, ds.audioman))
	}

	if len(ds.replyButtons) == 0 {
		panic("setting up reply buttons, but no valid replies were found")
	}
}

func (ds *DialogSession) GetTopicOptions() []defs.DialogTopic {
	seenTopics := make(map[defs.TopicID]bool) // ensure no duplicates
	topicOptions := []defs.DialogTopic{}

	// first, get them from the profile
	for _, topicID := range ds.ProfileDef.TopicsIDs {
		if seenTopics[topicID] {
			continue
		}
		seenTopics[topicID] = true

		topic := ds.dataman.GetDialogTopic(topicID)
		if ConditionsMet(topic.Conditions, ds.Ctx) {
			topicOptions = append(topicOptions, *topic)
		}
	}

	// next, get knowledge topics that both player and NPC have access to
	for _, topicID := range ds.Ctx.GetKnowledgeTopics() {
		if seenTopics[topicID] {
			continue
		}
		seenTopics[topicID] = true

		topic := ds.dataman.GetDialogTopic(topicID)
		// TODO: I wonder if we should find a way to avoid checking conditions everytime.
		// one idea is to cache the result, and only recalculate whenever any effect happens, since that's probably the only
		// time conditions could be affected.
		if ConditionsMet(topic.Conditions, ds.Ctx) {
			topicOptions = append(topicOptions, *topic)
		}
	}

	return topicOptions
}

func (ds *DialogSession) ApplyResponse(dr defs.DialogResponse) {
	if ds.LineWriter.IsWriting() {
		panic("tried to apply response, but linewriter is already writing")
	}
	if dr.ID != "" && dr.Once && ds.Ctx.HasSeenResponse(dr.ID) {
		ds.panicln("tried to apply response that is only supposed to be seen once", "response ID:", dr.ID)
	}

	// response has started, so we don't need topic buttons anymore
	ds.topicButtons = []*button.Button{}
	ds.topicList = []defs.TopicID{}
	ds.topicLinks = []topicLink{}

	// find any topic links that might exist in the text
	ds.topicLinks = parseTextLinks(dr)

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

// topicLink is used to denote a link that is found in a dialog response text.
type topicLink struct {
	text    string       // text of the actual link
	topicID defs.TopicID // topic that it links to

	// the following is set after linewriter has finished writing text and knows all link positions

	linkButton *button.Button
	x, y       float64
}

func parseTextLinks(dr defs.DialogResponse) []topicLink {
	// find all "[...]" sections of the text, and record those as topic links.
	linkCandidates := []string{}
	re := regexp.MustCompile(`\[([^\]]+)\]`)

	matches := re.FindAllStringSubmatch(dr.Text, -1)

	if len(matches) == 0 {
		// no topic notation found
		return []topicLink{}
	}

	for _, match := range matches {
		// match[1] has the string stripped of square brackets
		linkCandidates = append(linkCandidates, match[1])
	}

	topicLinks := []topicLink{}

	// now, we associate a link candidate the topic IDs in NextTopics.
	// we assume nextTopics should be matched with link candidates in the same order as set in the slice.
	// panic if there are not enough nextTopics for the number of link candidates, but don't panic if there are more nextTopics than link candidates.
	// this is because every link candidate ("[...]") must have a topic assosciated, but we also allow to introduce next topics without including them in
	// the dialog text itself.
	if len(linkCandidates) > len(dr.NextTopics) {
		logz.Panicln("parseTextLinks", "more link candidates found than NextTopics")
	}
	for i, linkText := range linkCandidates {
		topicLinks = append(topicLinks, topicLink{
			text:    linkText,
			topicID: dr.NextTopics[i],
		})
	}

	return topicLinks
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
	case ActionTypeShowScreen:
		params, ok := action.Params.(ShowScreenActionParams)
		if !ok {
			panic("unable to resolve params as ShowScreenActionParams... was the wrong params type chosen?")
		}
		s := ds.scrMgr.GetScreen(params.ScreenID)
		sv := screen.NewScreenViewer(s, ds.dataman, ds.eventBus, ds.audioman, ds.Ctx.questman, ds.ctxForScreen, params.ScreenParams)
		ds.screenViewer = &sv
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

	for _, topicID := range ds.currentResponse.NextTopics {
		ds.Ctx.RecordTopicUnlocked(topicID)
	}

	for _, effect := range ds.currentResponse.Effects {
		effect.Apply(&ds.Ctx)
	}
	for _, effect := range ds.currentResponse.WorldEffects {
		effect.Apply(&ds.Ctx)
	}

	if ds.currentResponse.Exit {
		ds.Exit = true
		return
	}

	// If there is no text, check if this is a grouper;
	// even if it isn't a grouper, that's okay - just let the dialog be empty.
	if ds.currentResponse.Text == "" {
		ds.handleResponseGrouper()
		return
	}

	ds.setResponseText(ds.currentResponse.Text)

	// reply handling is done once text is finished

	ds.responseStatus = dialogResponseStarted
}

func (ds *DialogSession) handleResponseGrouper() {
	if ds.currentResponse.Goodbye {
		ds.panicln("response grouper was set to Goodbye. Goodbye should only be used after specific text has been shown, so a 'goodbye' reply can show")
	}

	// find the appropriate response option
	if ds.currentResponse.NextResponse != nil {
		ds.ApplyResponse(*ds.currentResponse.NextResponse)
		return
	}

	if len(ds.currentResponse.NextResponseOptions) > 0 {
		resp := chooseResponse(ds.currentResponse.NextResponseOptions, ds.Ctx)
		ds.ApplyResponse(resp)
		return
	}
}

func (ds *DialogSession) setResponseText(s string) {
	if s == "" {
		panic("tried to set an empty string")
	}

	ds.LineWriter.Clear()

	// fill in any dialog variables
	s = InsertDialogVariables(s, ds.Ctx.GameState.GetPlayerInfo(), ds.dataman)

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
	ds.replyBox = nil

	for _, effect := range dr.Effects {
		effect.Apply(&ds.Ctx)
	}
	for _, effect := range dr.WorldEffects {
		effect.Apply(&ds.Ctx)
	}

	if dr.Goodbye {
		// this is a "goodbye" reply, and so it should end the dialog
		ds.End()
		return
	}

	if dr.NextResponse == nil {
		if len(dr.NextResponseOptions) > 0 {
			resp := chooseResponse(dr.NextResponseOptions, ds.Ctx)
			ds.ApplyResponse(resp)
			return
		}
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
		ds.End()
		return
	}
	topic := ds.dataman.GetDialogTopic(topicID)
	ds.currentTopic = topic

	ds.Ctx.RecordTopicSeen(topicID)

	// get the linewriter started on writing
	resp := GetNPCResponse(*ds.currentTopic, ds.Ctx)
	ds.ApplyResponse(resp)

	if ds.responseStatus != dialogResponseStarted && ds.responseStatus != dialogResponseActionInProg {
		// after applying a response, we expect to either start the response or an action, if applicable
		logz.Panicln("SetTopic", "unexpected status after applying topic response:", ds.responseStatus)
	}
}

// End causes the DialogSession to end.
func (ds *DialogSession) End() {
	ds.Exit = true
	ds.eventBus.Publish(defs.Event{
		Type: pubsub.EventDialogEnded,
		Data: map[string]any{
			"profileID": ds.ProfileDef.ProfileID,
		},
	})
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
