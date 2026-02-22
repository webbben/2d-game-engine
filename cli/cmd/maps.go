package cmd

import "github.com/webbben/2d-game-engine/data/defs"

const (
	MapAquileiaPrisonShip defs.MapID = "prison_ship"
	MapAquileiaHarbor     defs.MapID = "aquileia_harbor"
)

func GetAllMapDefs() []defs.MapDef {
	maps := []defs.MapDef{
		{
			ID: MapAquileiaPrisonShip,
		},
		{
			ID: MapAquileiaHarbor,
		},
	}

	return maps
}
