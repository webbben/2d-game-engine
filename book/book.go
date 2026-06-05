// Package book defines logic for viewing BookDefs
package book

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/webbben/2d-game-engine/audio"
	"github.com/webbben/2d-game-engine/config"
	"github.com/webbben/2d-game-engine/data/datamanager"
	"github.com/webbben/2d-game-engine/data/defs"
	"github.com/webbben/2d-game-engine/dialogv2"
	"github.com/webbben/2d-game-engine/display"
	characterstate "github.com/webbben/2d-game-engine/entity/characterState"
	"github.com/webbben/2d-game-engine/imgutil/rendering"
	"github.com/webbben/2d-game-engine/logz"
	"github.com/webbben/2d-game-engine/pubsub"
	"github.com/webbben/2d-game-engine/ui/box"
	"github.com/webbben/2d-game-engine/ui/button"
	"github.com/webbben/2d-game-engine/ui/scrollarea"
	"github.com/webbben/2d-game-engine/ui/text"
	"github.com/webbben/2d-game-engine/utils"
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

	audioman *audio.AudioManager

	b box.Box

	// this image represents the current page's text, in its entirety
	pageImg *ebiten.Image

	scrollMode bool
	scrollArea scrollarea.ScrollArea

	// the image of the "book" - the borders and background image of paper or whatever the book is made of
	boxImage *ebiten.Image

	lw text.LineWriter

	titleFont font.Face

	closeBtn *button.Button
}

func (bs BookSession) GetBookID() defs.BookID {
	return bs.bookID
}

func (bs BookSession) Dimensions() (dx, dy int) {
	return bs.boxImage.Bounds().Dx(), bs.boxImage.Bounds().Dy()
}

// NewBookSession creates a book session.
//
// If linewriter max height is set, then the page will be made to be that tall, and paging will be in effect.
// If linewriter max height is unset (0), then we will make the book scrollable, so that instead of paging, you use the mouse wheel to scroll and see all the text.
func NewBookSession(
	bookID defs.BookID,
	dataman *datamanager.DataManager,
	audioman *audio.AudioManager,
	eventBus *pubsub.EventBus,
	playerInfo defs.PlayerInfo,
	params config.BookSessionParams,
) *BookSession {
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

	if params.LineWriterParams.LineWidthPx <= 0 {
		logz.Panic("lw params line width px was <= 0")
	}

	tilesize := config.GetScaledTilesize()

	// ensure that these are valid tilesizes
	params.LineWriterParams.LineWidthPx = utils.RoundUpToTile(params.LineWriterParams.LineWidthPx, int(tilesize))
	params.LineWriterParams.MaxHeightPx = utils.RoundUpToTile(params.LineWriterParams.MaxHeightPx, int(tilesize))

	sesh := &BookSession{
		audioman:  audioman,
		bookID:    bookID,
		bookDef:   dataman.GetBookDef(bookID),
		b:         box.NewBox(params.BoxTileset, params.BoxOrigin),
		titleFont: params.TitleFont,
	}

	if sesh.bookDef.Text == "" && sesh.bookDef.Title == "" {
		logz.Panicln("NewBookSession", "bookDef didn't have any text or title")
	}

	if sesh.bookDef.Text != "" {
		sesh.bookDef.Text = dialogv2.InsertDialogVariables(sesh.bookDef.Text, playerInfo, dataman)
	}

	// decide which font to use
	if sesh.bookDef.Font != nil {
		// use this font instead of whatever is in the linewriter params
		params.LineWriterParams.FontFace = sesh.bookDef.Font
	} else if params.LineWriterParams.FontFace == nil {
		params.LineWriterParams.FontFace = config.DefaultFont
	}

	if params.LineWriterParams.MaxHeightPx == 0 {
		params.LineWriterParams.MaxHeightPx = display.SCREEN_HEIGHT * 3
		sesh.scrollMode = true // no explicit height was set, so we will do scroll mode
	}

	// setup linewriter
	sesh.lw = text.NewLineWriter(audioman, params.LineWriterParams)
	if sesh.bookDef.Text != "" {
		sesh.lw.SetSourceText(sesh.bookDef.Text)
	}
	sesh.lw.Update()

	// create page image
	sesh.renderPageImage()

	// make box image bigger than page image
	w := sesh.pageImg.Bounds().Dx()
	h := sesh.pageImg.Bounds().Dy()

	// enforce min and max on height, since if we are in scroll mode the page img could be quite big or somewhat small
	h = min(h, display.SCREEN_HEIGHT*2/3)
	h = max(h, int(tilesize)*6)

	textAreaHeight := h
	textAreaWidth := w

	w += int(tilesize * 2)
	h += int(tilesize * 2) // height needs to include space for button
	w = utils.RoundUpToTile(w, int(tilesize))
	h = utils.RoundUpToTile(h, int(tilesize))

	sesh.boxImage = sesh.b.BuildBoxImage(w, h, config.UIScale)

	if sesh.scrollMode {
		sesh.scrollArea = scrollarea.NewScrollArea(scrollarea.ScrollAreaParams{
			Width:  textAreaWidth,
			Height: textAreaHeight,
		})
	}

	sesh.closeBtn = button.NewButton("Close", config.DefaultFont, 0, 0, audioman)

	// if the book def has knowledge topics, apply it to the player's character state
	if len(sesh.bookDef.KnowledgeTopics) > 0 {
		for _, topic := range sesh.bookDef.KnowledgeTopics {
			characterstate.AddKnowledge(topic, dataman, eventBus)
		}
	}

	if sesh.bookDef.OpenSFX != "" {
		audioman.PlaySFX(sesh.bookDef.OpenSFX, 0.5)
	}

	return sesh
}

