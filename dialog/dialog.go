package dialog

import (
	"fmt"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/webbben/2d-game-engine/internal/config"
	"github.com/webbben/2d-game-engine/internal/image"
	"github.com/webbben/2d-game-engine/internal/model"
	"github.com/webbben/2d-game-engine/internal/rendering"
	"github.com/webbben/2d-game-engine/internal/tiled"
	"golang.org/x/image/font"
)

const (
	// box is rendered on the bottom of the screen, starting from the left side.
	BOX_POS_BOTTOM = "bottom"
)

type boxDef struct {
	Left, Right, Middle,
	TopLeft, Top, TopRight,
	BottomLeft, Bottom, BottomRight *ebiten.Image
	TileWidth, TileHeight int
}

func (bd boxDef) verifyImages() {
	if bd.Left == nil {
		panic("box left is nil")
	}
	if bd.Right == nil {
		panic("box right is nil")
	}
	if bd.Middle == nil {
		panic("box middle is nil")
	}
	if bd.TopLeft == nil {
		panic("box top left is nil")
	}
	if bd.TopRight == nil {
		panic("box top right is nil")
	}
	if bd.Top == nil {
		panic("box top is nil")
	}
	if bd.BottomLeft == nil {
		panic("box bottom left is nil")
	}
	if bd.BottomRight == nil {
		panic("box bottom right is nil")
	}
	if bd.Bottom == nil {
		panic("box bottom is nil")
	}
	x := bd.Left.Bounds().Dx() + bd.Right.Bounds().Dx() + bd.Middle.Bounds().Dx() +
		bd.TopLeft.Bounds().Dx() + bd.TopRight.Bounds().Dx() + bd.Top.Bounds().Dx() +
		bd.BottomLeft.Bounds().Dx() + bd.BottomRight.Bounds().Dx() + bd.Bottom.Bounds().Dx()
	if x/9 != bd.TileWidth {
		panic("not all box tiles have the same width!")
	}
	y := bd.Left.Bounds().Dy() + bd.Right.Bounds().Dy() + bd.Middle.Bounds().Dy() +
		bd.TopLeft.Bounds().Dy() + bd.TopRight.Bounds().Dy() + bd.Top.Bounds().Dy() +
		bd.BottomLeft.Bounds().Dy() + bd.BottomRight.Bounds().Dy() + bd.Bottom.Bounds().Dy()
	if y/9 != bd.TileHeight {
		panic("not all box tiles have the same height!")
	}
}

type Font struct {
	fontFace font.Face // font used in the dialog
	Source   string    // path to font source file
}

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

type Dialog struct {
	boxDef                  // definition of the tiles that build this box
	BoxTilesetSource string // path to the tileset for the dialog box tiles

	// Positioning; dialog boxes default to rendering on the bottom of the screen, filling the screen width.

	Width, Height int           // width and height (in tiles) to make the box. defaults to the screen width, and about 1/4 of the screen height
	BoxPosition   string        // OPT: the relative position (on the screen) where the box should be rendered. defaults to bottom.
	TilePosition  model.Coords  // OPT: specify a tile position to place this box at (top left coords). If empty, will refer to BoxPosition instead.
	x, y          float64       // the actual (absolute) position to place the built image
	boxImage      *ebiten.Image // the built dialog box image
	TextFont      Font
	init          bool // flag to indicate if this Dialog's data has been loaded and is ready to render

	// Dialog text

	RootTopic    Topic  // the root topic to start up for this dialog
	currentTopic *Topic // the currently active topic for this dialog

	lineWriter
}

type lineWriter struct {
	sourceText      string // full text the line writer is currently aiming to write
	textUpdateTimer int    // number of ticks until the next text character is added
	maxLineWidth    int    // max width (in pixels) of a line. lines will be
	lineHeight      int    // the (max) height of a single line in this set of lines

	linesToWrite       []string // source text broken down into their lines
	currentLineNumber  int      // the current line (of linesToWrite) that we are writing
	currentLineIndex   int      // the index of the current line we are writing
	writtenLines       []string // the "output" that is actually drawn
	showContinueSymbol bool     // if true, the "continue" symbol/icon is shown on the bottom right
}

func (d *Dialog) initialize() {
	d.init = true

	// get box tiles
	d.loadBoxTiles()

	// build box image
	d.buildBoxImage()
	d.lineWriter.maxLineWidth = d.boxImage.Bounds().Dx() - 200

	// set box position
	d.x = 0
	d.y = float64(config.ScreenHeight - d.boxImage.Bounds().Dy())
	fmt.Println(d.x, d.y)

	d.TextFont.fontFace = image.LoadFont(d.TextFont.Source, 24, 72)

	d.setTopic(d.RootTopic)
}

