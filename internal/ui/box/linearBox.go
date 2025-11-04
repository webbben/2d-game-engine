package box

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/webbben/2d-game-engine/internal/config"
	"github.com/webbben/2d-game-engine/internal/rendering"
	"github.com/webbben/2d-game-engine/internal/tiled"
)

type LinearBox struct {
	tiles     []*ebiten.Image
	tileWidth int
	img       *ebiten.Image
}

func (lb LinearBox) Image() *ebiten.Image {
	return lb.img
}

type LinearBoxParams struct {
	TilesetSrc  string
	OriginIndex int
	TileWidth   int // width in tiles
}

func NewLinearBox(params LinearBoxParams) LinearBox {
	if params.TileWidth < 2 {
		panic("width must be at least 2 tiles")
	}
	b := LinearBox{}
	b.tiles = append(b.tiles, tiled.GetTileImage(params.TilesetSrc, params.OriginIndex))
	b.tiles = append(b.tiles, tiled.GetTileImage(params.TilesetSrc, params.OriginIndex+1))
	b.tiles = append(b.tiles, tiled.GetTileImage(params.TilesetSrc, params.OriginIndex+2))

	tileSize := int(config.TileSize * config.UIScale)

	// draw bar image
	b.img = ebiten.NewImage(params.TileWidth*tileSize, tileSize)
	for i := range params.TileWidth {
		var img *ebiten.Image
		switch i {
		case 0:
			img = b.tiles[0]
		case params.TileWidth - 1:
			img = b.tiles[2]
		default:
			img = b.tiles[1]
		}
		rendering.DrawImage(b.img, img, float64(i*tileSize), 0, config.UIScale)
	}

	return b
}
