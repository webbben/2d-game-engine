package box

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/webbben/2d-game-engine/internal/config"
	"github.com/webbben/2d-game-engine/internal/logz"
	"github.com/webbben/2d-game-engine/internal/rendering"
	"github.com/webbben/2d-game-engine/internal/tiled"
)

type Box struct {
	tiles    []*ebiten.Image
	Unscaled bool // if set to true, this box will not use the global UI scaling
}

// the real size of the tiles used in this box. applies any global scaling.
func (b Box) TileSize() int {
	if len(b.tiles) == 0 {
		panic("called tilesize before tiles were created")
	}

	dx := b.tiles[0].Bounds().Dx()

	if b.Unscaled {
		return dx
	}
	return int(float64(dx) * config.UIScale)
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
			img, err := tileset.GetTileImage(originTileIndex + col + (row * w))
			if err != nil {
				logz.Panicf("error loading tile image for box: %s", err)
			}
			box.tiles = append(box.tiles, img)
		}
	}

	return box
}

func (b *Box) BuildBoxImage(widthPx, heightPx int) *ebiten.Image {
	if widthPx <= 0 || heightPx <= 0 {
		logz.Panicf("box dimensions must be positive and greater than zero. dx: %v dy: %v", widthPx, heightPx)
	}
	tileSize := b.TileSize()

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
			rendering.DrawImage(baseImg, img, float64(x), float64(y), config.UIScale)
			i++
		}
	}

	return baseImg
}
