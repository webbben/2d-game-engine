package datamanager

import (
	"github.com/webbben/2d-game-engine/data/defs"
	"github.com/webbben/2d-game-engine/data/state"
)

// MapInfo gives the resolved values for any map Info that may be overridden from the def to the state.
type MapInfo struct {
	DisplayName string
	RegionID    defs.RegionID
}

// GetAllMapData gets all the data for a map, given its MapID (for the State - generated maps don't share mapID with original def!).
// Created this since maps have gotten a bit complicated, with the addition of generated maps, overrides in map state, etc.
// So this can serve as a "source of truth" for possibly overridden values (MapInfo) as well as a central place to get the right map state and def,
// given the ID that the map STATE uses.
func (dm *DataManager) GetAllMapData(mapStateID defs.MapID) (MapInfo, defs.MapDef, *state.MapState) {
	mapState := dm.GetMapState(mapStateID)

	mapInfo := MapInfo{}

	var mapDef defs.MapDef

	if mapState.IsGenerated {
		mapDef = dm.GetMapDef(mapState.GeneratedMapDefID)
	} else {
		mapDef = dm.GetMapDef(mapStateID)
	}

	if mapState.DisplayName != "" {
		mapInfo.DisplayName = mapState.DisplayName
	} else {
		mapInfo.DisplayName = mapDef.DisplayName
	}

	if mapState.RegionID != "" {
		mapInfo.RegionID = mapState.RegionID
	} else {
		mapInfo.RegionID = mapDef.Region
	}

	return mapInfo, mapDef, mapState
}
