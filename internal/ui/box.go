package ui

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/webbben/2d-game-engine/internal/config"
	"github.com/webbben/2d-game-engine/internal/rendering"
	"github.com/webbben/2d-game-engine/internal/tiled"
)

type BoxDef struct {
	Left, Right, Middle,
	TopLeft, Top, TopRight,
	BottomLeft, Bottom, BottomRight *ebiten.Image
	tileWidth, tileHeight int  // actual pixel size of the source tile images
	Unscaled              bool // if set to true, this box will not use the global UI scaling
}

// the real size of the tiles used in this box. applies any global scaling.
func (bd BoxDef) TileDimensions() (dx, dy int) {
	if bd.tileWidth == 0 || bd.tileHeight == 0 {
		panic("TileDimensions called on box when tileWidth or tileHeight not yet calculated")
	}
	if bd.Unscaled {
		return bd.tileWidth, bd.tileHeight
	}
	return int(float64(bd.tileWidth) * config.UIScale), int(float64(bd.tileHeight) * config.UIScale)
}

func (bd BoxDef) VerifyImages() {
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
	if x/9 != bd.tileWidth {
		panic("not all box tiles have the same width!")
	}
	y := bd.Left.Bounds().Dy() + bd.Right.Bounds().Dy() + bd.Middle.Bounds().Dy() +
		bd.TopLeft.Bounds().Dy() + bd.TopRight.Bounds().Dy() + bd.Top.Bounds().Dy() +
		bd.BottomLeft.Bounds().Dy() + bd.BottomRight.Bounds().Dy() + bd.Bottom.Bounds().Dy()
	if y/9 != bd.tileHeight {
		panic("not all box tiles have the same height!")
	}
}

func (bd *BoxDef) LoadBoxTiles(boxTilesetSource string, boxID string) {
	if boxID == "" {
		panic("no box tileset id given")
	}
	if boxTilesetSource == "" {
		panic("no box tileset source given")
	}
	tileset, err := tiled.LoadTileset(boxTilesetSource)
	if err != nil {
		panic("failed to load box tileset: " + err.Error())
	}

	for _, tile := range tileset.Tiles {
		id, found := tiled.GetStringProperty("box_id", tile.Properties)
		if !found || id != boxID {
			continue
		}
		for _, prop := range tile.Properties {
			if prop.Name == "box" {
				img, err := tileset.GetTileImage(tile.ID)
				if err != nil {
					panic("failed to load tile image: " + err.Error())
				}

				if bd.tileHeight == 0 {
					bd.tileHeight = img.Bounds().Dy()
				}
				if bd.tileWidth == 0 {
					bd.tileWidth = img.Bounds().Dx()
				}

				switch prop.GetStringValue() {
				case "TL":
					bd.TopLeft = img
				case "T":
					bd.Top = img
				case "TR":
					bd.TopRight = img
				case "L":
					bd.Left = img
				case "M":
					bd.Middle = img
				case "R":
					bd.Right = img
				case "BL":
					bd.BottomLeft = img
				case "B":
					bd.Bottom = img
				case "BR":
					bd.BottomRight = img
				default:
					panic("invalid box property value found: " + prop.GetStringValue())
				}
			}
		}
	}
}

// creates a box using the given definition and dimensions.
// clamps down the given dimensions to ensure they are a multiple of the actual tilesize of this box's tiles.
func (bd BoxDef) CreateBoxImage(width, height int) *ebiten.Image {
	// clamp down to multiple of the actual tile size
	tileWidth, tileHeight := bd.TileDimensions()
	height -= height % tileHeight
	width -= width % tileWidth

	boxImage := ebiten.NewImage(width, height)

	scale := config.UIScale
	if bd.Unscaled {
		scale = 1
	}

	// put together all the box tiles
	for y := 0; y < height; y += tileHeight {
		for x := 0; x < width; x += tileWidth {
			pos := ""
			if x == 0 {
				pos += "L"
			} else if x == width-tileWidth {
				pos += "R"
			} else {
				pos += "M"
			}
			if y == 0 {
				pos += "T"
			} else if y == height-tileHeight {
				pos += "B"
			} else {
				pos += "M"
			}

			switch pos {
			case "LT": // top left
				rendering.DrawImage(boxImage, bd.TopLeft, float64(x), float64(y), scale)
			case "LM": // left
				rendering.DrawImage(boxImage, bd.Left, float64(x), float64(y), scale)
			case "LB": // bottom left
				rendering.DrawImage(boxImage, bd.BottomLeft, float64(x), float64(y), scale)
			case "RT": // top right
				rendering.DrawImage(boxImage, bd.TopRight, float64(x), float64(y), scale)
			case "RM": // right
				rendering.DrawImage(boxImage, bd.Right, float64(x), float64(y), scale)
			case "RB": // bottom right
				rendering.DrawImage(boxImage, bd.BottomRight, float64(x), float64(y), scale)
			case "MT": // top
				rendering.DrawImage(boxImage, bd.Top, float64(x), float64(y), scale)
			case "MM": // middle
				rendering.DrawImage(boxImage, bd.Middle, float64(x), float64(y), scale)
			case "MB": // bottom
				rendering.DrawImage(boxImage, bd.Bottom, float64(x), float64(y), scale)
			default:
				panic("CreateBoxImage: invalid box tile position!")
			}
		}
	}

	return boxImage
}
