// Package characterstate defines everything that goes into a characters (npc, or player) state.
package characterstate

import (
	"fmt"
	"math"

	"github.com/webbben/2d-game-engine/data/defs"
	"github.com/webbben/2d-game-engine/data/state"
	"github.com/webbben/2d-game-engine/data/datamanager"
	"github.com/webbben/2d-game-engine/logz"
	"github.com/webbben/2d-game-engine/item"
)

// GetNetTraitModifiers returns all of the net modifiers on skills produced by the given traits
func GetNetTraitModifiers(traits []defs.TraitID, dataman *datamanager.DataManager) (skillMods map[defs.SkillID]int, attrMods map[defs.AttributeID]int) {
	if dataman == nil {
		panic("dataman was nil")
	}
	skillMods = make(map[defs.SkillID]int)
	attrMods = make(map[defs.AttributeID]int)

	for _, traitID := range traits {
		trait := dataman.GetTraitDef(traitID)
		for id, change := range trait.GetSkillChanges() {
			if _, exists := skillMods[id]; !exists {
				skillMods[id] = 0
			}
			skillMods[id] += change
		}
		for id, change := range trait.GetAttributeChanges() {
			if _, exists := attrMods[id]; !exists {
				attrMods[id] = 0
			}
			attrMods[id] += change
		}
	}

	return skillMods, attrMods
}

func AddItemToInventory(cs *state.CharacterState, invItem defs.InventoryItem) (bool, defs.InventoryItem) {
	succ, remaining := item.AddItemToStandardInventory(&cs.StandardInventory, invItem)
	cs.Validate()
	return succ, remaining
}

func RemoveItemFromInventory(cs *state.CharacterState, itemToRemove defs.InventoryItem) (bool, defs.InventoryItem) {
	return item.RemoveItemFromStandardInventory(&cs.StandardInventory, itemToRemove)
}

// SpendMoney spends the given amount of money from the entity's coin purse and/or inventory
func SpendMoney(inv *defs.StandardInventory, value int, dataman *datamanager.DataManager) {
	if dataman == nil {
		logz.Panicln("SpendMoney", "dataman passed was nil")
	}
	// first, calculate our wallet
	wallet := map[int]int{}
	for _, coin := range append(inv.CoinPurse, inv.InventoryItems...) {
		if coin == nil {
			continue
		}
		if coin.Def.IsCurrencyItem() {
			val := coin.Def.GetValue()
			_, exists := wallet[val]
			if !exists {
				wallet[val] = 0
			}
			wallet[val] += coin.Quantity
		}
	}

	payment, success := bestPayment(value, wallet)
	if !success {
		// wallet doesn't have enough money for payment!
		panic("player tried to spend money he doesn't have... add checks at transaction location")
	}

	totalPaid := 0
	for denom, numCoins := range payment {
		fmt.Println("denom:", denom, "num:", numCoins)
		totalPaid += denom * numCoins
	}
	fmt.Println("total paid", totalPaid)

	if totalPaid < value {
		logz.Panicln("SpendMoney", "total payment is less than what you're supposed to pay! did bestPayment calculate wrongly?")
	}

	overpaid := totalPaid - value

	// remove payment coins and add change coins
	for denom, numCoins := range payment {
		if numCoins == 0 {
			continue
		}
		itemID := fmt.Sprintf("currency_value_%v", denom)
		coinsToRemove := dataman.NewInventoryItem(defs.ItemID(itemID), numCoins)
		success, remaining := item.RemoveItemFromStandardInventory(inv, coinsToRemove)
		if !success || remaining.Quantity != 0 {
			logz.Panicf("failed to pay all coins. remaining unpaid coins: %s", remaining.String())
		}
	}

	// if change was given, put it into the player's coin purse
	if overpaid > 0 {
		change := CalculateCoins(overpaid)
		for denom, numCoins := range change {
			if numCoins == 0 {
				continue
			}
			itemID := fmt.Sprintf("currency_value_%v", denom)
			success, _ := item.AddItemToStandardInventory(inv, dataman.NewInventoryItem(defs.ItemID(itemID), numCoins))
			if !success {
				fmt.Println("failed to add coin to inventory")
			}
		}
	}
}

func EarnMoney(inv *defs.StandardInventory, value int, dataman *datamanager.DataManager) {
	coins := CalculateCoins(value)
	for denom, numCoins := range coins {
		if numCoins == 0 {
			continue
		}
		itemID := fmt.Sprintf("currency_value_%v", denom)
		success, _ := item.AddItemToStandardInventory(inv, dataman.NewInventoryItem(defs.ItemID(itemID), numCoins))
		if !success {
			fmt.Println("failed to add coin to inventory")
		}
	}
}

