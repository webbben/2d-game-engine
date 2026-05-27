package object

import (
	"github.com/webbben/2d-game-engine/data/defs"
	"github.com/webbben/2d-game-engine/logz"
	"github.com/webbben/2d-game-engine/tiled"
)

type Sign struct {
	bookID defs.BookID
}

func (obj *Object) loadSignObject(allProps []tiled.Property) {
	bookID, found := tiled.GetStringProperty("book_id", allProps)
	if !found {
		logz.Panicln("Object", "sign object didnt' have a book_id property.", obj.ID)
	}
	if bookID == "" {
		logz.Panicln("loadSignObject", "book_id was empty!", obj.ID)
	}
	obj.Sign.bookID = defs.BookID(bookID)
}

func (obj *Object) activateSign() ObjectUpdateResult {
	if obj.Sign.bookID == "" {
		logz.Panicln("activateSign", "book ID was empty!", obj.ID)
	}

	return ObjectUpdateResult{
		UpdateOccurred: true,
		SignBookID:     obj.Sign.bookID,
	}
}
