package room

import (
	"github.com/webbben/2d-game-engine/internal/general_util"
	"github.com/webbben/2d-game-engine/internal/model"
)

// euclidean distance function for coords structs, for ease of use
func euclideanDist(pointA, pointB model.Coords) float64 {
	return general_util.EuclideanDist(float64(pointA.X), float64(pointA.Y), float64(pointB.X), float64(pointB.Y))
}

// checks if the given position is within the room bounds
func posInRoomBounds(pos model.Coords, width, height int) bool {
	if pos.X < 0 || pos.Y < 0 {
		return false
	}
	if pos.X >= width || pos.Y >= height {
		return false
	}
	return true
}
