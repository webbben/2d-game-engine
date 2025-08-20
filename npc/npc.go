package npc

import "github.com/webbben/2d-game-engine/entity"

type NPC struct {
	NPCInfo
	Entity *entity.Entity
}

type NPCInfo struct {
	ID string
}
