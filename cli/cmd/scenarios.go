package cmd

import "github.com/webbben/2d-game-engine/data/defs"

const (
	ScenarioQ001PrisonShip    defs.ScenarioID = "Q001_prison_ship"
	ScenarioQ001Harbor        defs.ScenarioID = "Q001_harbor"
	ScenarioQ001CustomsOffice defs.ScenarioID = "Q001_customs_office"
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
			Characters: []defs.ScenarioCharDef{
				{
					CharDefID:   CharQ001ShipCaptain,
					SpawnCoordX: 33,
					SpawnCoordY: 50,
				},
				{
					CharDefID:   CharQ001MiscGuard,
					SpawnCoordX: 15,
					SpawnCoordY: 30,
				},
				{
					CharDefID:   CharQ001MiscGuard,
					SpawnCoordX: 15,
					SpawnCoordY: 31,
				},
				{
					CharDefID:   CharQ001MiscGuard,
					SpawnCoordX: 53,
					SpawnCoordY: 30,
				},
				{
					CharDefID:   CharQ001MiscGuard,
					SpawnCoordX: 53,
					SpawnCoordY: 31,
				},
				{
					CharDefID:   CharQ001MiscGuard,
					SpawnCoordX: 32,
					SpawnCoordY: 28,
				},
				{
					CharDefID:   CharQ001MiscGuard,
					SpawnCoordX: 36,
					SpawnCoordY: 28,
				},
				{
					CharDefID:   CharQ001MiscGuard,
					SpawnCoordX: 33,
					SpawnCoordY: 47,
				},
			},
		},
		{
			ID:    ScenarioQ001CustomsOffice,
			MapID: MapAquileiaCustomsOffice,
			Characters: []defs.ScenarioCharDef{
				{
					CharDefID:   CharQ001CustomsOfficer,
					SpawnCoordX: 12,
					SpawnCoordY: 11,
				},
				{
					CharDefID:   CharQ001MiscGuard,
					SpawnCoordX: 5,
					SpawnCoordY: 15,
				},
			},
		},
	}

	return scenarios
}
