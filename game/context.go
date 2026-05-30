package game

import (
	"github.com/webbben/2d-game-engine/clock"
	"github.com/webbben/2d-game-engine/data/defs"
	"github.com/webbben/2d-game-engine/data/state"
	"github.com/webbben/2d-game-engine/logz"
)

// PlacePlayerInMap is the same as EnterMap, but for putting the player at a specific position (instead of just at a spawn point).
// Used by the LoadGame flow, since you will be appearing not at a spawn point, but at the position you were in last when the game was saved.
func (g *Game) PlacePlayerInMap(mapID defs.MapID, x, y float64, doTransition bool) {
	if g.World == nil {
		panic("world was nil! you can't use this context function if the world doesn't exist yet...")
	}
	g.World.EnterMapAtPosition(mapID, x, y, doTransition)
}

func (g *Game) GetPlayerInfo() defs.PlayerInfo {
	charDef := g.Dataman.GetCharacterDef(g.World.Player.CharacterStateRef.DefID)
	return defs.PlayerInfo{
		PlayerName:    g.World.Player.CharacterStateRef.DisplayName,
		PlayerCulture: charDef.CultureID,
	}
}

func (g *Game) StartTradeSession(shopkeeperID defs.ShopID) {
	// shopkeeperDef := g.Dataman.GetShopkeeperDef(shopkeeperID)
	// shopkeeperState := g.Dataman.GetShopkeeperState(shopkeeperID)
	// g.TradeScreen.SetupTradeSession(*shopkeeperDef, shopkeeperState)
	// g.ShowTradeScreen = true
	logz.TODO("StartTradeSession", "is this used? everything in it is commented out")
}

// StartDialogSession starts a dialog session with the given dialog profile ID
func (g *Game) StartDialogSession(dialogProfileID defs.DialogProfileID, npcID string) {
	if npcID == "" {
		panic("npcID was empty")
	}
	if g.World == nil {
		panic("world was nil")
	}
	if g.World.ActiveMap == nil {
		panic("active map was nil. we need to be in an active map before starting a dialog")
	}

	g.World.ActiveMap.StartDialog(dialogProfileID, npcID)
}

func (g *Game) BroadcastEvent(e defs.Event) {
	// TODO: this is technically part of WorldEffectContext... should we use the World function?
	// Technically events could be used outside of the world, so no real reason to limit it to only being a world effect.
	g.EventBus.Publish(e)
}

func (g *Game) SetGameTime(gt clock.GameTime) {
	g.World.Clock.SetGameTime(gt)
}

func (g Game) GetMapID() defs.MapID {
	if g.World == nil {
		return ""
	}
	if g.World.ActiveMap == nil {
		return ""
	}
	return g.World.ActiveMap.MapID
}

func (g Game) GetActiveMapDef() defs.MapDef {
	if g.World == nil {
		logz.Panicln("GetActiveMapDef", "tried to get active map def, but world is nil")
	}
	if g.World.ActiveMap == nil {
		logz.Panicln("GetActiveMapDef", "tried to get active map def, but active map is nil")
	}
	mapID := g.World.ActiveMap.MapID
	mapDef := g.Dataman.GetMapDef(mapID)
	return mapDef
}

func (g Game) GetPlayerInventoryRef() *state.StandardInventory {
	if g.World == nil {
		panic("world was nil")
	}
	if g.World.Player == nil {
		panic("player was nil")
	}

	return &g.World.Player.CharacterStateRef.StandardInventory
}

func (g *Game) StartTimeLapse(newTime clock.GameTime) {
	g.World.TimeLapse(newTime)
}

func (g *Game) ShowMiscScreen(scrID defs.ScreenID, params any) {
	if g.World == nil {
		panic("world was nil")
	}

	scr := g.ScreenManager.GetScreen(scrID)

	g.World.ShowMiscScreen(scr, params)
}

func (g *Game) GetHoverTargetInfo() (*defs.NPCInfo, *defs.ObjectInfo) {
	n, obj := g.World.ActiveMap.GetHoverTarget()
	if n != nil {
		info := n.GetInfo()
		return &info, nil
	}
	if obj != nil {
		info := obj.GetInfo()
		return nil, &info
	}

	return nil, nil
}

func (g *Game) GetItemDef(itemID defs.ItemID) defs.ItemDef {
	return g.Dataman.GetItemDef(itemID)
}
