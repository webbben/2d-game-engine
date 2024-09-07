package general_util

import (
	"math"
	"math/rand"

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
