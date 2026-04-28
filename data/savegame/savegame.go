// Package savegame contains logic that handles saving a game
package savegame

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/webbben/2d-game-engine/clock"
	"github.com/webbben/2d-game-engine/config"
	"github.com/webbben/2d-game-engine/data/datamanager"
	"github.com/webbben/2d-game-engine/data/defs"
	"github.com/webbben/2d-game-engine/data/id"
	"github.com/webbben/2d-game-engine/data/state"
	characterstate "github.com/webbben/2d-game-engine/entity/characterState"
	"github.com/webbben/2d-game-engine/logz"
	"github.com/webbben/2d-game-engine/model"
	"github.com/webbben/2d-game-engine/quest"
	"github.com/webbben/2d-game-engine/utils/files"
)

const (
	TimestampLayout string = "20060102150405"
)

// SaveFile contains all the data that can be saved and loaded for an individual playthrough.
type SaveFile struct {
	SaveTime time.Time

	// Player data that is only defined in the save file:
	// we store the player's character def directly in the save file, instead of a JSON file like the other character defs.
	// the same for the classDef - the default ones are saved in code, but we only store the player's class def here.

	PlayerCustomClass  defs.ClassDef
	PlayerCharacterDef defs.CharacterDef

	// player location, general world info where the player exists
	CurrentMapID    defs.MapID
	MapCoords       model.Coords
	CurrentGameTime clock.GameTimestamp

	// TODO: do we need to store information about NPCs around the player?
	// Or, should saves only happen in private maps where other NPCs aren't around?
	// Currently, I'm thinking that saves will only happen when you sleep, and the "save" will be
	// set for the moment you are awake (similar to Stardew Valley). So, I'm not sure that it's important
	// to save NPC state, since they would theoretically have moved to different places and be doing different things
	// by the time the player has woken up.

	// state data for the rest of the game world

	CharacterStates     []state.CharacterState
	DialogProfileStates []state.DialogProfileState
	MapStates           []state.MapState
	ShopkeeperStates    []state.ShopkeeperState
	Quests              QuestStates
}

func (sf SaveFile) validate() {
	if sf.SaveTime.IsZero() {
		logz.Panic("save time is zero value")
	}
	if sf.CurrentMapID == "" {
		logz.Panic("current map ID is empty")
	}
	if sf.CurrentGameTime == "" {
		logz.Panic("CurrentGameTime is empty")
	}
	if len(sf.CharacterStates) == 0 {
		logz.Warnln("SaveFile", "no character states... this probably shouldn't happen, right?")
	}
	if len(sf.DialogProfileStates) == 0 {
		logz.Warnln("SaveFile", "no dialog profile states... this probably shouldn't happen, right?")
	}
	if len(sf.MapStates) == 0 {
		logz.Warnln("SaveFile", "no map states... this probably shouldn't happen, right?")
	}
	if len(sf.ShopkeeperStates) == 0 {
		logz.Warnln("SaveFile", "no shopkeeper states... this probably shouldn't happen, right?")
	}
}

type QuestStates struct {
	Active    []state.QuestState
	Completed []state.QuestState
	Failed    []state.QuestState
}

func SaveGame(
	dataman *datamanager.DataManager,
	questMgr *quest.QuestManager,
	gameTime clock.GameTime,
	mapID defs.MapID,
	mapCoords model.Coords,
) (saveFilePath string) {
	sf := SaveFile{
		SaveTime:        time.Now(),
		CurrentGameTime: gameTime.GetTimestamp(),
		CurrentMapID:    mapID,
		MapCoords:       mapCoords,
	}

	// sanity checks; make sure things exist
	_ = dataman.GetMapDef(mapID)
	_ = dataman.GetCharacterState(id.CharacterStateID(defs.PlayerID))

	// get player def
	playerDef := dataman.GetCharacterDef(defs.PlayerID)
	sf.PlayerCharacterDef = playerDef
	playerClass := dataman.GetClassDef(playerDef.ClassDefID)
	sf.PlayerCustomClass = playerClass

	uniqueID := sf.PlayerCharacterDef.UniquePlayerID
	characterstate.ValidateUniquePlayerID(uniqueID)

	for _, st := range dataman.CharacterStates {
		if st.Temp {
			// skip temporary character states, since they shouldn't be saved. (e.g. character states from scenarios)
			continue
		}
		sf.CharacterStates = append(sf.CharacterStates, *st)
	}
	for _, st := range dataman.DialogProfileStates {
		sf.DialogProfileStates = append(sf.DialogProfileStates, *st)
	}
	for _, st := range dataman.MapStates {
		sf.MapStates = append(sf.MapStates, *st)
	}
	for _, st := range dataman.ShopkeeperStates {
		sf.ShopkeeperStates = append(sf.ShopkeeperStates, *st)
	}

	active, comp, fail := questMgr.GetAllQuestStates()
	sf.Quests.Active = active
	sf.Quests.Completed = comp
	sf.Quests.Failed = fail

	timestamp := time.Now().Format(TimestampLayout)
	filename := fmt.Sprintf("%s.json", timestamp)

	config.EnsurePlayerSaveDirExists(uniqueID)
	saveFilePath = config.ResolveSaveFilePath(uniqueID, filename)

	// ensure this save file has no validation errors
	sf.validate()

	err := files.WriteToJSON(sf, saveFilePath)
	if err != nil {
		logz.Panicln("SAVEGAME", "error while writing save game file:", err.Error())
	}

	logz.Println("GAME SAVED", saveFilePath)

	return saveFilePath
}

