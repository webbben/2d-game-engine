package cmd

import (
	"github.com/webbben/2d-game-engine/data/defs"
	"github.com/webbben/2d-game-engine/definitions"
	"github.com/webbben/2d-game-engine/internal/audio"
	"github.com/webbben/2d-game-engine/internal/config"
	"github.com/webbben/2d-game-engine/internal/general_util"
)

func LoadDefMgr(defMgr *definitions.DefinitionManager) {
	// ITEMS
	itemDefs := GetItemDefs()
	defMgr.LoadItemDefs(itemDefs)

	// BODY PARTS
	bodySkins, eyeSets, hairSets := GetAllEntityBodyPartSets()
	for _, skin := range bodySkins {
		defMgr.LoadBodyPartDef(skin.Body)
		defMgr.LoadBodyPartDef(skin.Arms)
		defMgr.LoadBodyPartDef(skin.Legs)
	}
	for _, eyes := range eyeSets {
		defMgr.LoadBodyPartDef(eyes)
	}
	for _, hair := range hairSets {
		defMgr.LoadBodyPartDef(hair)
	}

	// CHARACTER DEFS
	charDefFilePaths := general_util.GetListOfFiles(config.CharacterDefsDirectory, true)
	for _, fp := range charDefFilePaths {
		var charDef defs.CharacterDef
		general_util.LoadJSONIntoStruct(fp, &charDef)
		defMgr.LoadCharacterDef(charDef)
	}

	// SHOPKEEPERS
	shopKeeperInventory := []defs.InventoryItem{}
	shopKeeperInventory = append(shopKeeperInventory, defMgr.NewInventoryItem("longsword_01", 1))
	shopkeeper := defs.ShopkeeperDef{
		ID:            "aurelius_tradehouse",
		ShopName:      "Aurelius' Tradehouse",
		BaseInventory: shopKeeperInventory,
		BaseGold:      1200,
	}
	defMgr.LoadShopkeeperDef(shopkeeper)

	// DIALOG
	// register topics
	topics := GetDialogTopics()
	for _, t := range topics {
		defMgr.LoadDialogTopic(&t)
	}
	// register profiles
	profiles := GetDialogProfiles()
	for _, p := range profiles {
		defMgr.LoadDialogProfile(&p)
	}

	// SOUNDS
	for _, footstepSfxDef := range GetFootstepSFXDefs() {
		defMgr.LoadFootstepSFXDef(footstepSfxDef)
	}

	// SCHEDULES
	for _, sched := range GetAllSchedules() {
		defMgr.LoadScheduleDef(sched)
	}
}

func LoadAudioManager(audioMgr *audio.AudioManager) {
	LoadAllSoundEffects(audioMgr)
}
