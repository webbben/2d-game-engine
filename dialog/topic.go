package dialog

import (
	"github.com/webbben/2d-game-engine/internal/logz"
	"github.com/webbben/2d-game-engine/internal/pubsub"
	"github.com/webbben/2d-game-engine/internal/ui/button"
)

type topicStatus string

const (
	// opened for the first time and is waiting for main text to show
	topic_status_showingMainText topicStatus = "showing_main_text"
	// main text is done showing; waiting for next status
	topic_status_mainTextDone topicStatus = "main_text_done"
	// main text has shown, but the text branch has options for the user to choose
	topic_status_awaitOption topicStatus = "await_option"
	// awaiting sub-topic selection (only valid if topic has sub-topics)
	topic_status_awaitSubtopic topicStatus = "await_subtopic"
	// preparing to go back to parent topic, and waiting for final text to show
	topic_status_goingBack topicStatus = "going_back"
	// has returned from a sub-topic
	topic_status_returned topicStatus = "returned"
)

var goodbyeTopic Topic = Topic{
	TopicText:        "Goodbye",
	isEndDialogTopic: true,
}

// A Topic represents a node in a dialog/conversation.
// It can have a main text, and then options to take you to a different node of the conversation.
type Topic struct {
	ID               string
	isEndDialogTopic bool // flag indicates if this topic should end the conversation

	parentTopic       *Topic      // when switching to a sub-topic, this will be populated so that we know which topic to revert back to
	TopicText         string      // text to show for this topic when in a sub-topics list
	MainText          TextBranch  // text to show when this topic is selected. will show before any associated action is triggered.
	currentTextBranch *TextBranch // the current text branch being shown; starts out as the MainText text branch
	DoneText          string      // text to show when this topic has finished and is about to go back to the parent.
	status            topicStatus

	ReturnText string // text to show when this topic has been returned to from a sub-topic. if previous topic had DoneText, this is ignored.

	SubTopics []Topic // list of topic options to select and proceed in the dialog
	// topic actions - for when a topic represents an action, rather than just showing text

	IsExitTopic bool   // if true, then activating this topic will exit the dialog.
	OnActivate  func() // a misc function to trigger some kind of action.

	// misc config

	ShowTextImmediately bool // if true, text will display immediately instead of the via a typing animation

	button *button.Button // a button for this topic, if it's a subtopic
}

func (d *Dialog) setCurrentTextBranch(textBranch TextBranch) {
	if d.currentTopic == nil {
		panic("current topic is nil")
	}

	d.currentTopic.currentTextBranch = &textBranch

	for i := range d.currentTopic.currentTextBranch.Options {
		op := d.currentTopic.currentTextBranch.Options[i]
		d.currentTopic.currentTextBranch.Options[i].button = button.NewButton(op.OptionText, nil, 0, 0)
	}

	d.currentTopic.status = topic_status_showingMainText
	d.lineWriter.Clear()
	d.lineWriter.SetSourceText(d.currentTopic.currentTextBranch.Text)

	d.currentTopic.currentTextBranch.init = true
}

// a TextBranch represents a "branch" of text in a dialog/topic.
// a text branch can simply be one body of text, but it can also have "options" which the user selects.
// These options then lead to new text branches.
// Useful when you need to have a conversation that progresses based on user answers.
type TextBranch struct {
	init       bool
	OptionText string // if this is an option for another TextBranch, this text shows for the option

	// the actual body of text that renders when this text branch is reached.
	// The minimum requirement for a text branch.
	Text string

	// the options to proceed from this text branch.
	// if there are no options for a text branch, that effectively ends the entire text branch "conversation".
	Options []TextBranch

	button *button.Button
}

func (d *Dialog) setTopic(t Topic, isReturning bool, eventBus *pubsub.EventBus) {
	if d.currentTopic != nil {
		logz.Println(d.currentTopic.TopicText, "setting new topic", t.TopicText)
	}

	if !d.init {
		panic("dialog must be initialized before setting a topic. otherwise, lineWriter won't exist.")
	}

	// if we are going from a topic to a sub-topic, set the parent topic relationship is set
	if !isReturning {
		t.parentTopic = nil
		t.parentTopic = d.currentTopic
	}

	d.currentTopic = &t

	// prepare sub-topic buttons
	buttonHeight := 35
	tileSize := d.Box.TileSize()
	for i := range d.currentTopic.SubTopics {
		buttonWidth := d.topicBoxWidth - tileSize
		subtopic := d.currentTopic.SubTopics[i]

		d.currentTopic.SubTopics[i].button = button.NewButton(subtopic.TopicText, nil, buttonWidth, buttonHeight)
	}

	if isReturning {
		t.status = topic_status_returned
	} else {
		// this is a new topic
		d.setCurrentTextBranch(d.currentTopic.MainText)
	}

	eventBus.Publish(pubsub.Event{
		Type: pubsub.Event_StartTopic,
		Data: map[string]any{
			"TopicID":    t.ID,
			"TopicTitle": t.TopicText,
		},
	})
}

func (d *Dialog) returnToParentTopic(eventBus *pubsub.EventBus) {
	if d.currentTopic == nil {
		panic("tried to return to parent topic, but current topic is nil!")
	}
	if d.currentTopic.parentTopic == nil {
		// this is the root topic; end the conversation
		d.EndDialog(eventBus)
		return
	}

	textToShow := ""
	if d.currentTopic.DoneText != "" {
		textToShow = d.currentTopic.DoneText
	} else if d.currentTopic.parentTopic.ReturnText != "" {
		textToShow = d.currentTopic.parentTopic.ReturnText
	}

	// we've already displayed the done/return text, so we are ready to change the topic now
	// or, if no final text is found, just return while leaving the current text on the screen
	if d.currentTopic.status == topic_status_goingBack || textToShow == "" {
		d.setTopic(*d.currentTopic.parentTopic, true, eventBus)
		return
	}

	// show the final text
	d.lineWriter.Clear()
	d.lineWriter.SetSourceText(textToShow)
	d.currentTopic.status = topic_status_goingBack
}
