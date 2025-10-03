package ui

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/webbben/2d-game-engine/internal/rendering"
	"github.com/webbben/2d-game-engine/internal/tiled"
)

type BoxDef struct {
	Left, Right, Middle,
	TopLeft, Top, TopRight,
	BottomLeft, Bottom, BottomRight *ebiten.Image
	TileWidth, TileHeight int
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
	err = tileset.GenerateTiles()
	if err != nil {
		panic("failed to generate box tileset images: " + err.Error())
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

				if bd.TileHeight == 0 {
					bd.TileHeight = img.Bounds().Dy()
				}
				if bd.TileWidth == 0 {
					bd.TileWidth = img.Bounds().Dx()
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
// assumes width and height are multiples of the correct tile width/heights.
func (box BoxDef) CreateBoxImage(width, height int) *ebiten.Image {
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
