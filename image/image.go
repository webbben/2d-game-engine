package image

import (
	"errors"
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/webbben/2d-game-engine/tileset"
)

// loads an individual image
func LoadImage(imagePath string) (*ebiten.Image, error) {
	img, _, err := ebitenutil.NewImageFromFile(imagePath)
	if err != nil {
		return nil, err
	}
	return img, nil
}

type BoxTiles struct {
	Top, TopLeft, TopRight, Left, Right, BottomLeft, Bottom, BottomRight, Fill *ebiten.Image
}

// Verify checks that all images are non-nil and have the same dimensions
func (bt BoxTiles) Verify() error {
	// check for nil images
	if bt.Top == nil {
		return errors.New("top image is nil")
	}
	if bt.TopLeft == nil {
		return errors.New("topLeft image is nil")
	}
	if bt.TopRight == nil {
		return errors.New("topRight image is nil")
	}
	if bt.Left == nil {
		return errors.New("left image is nil")
	}
	if bt.Right == nil {
		return errors.New("right image is nil")
	}
	if bt.BottomLeft == nil {
		return errors.New("bottomLeft image is nil")
	}
	if bt.Bottom == nil {
		return errors.New("bottom image is nil")
	}
	if bt.BottomRight == nil {
		return errors.New("bottomRight image is nil")
	}
	if bt.Fill == nil {
		return errors.New("fill image is nil")
	}
	// make sure all images have same dimensions
	width, height := bt.Top.Bounds().Dx(), bt.Top.Bounds().Dy() // use this as the reference size
	if width != bt.TopLeft.Bounds().Dx() || height != bt.TopLeft.Bounds().Dy() {
		return errors.New("topLeft image has different dimensions")
	}
	if width != bt.TopRight.Bounds().Dx() || height != bt.TopRight.Bounds().Dy() {
		return errors.New("topRight image has different dimensions")
	}
	if width != bt.Left.Bounds().Dx() || height != bt.Left.Bounds().Dy() {
		return errors.New("left image has different dimensions")
	}
	if width != bt.Right.Bounds().Dx() || height != bt.Right.Bounds().Dy() {
		return errors.New("right image has different dimensions")
	}
	if width != bt.BottomLeft.Bounds().Dx() || height != bt.BottomLeft.Bounds().Dy() {
		return errors.New("bottomLeft image has different dimensions")
	}
	if width != bt.Bottom.Bounds().Dx() || height != bt.Bottom.Bounds().Dy() {
		return errors.New("bottom image has different dimensions")
	}
	if width != bt.BottomRight.Bounds().Dx() || height != bt.BottomRight.Bounds().Dy() {
		return errors.New("bottomRight image has different dimensions")
	}
	if width != bt.Fill.Bounds().Dx() || height != bt.Fill.Bounds().Dy() {
		return errors.New("fill image has different dimensions")
	}
	return nil
}

// LoadBoxTileSet loads a set of box tiles from a tileset image folder
//
// tilesetPath: the path to the tileset image folder
//
// returns: the box tiles and an error if one occurred
//
// The tileset image folder must contain the following images:
// T.png: the top tile, TL.png: the top left tile, TR.png: the top right tile,
// L.png: the left tile, R.png: the right tile, BL.png: the bottom left tile,
// B.png: the bottom tile, BR.png: the bottom right tile, F.png: the fill tile
func LoadBoxTileSet(tilesetPath string) (BoxTiles, error) {
	tiles := BoxTiles{}
	tileset, err := tileset.LoadTilesetByPath(tilesetPath)
	if err != nil {
		fmt.Println("failed to set dialog tiles:", err)
		return BoxTiles{}, err
	}
	tiles.Top = tileset["T"]
	tiles.TopLeft = tileset["TL"]
	tiles.TopRight = tileset["TR"]
	tiles.Left = tileset["L"]
	tiles.Right = tileset["R"]
	tiles.BottomLeft = tileset["BL"]
	tiles.Bottom = tileset["B"]
	tiles.BottomRight = tileset["BR"]
	tiles.Fill = tileset["F"]

	return tiles, tiles.Verify()
}

// CreateBox creates a box image from the given tiles
//
// numTilesWide: the number of tiles wide the box should be
//
// numTilesHigh: the number of tiles high the box should be
//
// t: the box tiles
//
// borderOpacity: the opacity of the border tiles (alpha scale value)
//
// fillOpacity: the opacity of the fill tiles (alpha scale value)
func CreateBox(numTilesWide, numTilesHigh int, t BoxTiles, borderOpacity float32, fillOpacity float32) *ebiten.Image {
	err := t.Verify()
	if err != nil {
		panic(err)
	}
	tileSize := t.Top.Bounds().Dx()
	box := ebiten.NewImage(numTilesWide*tileSize, numTilesHigh*tileSize)
	for x := 0; x < numTilesWide; x++ {
		for y := 0; y < numTilesHigh; y++ {
			// get the image we will place
			var img *ebiten.Image
			op := &ebiten.DrawImageOptions{}
			a := borderOpacity
			if x == 0 {
				if y == 0 {
					// top left
					img = t.TopLeft
				} else if y == numTilesHigh-1 {
					// bottom left
					img = t.BottomLeft
				} else {
					// left
					img = t.Left
				}
			} else if x == numTilesWide-1 {
				if y == 0 {
					// top right
					img = t.TopRight
				} else if y == numTilesHigh-1 {
					// bottom right
					img = t.BottomRight
				} else {
					// right
					img = t.Right
				}
			} else if y == 0 {
				img = t.Top
			} else if y == numTilesHigh-1 {
				img = t.Bottom
			} else {
				img = t.Fill
				a = fillOpacity
			}
			// draw the tile
			op.ColorScale.ScaleAlpha(a)
			op.GeoM.Translate(float64(x*tileSize), float64(y*tileSize))
			box.DrawImage(img, op)
		}
	}
	return box
}
