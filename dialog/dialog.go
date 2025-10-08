package dialog

import (
	"image/color"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/webbben/2d-game-engine/internal/config"
	"github.com/webbben/2d-game-engine/internal/display"
	"github.com/webbben/2d-game-engine/internal/image"
	"github.com/webbben/2d-game-engine/internal/pubsub"
	"github.com/webbben/2d-game-engine/internal/text"
	"github.com/webbben/2d-game-engine/internal/ui"
	"golang.org/x/image/font"
)

type Font struct {
	fontFace font.Face // font used in the dialog
	Source   string    // path to font source file
}

type Dialog struct {
	ID                   string
	NPCID                string
	EntID                string
	ui.BoxDef                   // definition of the tiles that build this box
	BoxTilesetSource     string // path to the tileset for the dialog box tiles
	BoxOriginTileIndex   int    // index of top left tile for this box in the tileset
	TextFont             Font
	init                 bool // flag to indicate if this Dialog's data has been loaded and is ready to render
	Exit                 bool // flag to indicate dialog has exited. cuts off dialog updates and draws.
	SuppressGoodbyeTopic bool // if set, the "Goodbye" exit sub-topic won't be shown for the root topic

	// main text box

	boxImage      *ebiten.Image // the built dialog box image
	width, height int           // width and height of the main dialog box
	x, y          float64       // the actual (absolute) position to place the built image

	// topic box

	TopicsEnabled                 bool          // if true, a topic box will show and topic selection will be enabled
	topicBoxImage                 *ebiten.Image // the built topic box image
	topicBoxWidth, topicBoxHeight int
	topicBoxX, topicBoxY          float64

	// Dialog topic and text

	RootTopic    Topic  // the root topic to start up for this dialog
	currentTopic *Topic // the currently active topic for this dialog

	lineWriter        text.LineWriter
	flashContinueIcon bool // flash icon to indicate the dialog will continue to the next page of text
	flashDoneIcon     bool // flash icon to indicate that current dialog text has finished
	iconFlashTimer    int  // timer for flashing icon, based on update ticks

	ticksSinceLastClick int
}

func (d *Dialog) initialize(eventBus *pubsub.EventBus) {
	d.init = true

	eventBus.Publish(pubsub.Event{
		Type: pubsub.Event_StartDialog,
		Data: map[string]any{
			"NPCID":    d.NPCID,
			"EntID":    d.EntID,
			"DialogID": d.ID,
		},
	})

	// get box tiles
	d.BoxDef = ui.NewBox(d.BoxTilesetSource, d.BoxOriginTileIndex)

	// build box image
	d.buildBoxImage()
	maxLineWidth := d.boxImage.Bounds().Dx() - 50
	maxHeight := d.boxImage.Bounds().Dy() - 50

	// setup lineWriter
	d.TextFont.fontFace = image.LoadFont(d.TextFont.Source, 24, 72)
	d.lineWriter = text.NewLineWriter(maxLineWidth, maxHeight, d.TextFont.fontFace, color.Black, color.RGBA{20, 20, 20, 75}, true)

	// set box position
	d.x = 0
	d.y = float64(display.SCREEN_HEIGHT - d.boxImage.Bounds().Dy())
	if d.TopicsEnabled {
		d.topicBoxY = float64(display.SCREEN_HEIGHT - d.topicBoxImage.Bounds().Dy())
		d.topicBoxX = d.x + float64(d.boxImage.Bounds().Dx())
	}
	// there might be a gap at the end; let's try to center these boxes a little
	endX := d.x + float64(d.boxImage.Bounds().Dx())
	if d.TopicsEnabled {
		endX += float64(d.topicBoxImage.Bounds().Dx())
	}
	pushX := (float64(display.SCREEN_WIDTH) - endX) / 2
	d.x += pushX
	if d.TopicsEnabled {
		d.topicBoxX += pushX
	}

	if !d.SuppressGoodbyeTopic {
		d.RootTopic.SubTopics = append([]Topic{goodbyeTopic}, d.RootTopic.SubTopics...)
	}

	d.setTopic(d.RootTopic, false, eventBus)
}

// call this to formally end the dialog
func (d *Dialog) EndDialog(eventBus *pubsub.EventBus) {
	if d.Exit {
		panic("tried to end dialog when Exit is already set to true. are multiple places trying to simultaneously end dialog?")
	}
	d.Exit = true
	eventBus.Publish(pubsub.Event{
		Type: pubsub.Event_EndDialog,
		Data: map[string]any{
			"NPCID":    d.NPCID,
			"EntID":    d.EntID,
			"DialogID": d.ID,
		},
	})
	// TODO handle any cleanup here?
}

func getStringSize(s string, f font.Face) (dx int, dy int) {
	bounds, _ := font.BoundString(f, s)
	return bounds.Max.X.Ceil() - bounds.Min.X.Floor(), bounds.Max.Y.Ceil() - bounds.Min.Y.Floor()
}

func ConvertStringToLines(s string, f font.Face, lineWidthPx int) []string {
	if lineWidthPx == 0 {
		panic("ConvertStringtoLines: lineWidthPx is 0!")
	}
	// get all the words (space separated) and find out how many words can be fit in a line
	words := strings.Fields(s)
	var currentLine strings.Builder
	currentLineWidth := 0
	lines := []string{}

	for i, w := range words {
		if i < len(words)-1 {
			w += " " // each word has a space
		}
		wordDx, _ := getStringSize(w, f)
		if currentLineWidth+wordDx > lineWidthPx {
			lines = append(lines, currentLine.String())
			currentLine.Reset()
			currentLineWidth = 0
		}
		currentLineWidth += wordDx
		currentLine.WriteString(w)
	}

	lines = append(lines, currentLine.String())

	return lines
}

func (d *Dialog) buildBoxImage() {
	// determine box size
	tileSize := d.BoxDef.TileSize()
	d.width = display.SCREEN_WIDTH
	d.width -= d.width % tileSize // round it to the size of the box tile

	if d.TopicsEnabled {
		// using config tilesize instead of box tilesize since box tilesize is scaled by UI scale
		d.topicBoxWidth = config.TileSize * 17
		d.topicBoxWidth -= d.topicBoxWidth % tileSize
		// fit the topic box into the main box width calculation
		d.width = display.SCREEN_WIDTH - d.topicBoxWidth
		d.width -= d.width % tileSize
		// set height to allow space for a character portrait
		d.topicBoxHeight = display.SCREEN_HEIGHT / 4 * 3 // 3/4 of the screen height
		d.topicBoxHeight -= d.topicBoxHeight % tileSize
	}

	d.height = display.SCREEN_HEIGHT / 4
	d.height -= d.height % tileSize

	d.boxImage = d.BoxDef.BuildBoxImage(d.width, d.height)
	d.topicBoxImage = d.BoxDef.BuildBoxImage(d.topicBoxWidth, d.topicBoxHeight)
}
