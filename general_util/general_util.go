package general_util

import (
	"math"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/webbben/2d-game-engine/model"
)

func EuclideanDist(x1, y1, x2, y2 float64) float64 {
	return math.Sqrt(math.Pow(x2-x1, 2) + math.Pow(y2-y1, 2))
}

// euclidean distance function for coords structs, for ease of use
func EuclideanDistCoords(pointA, pointB model.Coords) float64 {
	return EuclideanDist(float64(pointA.X), float64(pointA.Y), float64(pointB.X), float64(pointB.Y))
}

func RandInt(min, max int) int {
	return rand.Intn(max-min+1) + min
}

func RoundToDecimal(val float64, decimals int) float64 {
	factor := math.Pow(10, float64(decimals))
	return math.Round(val*factor) / factor
}

// IsHovering returns true if the mouse cursor is hovering within the given coordinates box
func IsHovering(x1, y1, x2, y2 int) bool {
	x, y := ebiten.CursorPosition()
	return x >= x1 && x <= x2 && y >= y1 && y <= y2
}

// IsClicked returns true if the left mouse button is clicked within the given coordinates box
func IsClicked(x1, y1, x2, y2 int) bool {
	if !ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
		return false
	}
	return IsHovering(x1, y1, x2, y2)
}
