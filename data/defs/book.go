package defs

import "golang.org/x/image/font"

type BookID string

// BookDef defines a "book" in the game. Books are basically just
// a body of text that can appear somewhere in the game, like book items, letter items,
// sign posts, etc.
type BookDef struct {
	ID    BookID
	Title string
	Text  string
	Font  font.Face

	KnowledgeTopics []TopicID
}
