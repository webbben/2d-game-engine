package object

import (
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/webbben/2d-game-engine/internal/config"
	"github.com/webbben/2d-game-engine/internal/image"
	"github.com/webbben/2d-game-engine/internal/model"
	"github.com/webbben/2d-game-engine/internal/rendering"
)

type Object struct {
	PosX, PosY            int      // position this object is drawn from (top-left corner of image)
	OffsetX, OffsetY      int      // add an offset to the X position - in case the image dimensions push it slightly outside its intended tiles
	TileWidth, TileHeight int      // number of tiles this object occupies
	CollisionMap          [][]bool // map of which tiles are collisions
	img                   *ebiten.Image
	ImageSource           string

	CanSeeBehind bool // if true, this object will become transparent when the player is standing behind it, so the player can see

	Door

	WorldContext
}

func (o Object) DrawPos(offsetX, offsetY float64) (drawX, drawY float64) {
	drawX, drawY = rendering.GetImageDrawPos(o.img, float64(o.PosX+o.OffsetX), float64(o.PosY+o.OffsetY), offsetX, offsetY)
	return drawX, drawY
}

func (o Object) ExtentPos(offsetX, offsetY float64) (extentX, extentY float64) {
	extentX, extentY = o.DrawPos(offsetX, offsetY)
	extentX += float64(o.img.Bounds().Dx())
	extentY += float64(o.img.Bounds().Dy())
	return extentX, extentY
}

type WorldContext interface {
	PlayerIsBehindObject(obj Object) bool
}

// for sorting among other renderables in map
func (o Object) Y() float64 {
	_, extentY := o.ExtentPos(0, 0)
	return extentY
}

type Door struct {
	IsDoor                        bool // if true, this object can be used as a door to another map
	DoorX, DoorY                  int  // position of the tiles which represent the door (relative to main X/Y)
	DoorTileWidth, DoorTileHeight int  // tile size of the door
}

func NewObject(imgSource string, tileWidth, tileHeight int) (Object, error) {
	obj := Object{
		TileWidth:  tileWidth,
		TileHeight: tileHeight,
	}
	img, err := image.LoadImage(imgSource)
	if err != nil {
		return Object{}, fmt.Errorf("failed to load source image: %w", err)
	}
	obj.img = img

	obj.CollisionMap = createCollisionMap(tileWidth, tileHeight)

	obj.validate()

	return obj, nil
}

func (o Object) validate() {
	if len(o.CollisionMap) != o.TileHeight {
		panic("object collision map height doesn't match tile height")
	}
	if len(o.CollisionMap[0]) != o.TileWidth {
		panic("object collision map width doesn't match tile width")
	}
}

// places the object based on the given door position, which is assumed to be at the middle bottom of the image.
//
// The coords should be the position of a tile. It will be rounded down to one, if it's not at an exact tile origin.
func (o *Object) PlaceByDoorCoords(doorCoords model.Coords, doorTileWidth, doorTileHeight int) {
	width := o.img.Bounds().Dx()
	height := o.img.Bounds().Dy()

	absDoorX := doorCoords.X * config.TileSize
	absDoorY := doorCoords.Y * config.TileSize

	o.DoorX = absDoorX
	o.DoorY = absDoorY
	o.DoorTileWidth = doorTileWidth
	o.DoorTileHeight = doorTileHeight

	// determine placement X position
	width -= o.DoorTileWidth * config.TileSize
	width /= 2
	placementX := absDoorX - width

	// determine placement Y position
	height -= o.DoorTileHeight * config.TileSize
	height /= 2
	placementY := absDoorY - height

	o.PosX = placementX
	o.PosY = placementY
}

func createCollisionMap(width, height int) [][]bool {
	out := [][]bool{}

	for range height {
		row := make([]bool, width)
		for j := range width {
			row[j] = true
		}
		out = append(out, row)
	}

	return out
}

func (o Object) Draw(screen *ebiten.Image, offsetX, offsetY float64) {
	op := ebiten.DrawImageOptions{}
	if o.CanSeeBehind && o.WorldContext.PlayerIsBehindObject(o) {
		op.ColorScale.ScaleAlpha(0.5)
	}
	drawX, drawY := o.DrawPos(offsetX, offsetY)
	op.GeoM.Translate(drawX, drawY)
	op.GeoM.Scale(config.GameScale, config.GameScale)
	screen.DrawImage(o.img, &op)
}
