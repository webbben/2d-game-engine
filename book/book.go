// Package book defines logic for viewing BookDefs
package book

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/webbben/2d-game-engine/audio"
	"github.com/webbben/2d-game-engine/config"
	"github.com/webbben/2d-game-engine/data/datamanager"
	"github.com/webbben/2d-game-engine/data/defs"
	"github.com/webbben/2d-game-engine/imgutil/rendering"
	"github.com/webbben/2d-game-engine/logz"
	"github.com/webbben/2d-game-engine/ui/box"
	"github.com/webbben/2d-game-engine/ui/button"
	"github.com/webbben/2d-game-engine/ui/text"
	"golang.org/x/image/font"
)

// A BookSession is used for viewing a book (BookDef) in the game.
// It provides a UI and handles runtime mechanics such as turning pages.
//
// For now, I'm just implementing a simple note view:
//   - box with text inside of it
//
// In the future, I'll probably make a pageable book or something like that.
type BookSession struct {
	exit    bool
	bookID  defs.BookID
	bookDef defs.BookDef

	b box.Box

	boxImage *ebiten.Image

	lw text.LineWriter

	titleFont font.Face

	closeBtn *button.Button
}

func (bs BookSession) Dimensions() (dx, dy int) {
	return bs.boxImage.Bounds().Dx(), bs.boxImage.Bounds().Dy()
}

func NewBookSession(bookID defs.BookID, dataman *datamanager.DataManager, audioman *audio.AudioManager, params config.BookSessionParams) *BookSession {
	if dataman == nil {
		panic("dataman was nil")
	}
	if audioman == nil {
		panic("audioman was nil")
	}
	if bookID == "" {
		panic("book ID was empty")
	}
	if params.BoxTileset == "" {
		panic("box tileset was empty")
	}
	if params.TitleFont == nil {
		params.TitleFont = config.DefaultTitleFont
	}

	sesh := &BookSession{
		bookID:    bookID,
		bookDef:   dataman.GetBookDef(bookID),
		b:         box.NewBox(params.BoxTileset, params.BoxOrigin),
		titleFont: params.TitleFont,
	}

	if sesh.bookDef.Text == "" && sesh.bookDef.Title == "" {
		logz.Panicln("NewBookSession", "bookDef didn't have any text or title")
	}

	// decide which font to use
	if sesh.bookDef.Font != nil {
		// use this font instead of whatever is in the linewriter params
		params.LineWriterParams.FontFace = sesh.bookDef.Font
	} else if params.LineWriterParams.FontFace == nil {
		params.LineWriterParams.FontFace = config.DefaultFont
	}

	tilesize := config.GetScaledTilesize()

	w := params.LineWriterParams.LineWidthPx + int(tilesize)
	h := params.LineWriterParams.MaxHeightPx + int(tilesize)

	sesh.boxImage = sesh.b.BuildBoxImage(w, h, config.UIScale)

	if sesh.bookDef.Text != "" {
		sesh.lw = text.NewLineWriter(audioman, params.LineWriterParams)
		sesh.lw.SetSourceText(sesh.bookDef.Text)
	}

	sesh.closeBtn = button.NewButton("Close", config.DefaultFont, 0, 0, audioman)

	return sesh
}

func (bs BookSession) IsDone() bool {
	return bs.exit
}

func (bs *BookSession) Update() {
	if bs.bookDef.Text != "" {
		bs.lw.Update()
	}

	if bs.closeBtn.Update().Clicked {
		bs.exit = true
	}
}

func (bs *BookSession) Draw(screen *ebiten.Image, x, y float64) {
	tilesize := config.GetScaledTilesize()
	drawX := x
	drawY := y
	rendering.DrawImage(screen, bs.boxImage, drawX, drawY, 0)

	drawY += tilesize
	drawX += tilesize

	if bs.bookDef.Title != "" {
		middleX := x + float64(bs.boxImage.Bounds().Dx()/2)
		titleX := text.CenterTextOnXPos(bs.bookDef.Title, bs.titleFont, middleX)

		text.DrawShadowText(screen, bs.bookDef.Title, bs.titleFont, int(titleX), int(drawY), nil, nil, 0, 0)
		drawY += tilesize * 2
	}

	if bs.bookDef.Text != "" {
		bs.lw.Draw(screen, int(drawX), int(drawY))
	}

	btnDx, btnDy := float64(bs.closeBtn.Width), float64(bs.closeBtn.Height)
	drawX = x + float64(bs.boxImage.Bounds().Dx()) - btnDx - tilesize
	drawY = y + float64(bs.boxImage.Bounds().Dy()) - btnDy - tilesize
	bs.closeBtn.Draw(screen, int(drawX), int(drawY))
}
