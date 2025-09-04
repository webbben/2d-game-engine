package dialog

type topicStatus string

const (
	// opened for the first time and is waiting for main text to show
	topic_status_showingMainText topicStatus = "showing_main_text"
	// main text is done showing; waiting for next status
	topic_status_mainTextDone topicStatus = "main_text_done"
	// awaiting sub-topic selection (only valid if topic has sub-topics)
	topic_status_awaitSubtopic topicStatus = "await_subtopic"
	// preparing to go back to parent topic, and waiting for final text to show
	topic_status_goingBack topicStatus = "going_back"
	// has returned from a sub-topic
	topic_status_returned topicStatus = "returned"
)

// A Topic represents a node in a dialog/conversation.
// It can have a main text, and then options to take you to a different node of the conversation.
type Topic struct {
	ParentTopic *Topic // parent topic to revert to when this topic has finished. for the "root", this will be nil.
	TopicText   string // text to show for this topic when in a sub-topics list
	MainText    string // text to show when this topic is selected. will show before any associated action is triggered.
	DoneText    string // text to show when this topic has finished and is about to go back to the parent.
	status      topicStatus

	returningFromSubtopic bool   // flag to indicate if we are just returning from a subtopic. implies that main text should not be shown.
	ReturnText            string // text to show when this topic has been returned to from a sub-topic. if previous topic had DoneText, this is ignored.

	SubTopics []*Topic // list of topic options to select and proceed in the dialog
	// topic actions - for when a topic represents an action, rather than just showing text

	IsExitTopic bool   // if true, then activating this topic will exit the dialog.
	OnActivate  func() // a misc function to trigger some kind of action.

	// misc config

	ShowTextImmediately bool // if true, text will display immediately instead of the via a typing animation
}

func (d *Dialog) setTopic(t Topic, isReturning bool) {
	if !d.init {
		panic("dialog must be initialized before setting a topic. otherwise, lineWriter won't exist.")
	}

	d.currentTopic = &t
	t.returningFromSubtopic = isReturning
	if isReturning {
		t.status = topic_status_returned
	} else {
		// this is a new topic
		t.status = topic_status_showingMainText
	}

	d.lineWriter.Clear()
	d.lineWriter.SetSourceText(d.currentTopic.MainText)
}

func (d *Dialog) returnToParentTopic() {
	if d.currentTopic == nil {
		panic("tried to return to parent topic, but current topic is nil!")
	}
	if d.currentTopic.ParentTopic == nil {
		// this is the root topic; end the conversation
		d.EndDialog()
		return
	}

	textToShow := ""
	if d.currentTopic.DoneText != "" {
		textToShow = d.currentTopic.DoneText
	} else if d.currentTopic.ParentTopic.ReturnText != "" {
		textToShow = d.currentTopic.ParentTopic.ReturnText
	}

	// we've already displayed the done/return text, so we are ready to change the topic now
	// or, if no final text is found, just return while leaving the current text on the screen
	if d.currentTopic.status == topic_status_goingBack || textToShow == "" {
		d.setTopic(*d.currentTopic.ParentTopic, true)
		return
	}

	// show the final text
	d.lineWriter.Clear()
	d.lineWriter.SetSourceText(textToShow)
	d.currentTopic.status = topic_status_goingBack
}
