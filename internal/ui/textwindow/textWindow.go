// Package textwindow provides text windows and tooltip UI components
package textwindow

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/webbben/2d-game-engine/internal/config"
	"github.com/webbben/2d-game-engine/internal/logz"
	"github.com/webbben/2d-game-engine/internal/rendering"
	"github.com/webbben/2d-game-engine/internal/text"
	"github.com/webbben/2d-game-engine/internal/tiled"
	"golang.org/x/image/font"
)

type TextWindowBox struct {
	tiles       []*ebiten.Image
	windowImage *ebiten.Image
}

func (twb TextWindowBox) TileSize() int {
	dx := int(float64(twb.tiles[0].Bounds().Dx()) * config.UIScale)
	if dx == 0 {
		panic("tilesize is 0")
	}
	return dx
}

func NewTextWindowBox(tilesetSource string, originTileIndex int) TextWindowBox {
	box := TextWindowBox{}

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

func (twb *TextWindowBox) BuildWindowImage(widthPx, heightPx int) {
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

type TextWindow struct {
	x, y float64

	box        TextWindowBox
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

	// if set (not 0), will try to split lines based on this value in pixels.
	// if not set (= 0), a suitable width will be calculated
	LineWidthPx int
}

func NewTextWindow(title, bodyText string, params TextWindowParams) TextWindow {
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

	textWindow := TextWindow{
		Title:    title,
		BodyText: bodyText,
	}

	textWindow.box = NewTextWindowBox(params.TilesetSource, params.OriginTileIndex)

	// determine window size based on source text
	// title = 3 tiles tall
	tileSize := textWindow.box.TileSize()
	minWidth := tileSize * 5
	lineWidth := tileSize * 3
	windowHeight := tileSize * 2

	// determine window width by text options
	windowWidth := params.LineWidthPx + (tileSize * 2)
	if windowWidth <= minWidth {
		windowWidth = minWidth
		titleWidth, _, _ := text.GetStringSize(title, params.TitleFont)
		titleWidth += tileSize
		// ensure width is big enough to fit title at least
		if windowWidth < titleWidth {
			windowWidth = titleWidth + tileSize
			windowWidth -= windowWidth % tileSize
		}

		lineWidth = windowWidth - (tileSize * 2)
		lineWidth -= lineWidth % tileSize
	}

	bodyTextHeight := text.GetStringLinesHeight(bodyText, params.BodyFont, lineWidth)

	windowHeight += bodyTextHeight + (tileSize)
	windowHeight -= windowHeight % tileSize

	textWindow.box.BuildWindowImage(windowWidth, windowHeight)

	textWindow.lineWriter = text.NewLineWriter(lineWidth, bodyTextHeight+tileSize, params.BodyFont, nil, nil, true, true)

	textWindow.lineWriter.SetSourceText(bodyText)

	textWindow.titleFont = params.TitleFont

	return textWindow
}

func (tw *TextWindow) Draw(screen *ebiten.Image, x, y float64) {
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

func (tw *TextWindow) Update() {
	tw.lineWriter.Update()
}

func (tw TextWindow) Dimensions() (dx, dy int) {
	b := tw.box.windowImage.Bounds()
	return b.Dx(), b.Dy()
}