func (bs BookSession) IsDone() bool {
	return bs.exit
}

func (bs *BookSession) Update() {
	if bs.scrollMode {
		bs.scrollArea.Update()
	} else {
		if bs.bookDef.Text != "" {
			bs.lw.Update()
		}
	}

	if bs.closeBtn.Update().Clicked {
		bs.exit = true
		if bs.bookDef.CloseSFX != "" {
			bs.audioman.PlaySFX(bs.bookDef.CloseSFX, 0.5)
		}
	}
}

func (bs *BookSession) renderPageImage() {
	tilesize := int(config.GetScaledTilesize())

	if bs.pageImg == nil {
		totalHeight := tilesize * 2 // top and bottom margin

		if bs.bookDef.Title != "" {
			titleDy, desc := text.GetRealisticFontMetrics(bs.titleFont)
			totalHeight += titleDy + desc
			totalHeight += tilesize
		}

		dx, dy := bs.lw.CurrentDimensions()
		totalHeight += dy + tilesize
		width := dx

		bs.pageImg = ebiten.NewImage(width, totalHeight)
	} else {
		bs.pageImg.Clear()
	}

	drawX := 0
	drawY := tilesize

	if bs.bookDef.Title != "" {
		titleHeight, desc := text.GetRealisticFontMetrics(bs.titleFont)
		drawY += titleHeight + desc
		middleX := bs.pageImg.Bounds().Dx() / 2
		drawX = int(text.CenterTextOnXPos(bs.bookDef.Title, bs.titleFont, float64(middleX)))
		text.DrawShadowText(bs.pageImg, bs.bookDef.Title, bs.titleFont, drawX, drawY, nil, nil, 0, 0)
		drawY += tilesize
	}

	if bs.bookDef.Text != "" {
		drawX = 0
		bs.lw.Draw(bs.pageImg, drawX, drawY)
	}
}

func (bs *BookSession) Draw(screen *ebiten.Image, x, y float64) {
	if bs.pageImg == nil {
		logz.Panic("page img is nil!")
	}

	tilesize := config.GetScaledTilesize()
	drawX := x
	drawY := y
	rendering.DrawImage(screen, bs.boxImage, drawX, drawY, 0)

	drawY += tilesize / 2
	drawX += tilesize

	if bs.scrollMode {
		bs.scrollArea.Draw(screen, bs.pageImg, drawX, drawY)
	} else {
		rendering.DrawImage(screen, bs.pageImg, drawX, drawY, 0)
	}

	btnDx, btnDy := float64(bs.closeBtn.Width), float64(bs.closeBtn.Height)
	drawX = x + float64(bs.boxImage.Bounds().Dx()) - btnDx - tilesize
	drawY = y + float64(bs.boxImage.Bounds().Dy()) - btnDy - tilesize
	bs.closeBtn.Draw(screen, int(drawX), int(drawY))
}
