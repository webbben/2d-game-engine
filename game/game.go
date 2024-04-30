package game

import (
	"fmt"
	"image/color"
	"sort"

	"github.com/webbben/2d-game-engine/camera"
	"github.com/webbben/2d-game-engine/config"
	"github.com/webbben/2d-game-engine/entity"
	"github.com/webbben/2d-game-engine/model"
	"github.com/webbben/2d-game-engine/object"
	"github.com/webbben/2d-game-engine/player"
	"github.com/webbben/2d-game-engine/room"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// information about the current room the player is in
type RoomInfo struct {
	Room     room.Room                // the room the player is currently in
	Entities []entity.Entity          // the entities in the current room
	Objects  []object.Object          // the objects in the current room
	ImageMap map[string]*ebiten.Image // the map of images (tiles) used in rendering the current room
}

// pre-processes data in the room for various purposes, but mostly to prepare for rendering
// and performance improvements
func (ri *RoomInfo) Preprocess() {
	sort.Slice(ri.Objects, func(i, j int) bool {
		return ri.Objects[i].Y < ri.Objects[j].Y
	})
}

// game state
type Game struct {
	RoomInfo
	Player player.Player // the player
	Camera camera.Camera // the camera/viewport
}

// generates a cost map for the contents of the game state
//
// currently includes:
//
// * entities in the room
func (g Game) GenerateCostMap() [][]int {
	costMap := make([][]int, g.Room.Height)
	for i := 0; i < len(costMap); i++ {
		costMap[i] = make([]int, g.Room.Width)
	}
	for i := 0; i < len(g.Entities); i++ {
		costMap[int(g.Entities[i].Y)][int(g.Entities[i].X)] = 10
	}
	return costMap
}

func (g *Game) Update() error {
	// Your game logic goes here

	// handle player updates
	g.Player.Update(g.Room.BarrierLayout)

	// move camera as needed
	g.Camera.MoveCamera(g.Player.X, g.Player.Y)

	// sort entities by Y position for rendering
	if len(g.Entities) > 1 {
		sort.Slice(g.Entities, func(i, j int) bool {
			return g.Entities[i].Y < g.Entities[j].Y
		})
	}

	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	offsetX, offsetY := g.Camera.GetAbsPos()

	// draw the terrain of the room
	g.Room.DrawFloor(screen, offsetX, offsetY)
	g.Room.DrawCliffs(screen, offsetX, offsetY)
	if config.DrawGridLines {
		g.drawGridLines(screen, offsetX, offsetY)
	}

	// draw objects, entities, and the player in order of Y position (higher renders first)
	o, e := 0, 0
	obj, ent := g.Objects[0], g.Entities[0]
	playerDrawn := false
	for o < len(g.Objects) || e < len(g.Entities) {
		if !playerDrawn && g.Player.Y <= ent.Y && g.Player.Y <= float64(obj.Y) {
			g.Player.Draw(screen, offsetX, offsetY)
			playerDrawn = true
		}
		if o < len(g.Objects) && float64(obj.Y) < ent.Y {
			// if there are objects to draw still and this one is higher than the next entity
			obj.Draw(screen, offsetX, offsetY, g.ImageMap)
			o++
			if o < len(g.Objects) {
				obj = g.Objects[o]
			} else {
				obj = object.Object{Y: 99999} // no more objects; set to high Y value so it won't stop other things from being drawn
			}
		} else {
			// else - if there are no objects to draw anymore, or this entity is same or higher than the next object
			ent.Draw(screen, offsetX, offsetY)
			e++
			if e < len(g.Entities) {
				ent = g.Entities[e]
			} else {
				ent = entity.Entity{Position: model.Position{Y: 99999}} // no more entities
			}
		}
	}
	if !playerDrawn {
		g.Player.Draw(screen, offsetX, offsetY)
	}

	if config.ShowPlayerCoords {
		ebitenutil.DebugPrint(screen, fmt.Sprintf("Player pos: [%v, %v]", g.Player.X, g.Player.Y))
	}
}

func (g *Game) drawGridLines(screen *ebiten.Image, offsetX float64, offsetY float64) {
	offsetX = offsetX * config.GameScale
	offsetY = offsetY * config.GameScale
	lineColor := color.RGBA{255, 0, 0, 255}
	maxWidth := float64(len(g.Room.TileLayout[0]) * config.TileSize * config.GameScale)
	maxHeight := float64(len(g.Room.TileLayout) * config.TileSize * config.GameScale)

	// vertical lines
	for x := 0; x < len(g.Room.TileLayout[0]); x++ {
		drawX := float64(x*config.TileSize*config.GameScale) - offsetX
		drawY := -offsetY
		vector.StrokeLine(screen, float32(drawX), float32(drawY), float32(drawX), float32(maxHeight-offsetY), 1, lineColor, true)
	}
	// horizontal lines
	for y := 0; y < len(g.Room.TileLayout); y++ {
		drawX := -offsetX
		drawY := float64(y*config.TileSize*config.GameScale) - offsetY
		vector.StrokeLine(screen, float32(drawX), float32(drawY), float32(maxWidth-offsetX), float32(drawY), 1, lineColor, true)
	}
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return config.ScreenWidth, config.ScreenHeight
}