func (d *Dialog) setTopic(t Topic) {
	d.currentTopic = &t

	if d.lineWriter.maxLineWidth == 0 {
		panic("lineWriter maxLineWidth not set")
	}

	d.lineWriter.sourceText = d.currentTopic.MainText
	d.lineWriter.linesToWrite = ConvertStringToLines(d.lineWriter.sourceText, d.TextFont.fontFace, d.lineWriter.maxLineWidth)
	d.lineWriter.currentLineIndex = 0
	d.lineWriter.currentLineNumber = 0
	d.lineWriter.writtenLines = []string{""}

	// determine line height
	for _, line := range d.lineWriter.linesToWrite {
		_, lineHeight := getStringSize(line, d.TextFont.fontFace)
		if lineHeight > d.lineWriter.lineHeight {
			d.lineWriter.lineHeight = lineHeight
		}
		fmt.Println(line)
	}
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

func (d *Dialog) loadBoxTiles() {
	tileset, err := tiled.LoadTileset(d.BoxTilesetSource)
	if err != nil {
		panic("failed to load box tileset: " + err.Error())
	}
	err = tileset.GenerateTiles()
	if err != nil {
		panic("failed to generate box tileset images: " + err.Error())
	}

	for _, tile := range tileset.Tiles {
		for _, prop := range tile.Properties {
			if prop.Name == "box" {
				img, err := tileset.GetTileImage(tile.ID)
				if err != nil {
					panic("failed to load tile image: " + err.Error())
				}

				if d.TileHeight == 0 {
					d.TileHeight = img.Bounds().Dy()
				}
				if d.TileWidth == 0 {
					d.TileWidth = img.Bounds().Dx()
				}

				switch prop.GetStringValue() {
				case "TL":
					d.boxDef.TopLeft = img
				case "T":
					d.boxDef.Top = img
				case "TR":
					d.boxDef.TopRight = img
				case "L":
					d.boxDef.Left = img
				case "M":
					d.boxDef.Middle = img
				case "R":
					d.boxDef.Right = img
				case "BL":
					d.boxDef.BottomLeft = img
				case "B":
					d.boxDef.Bottom = img
				case "BR":
					d.boxDef.BottomRight = img
				default:
					panic("invalid box property value found: " + prop.GetStringValue())
				}
			}
		}
	}
}

func (d *Dialog) buildBoxImage() {
	// verify box tile images
	d.boxDef.verifyImages()
	// determine box size
	if d.Width == 0 {
		d.Width = config.ScreenWidth
		d.Width -= d.Width % d.TileWidth // round it to the size of the box tile
	}
	if d.Height == 0 {
		d.Height = config.ScreenHeight / 4
		d.Height -= d.Height % d.TileHeight
	}
	d.boxImage = ebiten.NewImage(d.Width, d.Height)

	// put together all the box tiles
	for y := 0; y < d.Height; y += d.TileHeight {
		for x := 0; x < d.Width; x += d.TileWidth {
			pos := ""
			if x == 0 {
				pos += "L"
			} else if x == d.Width-d.TileWidth {
				pos += "R"
			} else {
				pos += "M"
			}
			if y == 0 {
				pos += "T"
			} else if y == d.Height-d.TileHeight {
				pos += "B"
			} else {
				pos += "M"
			}

			switch pos {
			case "LT": // top left
				rendering.DrawImage(d.boxImage, d.boxDef.TopLeft, float64(x), float64(y), 0)
			case "LM": // left
				rendering.DrawImage(d.boxImage, d.boxDef.Left, float64(x), float64(y), 0)
			case "LB": // bottom left
				rendering.DrawImage(d.boxImage, d.boxDef.BottomLeft, float64(x), float64(y), 0)
			case "RT": // top right
				rendering.DrawImage(d.boxImage, d.boxDef.TopRight, float64(x), float64(y), 0)
			case "RM": // right
				rendering.DrawImage(d.boxImage, d.boxDef.Right, float64(x), float64(y), 0)
			case "RB": // bottom right
				rendering.DrawImage(d.boxImage, d.boxDef.BottomRight, float64(x), float64(y), 0)
			case "MT": // top
				rendering.DrawImage(d.boxImage, d.boxDef.Top, float64(x), float64(y), 0)
			case "MM": // middle
				rendering.DrawImage(d.boxImage, d.boxDef.Middle, float64(x), float64(y), 0)
			case "MB": // bottom
				rendering.DrawImage(d.boxImage, d.boxDef.Bottom, float64(x), float64(y), 0)
			default:
				panic("buildBoxImage: invalid box tile position!")
			}
		}
	}
}
