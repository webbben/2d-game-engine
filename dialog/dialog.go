package dialog

import (
	"image/color"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/webbben/2d-game-engine/internal/display"
	"github.com/webbben/2d-game-engine/internal/image"
	"github.com/webbben/2d-game-engine/internal/rendering"
	"github.com/webbben/2d-game-engine/internal/text"
	"github.com/webbben/2d-game-engine/internal/tiled"
	"golang.org/x/image/font"
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

type Dialog struct {
	boxDef                      // definition of the tiles that build this box
	BoxTilesetSource     string // path to the tileset for the dialog box tiles
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
}

func (d *Dialog) initialize() {
	d.init = true

	// get box tiles
	d.loadBoxTiles()

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

	d.setTopic(d.RootTopic, false)
}

// call this to formally end the dialog
func (d *Dialog) EndDialog() {
	d.Exit = true
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
	d.width = display.SCREEN_WIDTH
	d.width -= d.width % d.TileWidth // round it to the size of the box tile

	if d.TopicsEnabled {
		d.topicBoxWidth = d.TileWidth * 10
		// fit the topic box into the main box width calculation
		d.width = display.SCREEN_WIDTH - d.topicBoxWidth
		d.width -= d.width % d.TileWidth
		// set height to allow space for a character portrait
		d.topicBoxHeight = display.SCREEN_HEIGHT / 4 * 3 // 3/4 of the screen height
		d.topicBoxHeight -= d.topicBoxHeight % d.TileHeight
	}

	d.height = display.SCREEN_HEIGHT / 4
	d.height -= d.height % d.TileHeight

	d.boxImage = createBoxImage(d.boxDef, d.width, d.height)
	d.topicBoxImage = createBoxImage(d.boxDef, d.topicBoxWidth, d.topicBoxHeight)
}

// creates a box using the given definition and dimensions.
// assumes width and height are multiples of the correct tile width/heights.
func createBoxImage(box boxDef, width, height int) *ebiten.Image {
	boxImage := ebiten.NewImage(width, height)

	// put together all the box tiles
	for y := 0; y < height; y += box.TileHeight {
		for x := 0; x < width; x += box.TileWidth {
			pos := ""
			if x == 0 {
				pos += "L"
			} else if x == width-box.TileWidth {
				pos += "R"
			} else {
				pos += "M"
			}
			if y == 0 {
				pos += "T"
			} else if y == height-box.TileHeight {
				pos += "B"
			} else {
				pos += "M"
			}

			switch pos {
			case "LT": // top left
				rendering.DrawImage(boxImage, box.TopLeft, float64(x), float64(y), 0)
			case "LM": // left
				rendering.DrawImage(boxImage, box.Left, float64(x), float64(y), 0)
			case "LB": // bottom left
				rendering.DrawImage(boxImage, box.BottomLeft, float64(x), float64(y), 0)
			case "RT": // top right
				rendering.DrawImage(boxImage, box.TopRight, float64(x), float64(y), 0)
			case "RM": // right
				rendering.DrawImage(boxImage, box.Right, float64(x), float64(y), 0)
			case "RB": // bottom right
				rendering.DrawImage(boxImage, box.BottomRight, float64(x), float64(y), 0)
			case "MT": // top
				rendering.DrawImage(boxImage, box.Top, float64(x), float64(y), 0)
			case "MM": // middle
				rendering.DrawImage(boxImage, box.Middle, float64(x), float64(y), 0)
			case "MB": // bottom
				rendering.DrawImage(boxImage, box.Bottom, float64(x), float64(y), 0)
			default:
				panic("buildBoxImage: invalid box tile position!")
			}
		}
	}

	return boxImage
}
