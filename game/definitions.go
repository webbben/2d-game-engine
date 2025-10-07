package game

import "github.com/webbben/2d-game-engine/item"

/*
Purpose: a global, central location for loading definitions of different in-game data.

For example, when loading item definitions, we want a central place for all those to be stored.
Then, when an item is present somewhere, its data can be loaded/reloaded from here rather than constantly
re-loading the file data.

Tracked data:

- Item Defs
*/

func (g *Game) LoadItemDefs(itemDefs []item.ItemDef) {
	if g.ItemDefs == nil {
		g.ItemDefs = map[string]item.ItemDef{}
	}
	for _, itemDef := range itemDefs {
		id := itemDef.GetID()
		itemDef.Load()
		g.ItemDefs[id] = itemDef
	}
}
