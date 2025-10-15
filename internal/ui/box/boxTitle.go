package box

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/webbben/2d-game-engine/internal/config"
	"github.com/webbben/2d-game-engine/internal/logz"
	"github.com/webbben/2d-game-engine/internal/rendering"
	"github.com/webbben/2d-game-engine/internal/text"
	"github.com/webbben/2d-game-engine/internal/tiled"
	"golang.org/x/image/font"
)

// A Box Title is a part of a box that can hold a title. it's designed to blend into the top border of a box.
type BoxTitle struct {
	builtImage *ebiten.Image

	title string
	f     font.Face

	tilesetSource   string
	originTileIndex int
}

func (bt BoxTitle) Width() int {
	return bt.builtImage.Bounds().Dx()
}

func NewBoxTitle(tilesetSrc string, originIndex int, titleString string, f font.Face) BoxTitle {
	tileset, err := tiled.LoadTileset(tilesetSrc)
	if err != nil {
		logz.Panicf("failed to load tileset: %s", err)
	}
	// box titles have 10 source tiles: 5 for the top row, 5 for the bottom.
	// the far left and right sides are the "transitions", the sides next to those are the edges of the title box.
	// The middle top and bottom rows are for
	// making the actual box that can extend as long as the title is. a title box has a set height, and no middle rows.

	if f == nil {
		f = config.DefaultTitleFont
	}

	tileSize := int(config.TileSize * config.UIScale)
	// 2 edges, the title width
	titleWidth, _, _ := text.GetStringSize(titleString, f)
	titleTileWidth := (titleWidth + (tileSize * 3 / 2)) / tileSize
	width := tileSize * (2 + titleTileWidth)
	height := tileSize * 2

	baseImg := ebiten.NewImage(width, height)

	// get source tiles
	tiles := []*ebiten.Image{}
	for row := range 2 {
		for col := range 5 {
			tileImg, err := tileset.GetTileImage(originIndex + col + (row * tileset.Columns))
			if err != nil {
				logz.Panicf("failed to load box title tile image: %s", err)
			}
			tiles = append(tiles, tileImg)
		}
	}

	// build image
	numCols := width / tileSize
	for row := range 2 {
		for col := range numCols {
			var img *ebiten.Image
			switch col {
			case 0:
				// left transition
				img = tiles[row*5]
			case 1:
				// left border
				img = tiles[1+(row*5)]
			case numCols - 2:
				// right border
				img = tiles[3+(row*5)]
			case numCols - 1:
				// right transition
				img = tiles[4+(row*5)]
			default:
				// middle title area
				img = tiles[2+(row*5)]
			}
			rendering.DrawImage(baseImg, img, float64(col*tileSize), float64(row*tileSize), config.UIScale)
		}
	}

	return BoxTitle{
		builtImage:      baseImg,
		tilesetSource:   tilesetSrc,
		originTileIndex: originIndex,
		title:           titleString,
		f:               f,
	}
}

func (bt *BoxTitle) SetTitle(title string) {
	bt.title = title
}

func (bt BoxTitle) GetTitle() string {
	return bt.title
}

func (bt BoxTitle) Draw(screen *ebiten.Image, x, y float64) {
	tileSize := int(config.TileSize * config.UIScale)

	rendering.DrawImage(screen, bt.builtImage, x, y, 0)
	titleWidth, _, _ := text.GetStringSize(bt.title, bt.f)
	titleX := int(x) + (bt.Width() / 2) - (titleWidth / 2)
	titleY := int(y) + tileSize
	text.DrawShadowText(screen, bt.title, bt.f, titleX, titleY, nil, nil, 0, 0)
}
