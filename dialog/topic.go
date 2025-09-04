package dialog

// A Topic represents a node in a dialog/conversation.
// It can have a main text, and then options to take you to a different node of the conversation.
type Topic struct {
	ParentTopic *Topic   // parent topic to revert to when this topic has finished. for the "root", this will be nil.
	TopicText   string   // text to show for this topic when in a sub-topics list
	MainText    string   // text to show when this topic is selected. will show before any associated action is triggered.
	DoneText    string   // text to show when this topic has finished and is about to go back to the parent.
	ReturnText  string   // text to show when this topic has been returned to from a sub-topic.
	SubTopics   []*Topic // list of topic options to select and proceed in the dialog

	// topic actions - for when a topic represents an action, rather than just showing text

	IsExitTopic bool   // if true, then activating this topic will exit the dialog.
	OnActivate  func() // a misc function to trigger some kind of action.

	// misc config

	ShowTextImmediately bool // if true, text will display immediately instead of the via a typing animation
}

func (d *Dialog) setTopic(t Topic) {
	d.currentTopic = &t

	if !d.init {
		panic("dialog must be initialized before setting a topic. otherwise, lineWriter won't exist.")
	}

	d.lineWriter.SetSourceText(d.currentTopic.MainText)
}