type LoadedWorldInfo struct {
	CurrentTime      clock.GameTime
	CurrentMapID     defs.MapID
	CurrentMapCoords model.Coords
}

// LoadSave loads the data from a save file. You must pass data and quest managers that already have all of their defs loaded into them.
// The various states will be loaded into these managers.
func LoadSave(saveFilePath string, dataman *datamanager.DataManager, questMgr *quest.QuestManager) (LoadedWorldInfo, error) {
	if dataman == nil {
		panic("dataman was nil")
	}
	if questMgr == nil {
		panic("questMgr was nil")
	}

	info := LoadedWorldInfo{}

	// find save file
	if !config.FileExists(saveFilePath) {
		logz.Warnln("LOADGAME", "failed to find save file:", saveFilePath)
		return info, fmt.Errorf("failed to find save file: %s", saveFilePath)
	}

	// load JSON data into SaveFile
	sf := loadSaveFileStruct(saveFilePath)

	// get world info
	info.CurrentTime = clock.TimestampToGameTime(sf.CurrentGameTime)
	info.CurrentMapID = sf.CurrentMapID
	info.CurrentMapCoords = sf.MapCoords

	// load data and quest managers
	// player defs
	dataman.LoadCharacterDef(sf.PlayerCharacterDef)
	dataman.LoadClassDef(sf.PlayerCustomClass)

	// states
	for _, st := range sf.CharacterStates {
		dataman.LoadCharacterState(&st)
	}
	for _, st := range sf.DialogProfileStates {
		dataman.LoadDialogProfileState(&st)
	}
	for _, st := range sf.ShopkeeperStates {
		dataman.LoadShopkeeperState(st)
	}
	for _, st := range sf.MapStates {
		dataman.LoadMapState(st)
	}

	// quest states
	allQuestStates := []state.QuestState{}
	allQuestStates = append(allQuestStates, sf.Quests.Active...)
	allQuestStates = append(allQuestStates, sf.Quests.Completed...)
	allQuestStates = append(allQuestStates, sf.Quests.Failed...)
	for _, st := range allQuestStates {
		questMgr.LoadQuestState(st)
	}

	questMgr.CreateEventTypeIndices()

	return info, nil
}

func GetAllExistingCharacters() []defs.ExistingCharacterInfo {
	existingChars := []defs.ExistingCharacterInfo{}

	// each of these should be directories for a single character's saves
	characterSaveDirs := config.GetAllSaveDirs()

	if len(characterSaveDirs) == 0 {
		logz.Println("GetAllExistingCharacters", "no character save directories found.")
		return existingChars
	}

	// for each character save directory, get info about the character and the recent save
	for _, saveDir := range characterSaveDirs {
		charInfo := defs.ExistingCharacterInfo{}
		saveFiles := files.GetListOfFiles(saveDir, true)
		if len(saveFiles) == 0 {
			logz.Panicln("GetAllExistingCharacters", "a character saves directory was empty, and had no save files.")
		}
		// find the most recent save file
		var mostRecentSavePath string
		var mostRecentTime time.Time
		for _, saveFilePath := range saveFiles {
			charInfo.SaveFilePaths = append(charInfo.SaveFilePaths, saveFilePath)
			// the time of the save is in the filename; it's a timestamp.
			// so, parse the filenames into time data, and get the most recent one.
			name := filepath.Base(saveFilePath)
			timestamp := strings.Split(name, ".")[0]
			t, err := time.Parse(TimestampLayout, timestamp)
			if err != nil {
				logz.Panicln("GetAllExistingCharacters", "found save file with malformed timestamp filename. error:", err)
			}
			if t.After(mostRecentTime) || mostRecentSavePath == "" {
				mostRecentTime = t
				mostRecentSavePath = saveFilePath
			}
		}

		// load the data of the most recent save
		charInfo.RecentSave = GetSaveInfo(mostRecentSavePath)
		charInfo.DisplayName = charInfo.RecentSave.CharacterName
		charInfo.UniquePlayerID = charInfo.RecentSave.UniquePlayerID

		existingChars = append(existingChars, charInfo)
	}

	return existingChars
}

func GetSaveInfo(saveFilePath string) defs.SaveInfo {
	logz.Println("GetSaveInfo", saveFilePath)
	si := defs.SaveInfo{}
	si.SaveFilePath = saveFilePath

	sf := loadSaveFileStruct(saveFilePath)

	si.CharacterName = sf.PlayerCharacterDef.DisplayName
	si.LastPlay = sf.SaveTime
	si.UniquePlayerID = sf.PlayerCharacterDef.UniquePlayerID
	si.CurrentMapID = sf.CurrentMapID
	si.CurrentGameTime = clock.TimestampToGameTime(sf.CurrentGameTime)

	return si
}

func loadSaveFileStruct(saveFilePath string) SaveFile {
	var sf SaveFile

	fileData, err := os.ReadFile(saveFilePath)
	if err != nil {
		logz.Panicln("loadSaveFileStruct", "failed to read save file:", err)
	}

	err = json.Unmarshal(fileData, &sf)
	if err != nil {
		logz.Panicln("loadSaveFileStruct", "failed to unmarshal save file data:", err)
	}

	sf.validate()

	return sf
}