func bestPayment(price int, wallet map[int]int) (map[int]int, bool) {
	totalValue := 0
	for d, c := range wallet {
		totalValue += d * c
	}
	if totalValue < price {
		return nil, false
	}

	denoms := []int{1000, 100, 50, 10, 5, 1}
	bestOverpay := math.MaxInt
	bestUsed := math.MaxInt
	var bestCombo map[int]int

	var dfs func(idx, paid, used int, combo map[int]int)
	dfs = func(idx, paid, used int, combo map[int]int) {
		if paid >= price {
			overpay := paid - price
			if overpay < bestOverpay || (overpay == bestOverpay && used < bestUsed) {
				bestOverpay = overpay
				bestUsed = used
				bestCombo = make(map[int]int)
				for k, v := range combo {
					bestCombo[k] = v
				}
			}
			return
		}

		if idx >= len(denoms) {
			return
		}

		denom := denoms[idx]
		for count := 0; count <= wallet[denom]; count++ {
			newPaid := paid + denom*count
			if newPaid > price+1000 { // pruning
				break
			}
			combo[denom] = count
			dfs(idx+1, newPaid, used+count, combo)
			delete(combo, denom)
		}
	}

	dfs(0, 0, 0, map[int]int{})
	return bestCombo, true
}

func CalculateCoins(value int) map[int]int {
	coins := map[int]int{}
	coins[1000] = 0
	coins[100] = 0
	coins[50] = 0
	coins[10] = 0
	coins[5] = 0
	coins[1] = 0

	if value >= 1000 {
		value1000 := value / 1000
		value -= value1000 * 1000
		coins[1000] = value1000
	}
	if value >= 100 {
		value100 := value / 100
		value -= value100 * 100
		coins[100] = value100
	}
	if value >= 50 {
		value50 := value / 50
		value -= value50 * 50
		coins[50] = value50
	}
	if value >= 10 {
		value10 := value / 10
		value -= value10 * 10
		coins[10] = value10
	}
	if value >= 5 {
		value5 := value / 5
		value -= value5 * 5
		coins[5] = value5
	}
	if value >= 1 {
		value1 := value
		value = 0
		coins[1] = value1
	}

	if value != 0 {
		logz.Panicln("CalculateCoins", "remaining value ended up not being zero... is the logic here broken? remaining value:", value)
	}

	return coins
}

// EquipItem equips a weapon, body armor, clothes, or other equipable items that go onto the entity's body or equipment slots
func EquipItem(cs *state.CharacterState, i defs.InventoryItem) (success bool) {
	i.Validate()
	if !i.Def.IsEquipable() {
		logz.Panicln(cs.DisplayName, "tried to equip an inequipable item:", i.Def.GetID())
	}

	switch i.Def.GetItemType() {
	case item.TypeHeadwear:
		if cs.Equipment.EquipedHeadwear != nil {
			// already equiped; remove it and put it in a regular inventory slot
			succ, _ := AddItemToInventory(cs, *cs.Equipment.EquipedHeadwear)
			if !succ {
				return false
			}
		}
		cs.Equipment.EquipedHeadwear = &i
	case item.TypeFootwear:
		if cs.Equipment.EquipedFootwear != nil {
			succ, _ := AddItemToInventory(cs, *cs.Equipment.EquipedFootwear)
			if !succ {
				return false
			}
		}
	case item.TypeBodywear:
		if cs.Equipment.EquipedBodywear != nil {
			// already equiped; remove it and put it in a regular inventory slot
			succ, _ := AddItemToInventory(cs, *cs.Equipment.EquipedBodywear)
			if !succ {
				return false
			}
		}
	case item.TypeWeapon:
		if cs.Equipment.EquipedWeapon != nil {
			succ, _ := AddItemToInventory(cs, *cs.Equipment.EquipedWeapon)
			if !succ {
				return false
			}
		}
	case item.TypeAuxiliary:
		if cs.Equipment.EquipedAuxiliary != nil {
			// already equiped; remove it and put it in a regular inventory slot
			succ, _ := AddItemToInventory(cs, *cs.Equipment.EquipedAuxiliary)
			if !succ {
				return false
			}
		}
	default:
		logz.Panicln(cs.DisplayName, "tried to equip item, but it's type didn't match in the switch statement... (this probably should be caught by the IsEquipable check)")
	}

	// cd.Validate()
	return true
}
