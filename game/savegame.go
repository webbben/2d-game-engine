package game

import (
	"github.com/webbben/2d-game-engine/config"
	"github.com/webbben/2d-game-engine/data/defs"
	"github.com/webbben/2d-game-engine/data/savegame"
	"github.com/webbben/2d-game-engine/data/state"
	"github.com/webbben/2d-game-engine/entity"
	"github.com/webbben/2d-game-engine/entity/player"
	"github.com/webbben/2d-game-engine/logz"
)

func (g Game) SaveGame() (saveFilePath string) {
	return savegame.SaveGame(
		g.Dataman,
		g.QuestManager,
		g.Clock.GetCurrentGameTime(),
		g.MapInfo.MapID,
		g.Player.Entity.TilePos(),
	)
}

// LoadGame loads a save file, sets up the game world/map, and places the player into that map.
// After calling this, the game should start to run in the game world.
func (g *Game) LoadGame(saveFilePath string) {
	worldInfo, err := savegame.LoadSave(saveFilePath, g.Dataman, g.QuestManager)
	if err != nil {
		logz.Panicln("LoadGame", "failed to load save. err:", err)
	}

	// all states should've been loaded in LoadSave; handle the worldInfo and placing the player in the world.
	g.gameStage = InGameWorld

	g.SetGameTime(worldInfo.CurrentTime)

	playerEnt := entity.LoadCharacterStateIntoEntity(state.CharacterStateID(defs.PlayerID), g.Dataman, g.AudioManager)
	p := player.NewPlayer(g.Dataman, playerEnt)
	x := worldInfo.CurrentMapCoords.X * config.TileSize
	y := worldInfo.CurrentMapCoords.Y * config.TileSize
	g.MapInfo.AddPlayerToMap(&p, float64(x), float64(y))
	g.Camera.SetCameraPosition(float64(x), float64(y))
}
