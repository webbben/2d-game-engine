package game

import (
	"github.com/webbben/2d-game-engine/config"
	"github.com/webbben/2d-game-engine/data/defs"
	"github.com/webbben/2d-game-engine/data/savegame"
	"github.com/webbben/2d-game-engine/logz"
	"github.com/webbben/2d-game-engine/model"
)

func (g Game) SaveGame() (saveFilePath string) {
	mapCoords := g.World.Player.Entity.TilePos()
	if g.World.Player.Entity.IsSleeping {
		// if the player is in a bed, get the position they'd be standing in if they left the bed
		_, vec, _ := g.World.Player.Entity.GetSleepInfo()
		mapCoords = model.ConvertPxToTilePos(vec.X, vec.Y)
	}

	if g.World.ActiveMap.IsTileCollision(mapCoords) {
		logz.Panicln("SAVE", "map coords that were going to be saved are apparently a collision in the map!", mapCoords)
	}

	return savegame.SaveGame(
		g.Dataman,
		g.QuestManager,
		g.EventBus,
		g.World.Clock.GetCurrentGameTime(),
		g.World.ActiveMap.MapID,
		mapCoords,
	)
}

// LoadGame loads a save file, sets up the game world/map, and places the player into that map.
// After calling this, the game should start to run in the game world.
func (g *Game) LoadGame(saveFilePath string) {
	worldInfo, err := savegame.LoadSave(saveFilePath, g.Dataman, g.QuestManager, g.EventBus)
	if err != nil {
		logz.Panicln("LoadGame", "failed to load save. err:", err)
	}

	g.InitializeGameWorld(worldInfo.CurrentTime)

	x := worldInfo.CurrentMapCoords.X * config.TileSize
	y := worldInfo.CurrentMapCoords.Y * config.TileSize
	g.PlacePlayerInMap(worldInfo.CurrentMapID, float64(x), float64(y), false)

	g.gameStage = InGameWorld
}

func (g Game) GetAllExistingCharacters() []defs.ExistingCharacterInfo {
	return savegame.GetAllExistingCharacters()
}
