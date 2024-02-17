package dialog

type DialogStep struct {
	Text         string
	TakeInput    bool
	InputOptions []string
}

type Dialog struct {
	Type        int // the type of dialog screen
	Steps       []DialogStep
	CurrentStep int
	SpeakerName string // name of the person that the user is interacting with
}
