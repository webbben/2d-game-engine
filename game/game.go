package game

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/webbben/2d-game-engine/camera"
	"github.com/webbben/2d-game-engine/entity"
	"github.com/webbben/2d-game-engine/object"
	"github.com/webbben/2d-game-engine/player"
	"github.com/webbben/2d-game-engine/room"
)

// information about the current room the player is in
type RoomInfo struct {
	Room     room.Room                // the room the player is currently in
	Entities []entity.Entity          // the entities in the current room
	Objects  []object.Object          // the objects in the current room
	ImageMap map[string]*ebiten.Image // the map of images (tiles) used in rendering the current room
}

// game state
type Game struct {
	RoomInfo
	Player player.Player // the player
	Camera camera.Camera // the camera/viewport
}
