package cmd

import "github.com/webbben/2d-game-engine/data/defs"

const (
	ScenarioQ001PrisonShip defs.ScenarioID = "Q001_prison_ship"
	ScenarioQ001Harbor     defs.ScenarioID = "Q001_harbor"
)

func GetAllScenarios() []defs.ScenarioDef {
	scenarios := []defs.ScenarioDef{
		{
			ID:    ScenarioQ001PrisonShip,
			MapID: MapAquileiaPrisonShip,
			Characters: []defs.ScenarioCharDef{
				{
					CharDefID:       CharJovePrisonShip,
					DefaultSchedule: ScheduleIdle,
					SpawnCoordX:     14,
					SpawnCoordY:     35,
				},
				{
					CharDefID:       CharPrisonShipGuard01,
					DefaultSchedule: ScheduleIdle,
					SpawnCoordX:     6,
					SpawnCoordY:     21,
				},
			},
		},
		{
			ID:    ScenarioQ001Harbor,
			MapID: MapAquileiaHarbor,
			// TODO: make characters for this scenario
			Characters: []defs.ScenarioCharDef{},
		},
	}

	return scenarios
}
