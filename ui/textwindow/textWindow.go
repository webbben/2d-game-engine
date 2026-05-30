// Package textwindow provides text windows and tooltip UI components
package textwindow

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/webbben/2d-game-engine/config"
	"github.com/webbben/2d-game-engine/imgutil/rendering"
	"github.com/webbben/2d-game-engine/logz"
	"github.com/webbben/2d-game-engine/tiled"
	"github.com/webbben/2d-game-engine/ui/text"
	"github.com/webbben/2d-game-engine/utils"
	"golang.org/x/image/font"
)

// textWindowBox is a box specifically for simple text windows
type textWindowBox struct {
	tiles       []*ebiten.Image
	windowImage *ebiten.Image // the built image of the window box
}

func (twb textWindowBox) TileSize() int {
	dx := int(float64(twb.tiles[0].Bounds().Dx()) * config.UIScale)
	if dx == 0 {
		panic("tilesize is 0")
	}
	return dx
}

func newTextWindowBox(tilesetSource string, originTileIndex int) textWindowBox {
	box := textWindowBox{}

	tileset, err := tiled.LoadTileset(tilesetSource)
	if err != nil {
		panic(err)
	}

	w := tileset.Columns
	for row := range 4 {
		for col := range 3 {
			img, err := tileset.GetTileImage(originTileIndex+col+(row*w), true)
			if err != nil {
				logz.Panicf("error loading tile image for text window box: %s", err)
			}
			box.tiles = append(box.tiles, img)
		}
	}

	return box
}

func (twb *textWindowBox) buildWindowImage(widthPx, heightPx int) {
	tileSize := twb.TileSize()

	widthPx -= widthPx % tileSize
	heightPx -= heightPx % tileSize

	baseImg := ebiten.NewImage(widthPx, heightPx)

	rowCount := heightPx / tileSize
	colCount := widthPx / tileSize

	i := 0
	for row := range rowCount {
		for col := range colCount {
			var img *ebiten.Image
			switch row {
			case 0:
				// top row of header
				switch col {
				case 0:
					img = twb.tiles[0] // top left of header
				case colCount - 1:
					img = twb.tiles[2] // top right of header
				default:
					img = twb.tiles[1] // top middle of header
				}
			case 1:
				// bottom row of header
				switch col {
				case 0:
					img = twb.tiles[3] // bottom left of header
				case colCount - 1:
					img = twb.tiles[5] // bottom right of header
				default:
					img = twb.tiles[4] // bottom middle of header
				}
			case rowCount - 1:
				// very bottom row
				switch col {
				case 0:
					img = twb.tiles[9] // bottom left
				case colCount - 1:
					img = twb.tiles[11] // bottom right
				default:
					img = twb.tiles[10] // bottom middle
				}
			default:
				// body section row
				switch col {
				case 0:
					img = twb.tiles[6] // left
				case colCount - 1:
					img = twb.tiles[8] // right
				default:
					img = twb.tiles[7] // middle
				}
			}
			x := tileSize * col
			y := tileSize * row
			rendering.DrawImage(baseImg, img, float64(x), float64(y), config.UIScale)
			i++
		}
	}

	twb.windowImage = baseImg
}

// textWindow is a simple type of text window that just shows a title and body text.
type textWindow struct {
	x, y float64

	box        textWindowBox
	lineWriter text.LineWriter

	Title     string
	BodyText  string
	titleFont font.Face
}

type TextWindowParams struct {
	TilesetSource   string // the tileset that has the tiles for the window box
	OriginTileIndex int    // the index of the tile at the top-left of the window box (in Tiled)

	TitleFont font.Face // if nil, will use config default title font
	BodyFont  font.Face // if nil, will use config default body font
}

func newTextWindow(title, bodyText string, params TextWindowParams) textWindow {
	if title == "" {
		panic("title is empty")
	}
	if bodyText == "" {
		panic("bodytext is empty")
	}
	if params.TilesetSource == "" {
		panic("tilesetSource is empty")
	}
	if params.TitleFont == nil {
		params.TitleFont = config.DefaultTitleFont
	}
	if params.BodyFont == nil {
		params.BodyFont = config.DefaultFont
	}

	textWindow := textWindow{
		Title:    title,
		BodyText: bodyText,
	}

	textWindow.box = newTextWindowBox(params.TilesetSource, params.OriginTileIndex)

	tileSize := int(config.GetScaledTilesize())

	lineWidth := tileSize * 4 // default line width
	titleWidth, _, _ := text.GetStringSize(title, params.TitleFont)
	titleWidth = utils.RoundUpToTile(titleWidth, tileSize)
	lineWidth = max(lineWidth, titleWidth)

	// get full height of body text, and also confirm that no lines ended up larger than the line width.
	// that could happen if a single word in a line is longer than the lineWidth param.
	bodylines := text.ConvertStringToLines(bodyText, params.BodyFont, lineWidth)
	bodyTextHeight := 0
	longestLineDx := 0
	for _, line := range bodylines {
		dx, dy, _ := text.GetStringSize(line, params.BodyFont)
		bodyTextHeight += dy
		longestLineDx = max(longestLineDx, dx)
	}

	if longestLineDx > lineWidth {
		lineWidth = utils.RoundUpToTile(longestLineDx, tileSize)
	}

	bodyTextHeight = utils.RoundUpToTile(bodyTextHeight, tileSize)

	// title takes 2 tile height, and the border below the body gets 1 tile height
	windowHeight := tileSize*3 + bodyTextHeight
	// border tiles on both sides of text
	windowWidth := lineWidth + tileSize*2

	textWindow.box.buildWindowImage(windowWidth, windowHeight)

	lwParams := config.LineWriterParams{
		LineWidthPx:      lineWidth,
		MaxHeightPx:      bodyTextHeight + tileSize,
		FontFace:         params.BodyFont,
		UseShadow:        true,
		WriteImmediately: true,
	}
	textWindow.lineWriter = text.NewLineWriter(nil, lwParams)

	textWindow.lineWriter.SetSourceText(bodyText)

	textWindow.titleFont = params.TitleFont

	return textWindow
}

func (tw *textWindow) Draw(screen *ebiten.Image, x, y float64) {
	tw.x = x
	tw.y = y
	if tw.box.windowImage == nil {
		panic("no window image")
	}

	rendering.DrawImage(screen, tw.box.windowImage, tw.x, tw.y, 0)

	tileSize := tw.box.TileSize()
	text.DrawShadowText(screen, tw.Title, tw.titleFont, int(tw.x)+(tileSize/2), int(tw.y)+(tileSize)-5, nil, nil, 0, 0)

	tw.lineWriter.Draw(screen, int(tw.x)+(tileSize/2), int(tw.y)+tileSize+(tileSize/2))
}

func (tw *textWindow) Update() {
	tw.lineWriter.Update()
}

func (tw textWindow) Dimensions() (dx, dy int) {
	b := tw.box.windowImage.Bounds()
	return b.Dx(), b.Dy()
}
