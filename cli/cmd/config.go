package cmd

import (
	"github.com/webbben/2d-game-engine/internal/config"
	"github.com/webbben/2d-game-engine/internal/image"
)

func SetConfig() {
	// set config
	config.ShowPlayerCoords = true
	config.ShowGameDebugInfo = true
	// config.DrawGridLines = true
	// config.ShowEntityPositions = true
	// config.TrackMemoryUsage = true
	// config.HourSpeed = time.Second * 20
	// config.ShowCollisions = true
	// config.ShowNPCPaths = true

	config.GameDataPathOverride = "/Users/benwebb/dev/personal/ancient-rome"
	config.CharacterDefsDirectory = "/Users/benwebb/dev/personal/ancient-rome/src/data/characters/json"

	// CharacterDataSrc: "/Users/benwebb/dev/personal/ancient-rome/src/data/characters/json/character_02.json",

	config.DefaultFont = image.LoadFont("ashlander-pixel.ttf", 22, 0)
	config.DefaultTitleFont = image.LoadFont("ashlander-pixel.ttf", 28, 0)

	config.DefaultTooltipBox = config.DefaultBox{
		TilesetSrc:  "boxes/boxes.tsj",
		OriginIndex: 132,
	}
	config.DefaultUIBox = config.DefaultBox{
		TilesetSrc:  "boxes/boxes.tsj",
		OriginIndex: 16,
	}
}
