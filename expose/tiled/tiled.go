// Package tiled (exposed) is simply a wrapper around internal tiled so we can expose some specific functions.
// generally, tiled package should probably be used internally, but there are some specific cases where it's good to give access.
package tiled

import int_tiled "github.com/webbben/2d-game-engine/tiled"

func GetTileBoolProperty(tilesetSrc string, tileIndex int, propName string) (val bool, found bool) {
	tileset, err := int_tiled.LoadTileset(tilesetSrc)
	if err != nil {
		panic(err)
	}

	for _, tile := range tileset.Tiles {
		if tile.ID != tileIndex {
			continue
		}
		return int_tiled.GetBoolProperty(propName, tile.Properties)
	}

	return false, false
}
