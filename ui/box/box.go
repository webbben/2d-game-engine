// Package box provides a box UI component
package box

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/webbben/2d-game-engine/config"
	"github.com/webbben/2d-game-engine/imgutil/rendering"
	"github.com/webbben/2d-game-engine/logz"
	"github.com/webbben/2d-game-engine/tiled"
)

type Box struct {
	tiles []*ebiten.Image
}

func NewBox(tilesetSource string, originTileIndex int) Box {
	box := Box{}

	tileset, err := tiled.LoadTileset(tilesetSource)
	if err != nil {
		panic(err)
	}

	w := tileset.Columns
	for row := range 3 {
		for col := range 3 {
			img, err := tileset.GetTileImage(originTileIndex+col+(row*w), true)
			if err != nil {
				logz.Panicf("error loading tile image for box: %s", err)
			}
			box.tiles = append(box.tiles, img)
		}
	}

	return box
}

func (b *Box) BuildBoxImage(widthPx, heightPx int, scale float64) *ebiten.Image {
	if widthPx <= 0 || heightPx <= 0 {
		logz.Panicf("box dimensions must be positive and greater than zero. dx: %v dy: %v", widthPx, heightPx)
	}
	if scale <= 0 {
		logz.Panic("scale was <= 0")
	}

	tileSize := int(config.TileSize * scale)

	if widthPx < tileSize || heightPx < tileSize {
		logz.Panicf("given dimensions are smaller than a single tile! w: %v h: %v tileSize: %v", widthPx, heightPx, tileSize)
	}

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
				// top row
				switch col {
				case 0:
					img = b.tiles[0] // left
				case colCount - 1:
					img = b.tiles[2] // right
				default:
					img = b.tiles[1] // middle
				}
			case rowCount - 1:
				// bottom row
				switch col {
				case 0:
					img = b.tiles[6] // left
				case colCount - 1:
					img = b.tiles[8] // right
				default:
					img = b.tiles[7] // middle
				}
			default:
				// middle row
				switch col {
				case 0:
					img = b.tiles[3] // left
				case colCount - 1:
					img = b.tiles[5] // right
				default:
					img = b.tiles[4] // middle
				}
			}
			x := tileSize * col
			y := tileSize * row
			rendering.DrawImage(baseImg, img, float64(x), float64(y), scale)
			i++
		}
	}

	return baseImg
}
