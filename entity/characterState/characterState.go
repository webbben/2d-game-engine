// Package characterstate defines everything that goes into a characters (npc, or player) state.
package characterstate

import (
	"fmt"
	"maps"
	"math"
	"slices"
	"strings"

	"github.com/webbben/2d-game-engine/clock"
	"github.com/webbben/2d-game-engine/data/datamanager"
	"github.com/webbben/2d-game-engine/data/defs"
	"github.com/webbben/2d-game-engine/data/id"
	"github.com/webbben/2d-game-engine/data/state"
	"github.com/webbben/2d-game-engine/item"
	"github.com/webbben/2d-game-engine/logz"
	"github.com/webbben/2d-game-engine/pubsub"
	"github.com/webbben/2d-game-engine/utils"
)

func GenerateUniquePlayerID(displayName string) defs.UniquePlayerID {
	displayName = strings.TrimSpace(displayName)
	displayName = strings.ReplaceAll(displayName, " ", "_")
	id := fmt.Sprintf("%s_%s", displayName, utils.GenerateUUID()[:8])
	playerID := defs.UniquePlayerID(id)
	ValidateUniquePlayerID(playerID)
	return playerID
}

func ValidateUniquePlayerID(uniquePlayerID defs.UniquePlayerID) {
	s := string(uniquePlayerID)
	if s == "" {
		logz.Panicln("UniquePlayerID", "UniquePlayerID was empty! This should've been set at the start of a new game...")
	}
	if strings.Contains(s, " ") {
		logz.Panicln("UniquePlayerID", "UniquePlayerID had spaces in it! it should use other characters like underscores or dashes instead.")
	}
	if strings.Contains(s, "/") {
		logz.Panicln("UniquePlayerID", "UniquePlayerID had slashes in it! these must not be used, to avoid confusing the filepaths.")
	}
}

// GetNetTraitModifiers returns all of the net modifiers on skills produced by the given traits
func GetNetTraitModifiers(traits []defs.TraitID, dataman *datamanager.DataManager) (skillMods map[defs.SkillID]int, attrMods map[defs.AttributeID]int) {
	if dataman == nil {
		panic("dataman was nil")
	}
	skillMods = make(map[defs.SkillID]int)
	attrMods = make(map[defs.AttributeID]int)

	for _, traitID := range traits {
		trait := dataman.GetTraitDef(traitID)
		for id, change := range trait.SkillChanges {
			if _, exists := skillMods[id]; !exists {
				skillMods[id] = 0
			}
			skillMods[id] += change
		}
		for id, change := range trait.AttributeChanges {
			if _, exists := attrMods[id]; !exists {
				attrMods[id] = 0
			}
			attrMods[id] += change
		}
	}

	return skillMods, attrMods
}

func AddItemToInventory(cs *state.CharacterState, invItem state.ItemState, dataman *datamanager.DataManager) (bool, state.ItemState) {
	succ, remaining := item.AddItemToStandardInventory(&cs.StandardInventory, invItem, dataman)
	cs.Validate()
	return succ, remaining
}

func RemoveItemFromInventory(cs *state.CharacterState, itemToRemove state.ItemState, dataman *datamanager.DataManager) (bool, state.ItemState) {
	return item.RemoveItemFromStandardInventory(&cs.StandardInventory, itemToRemove, dataman)
}

// SpendMoney spends the given amount of money from the entity's coin purse and/or inventory
func SpendMoney(inv *state.StandardInventory, value int, dataman *datamanager.DataManager) {
	if dataman == nil {
		logz.Panicln("SpendMoney", "dataman passed was nil")
	}
	// first, calculate our wallet
	// TODO: shouldn't we use CountMoney here??
	wallet := map[int]int{}
	for _, coin := range append(inv.CoinPurse, inv.InventoryItems...) {
		if coin == nil {
			continue
		}
		itemDef := dataman.GetItemDef(coin.DefID)
		if itemDef.Type == defs.TypeCurrency {
			val := itemDef.Value
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
		coinsToRemove := dataman.NewItemState(defs.ItemID(itemID), numCoins)
		success, remaining := item.RemoveItemFromStandardInventory(inv, *coinsToRemove, dataman)
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
			coinItem := dataman.NewItemState(defs.ItemID(itemID), numCoins)
			success, _ := item.AddItemToStandardInventory(inv, *coinItem, dataman)
			if !success {
				fmt.Println("failed to add coin to inventory")
			}
		}
	}
}

func EarnMoney(inv *state.StandardInventory, value int, dataman *datamanager.DataManager) {
	coins := CalculateCoins(value)
	for denom, numCoins := range coins {
		if numCoins == 0 {
			continue
		}
		itemID := fmt.Sprintf("currency_value_%v", denom)
		coinItem := dataman.NewItemState(defs.ItemID(itemID), numCoins)
		success, _ := item.AddItemToStandardInventory(inv, *coinItem, dataman)
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
				maps.Copy(bestCombo, combo)
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
// TODO: I don't think this is used anywhere... should we delete?
func EquipItem(cs *state.CharacterState, i *state.ItemState, dataman *datamanager.DataManager) (success bool) {
	if i == nil {
		panic("item state was nil")
	}
	i.Validate()

	itemDef := dataman.GetItemDef(i.DefID)

	if !itemDef.IsEquipable() {
		logz.Panicln(cs.DisplayName, "tried to equip an inequipable item:", i.DefID)
	}

	switch itemDef.Type {
	case defs.TypeHeadwear:
		if cs.EquipedHeadwear != nil {
			// already equiped; remove it and put it in a regular inventory slot
			succ, _ := AddItemToInventory(cs, *cs.EquipedHeadwear, dataman)
			if !succ {
				return false
			}
		}
		cs.EquipedHeadwear = i
	case defs.TypeFootwear:
		if cs.EquipedFootwear != nil {
			succ, _ := AddItemToInventory(cs, *cs.EquipedFootwear, dataman)
			if !succ {
				return false
			}
		}
		cs.EquipedFootwear = i
	case defs.TypeBodywear:
		if cs.EquipedBodywear != nil {
			// already equiped; remove it and put it in a regular inventory slot
			succ, _ := AddItemToInventory(cs, *cs.EquipedBodywear, dataman)
			if !succ {
				return false
			}
		}
		cs.EquipedBodywear = i
	case defs.TypeWeapon:
		if cs.EquipedWeapon != nil {
			succ, _ := AddItemToInventory(cs, *cs.EquipedWeapon, dataman)
			if !succ {
				return false
			}
		}
		cs.EquipedWeapon = i
	case defs.TypeAuxiliary:
		if cs.EquipedAuxiliary != nil {
			// already equiped; remove it and put it in a regular inventory slot
			succ, _ := AddItemToInventory(cs, *cs.EquipedAuxiliary, dataman)
			if !succ {
				return false
			}
		}
		cs.EquipedAuxiliary = i
	default:
		logz.Panicln(cs.DisplayName, "tried to equip item, but it's type didn't match in the switch statement... (this probably should be caught by the IsEquipable check)")
	}

	return true
}

func IsItemEquipped(itemID defs.ItemID, charState state.CharacterState) bool {
	if itemID == "" {
		panic("itemID was empty")
	}

	if charState.EquipedHeadwear != nil && charState.EquipedHeadwear.DefID == itemID {
		return true
	}
	if charState.EquipedBodywear != nil && charState.EquipedBodywear.DefID == itemID {
		return true
	}
	if charState.EquipedFootwear != nil && charState.EquipedFootwear.DefID == itemID {
		return true
	}
	if charState.EquipedWeapon != nil && charState.EquipedWeapon.DefID == itemID {
		return true
	}
	if charState.EquipedAuxiliary != nil && charState.EquipedAuxiliary.DefID == itemID {
		return true
	}
	if charState.EquipedAmmo != nil && charState.EquipedAmmo.DefID == itemID {
		return true
	}
	if charState.EquipedRing1 != nil && charState.EquipedRing1.DefID == itemID {
		return true
	}
	if charState.EquipedRing2 != nil && charState.EquipedRing2.DefID == itemID {
		return true
	}
	if charState.EquipedAmulet != nil && charState.EquipedAmulet.DefID == itemID {
		return true
	}
	return false
}

func GetLockIDs(charState state.CharacterState, dataman *datamanager.DataManager) []string {
	// make a map, since there could be multiple keys that have the same lock IDs in them.
	seen := map[string]bool{}
	lockIDs := []string{}

	for _, is := range charState.InventoryItems {
		if is == nil {
			continue
		}

		itemDef := dataman.GetItemDef(is.DefID)

		if itemDef.Type == defs.TypeKey {
			for _, id := range itemDef.LockIDs {
				if seen[id] {
					continue
				}
				seen[id] = true
				lockIDs = append(lockIDs, id)
			}
		}
	}

	return lockIDs
}

func AddKnowledge(topicID defs.TopicID, dataman *datamanager.DataManager, eventBus *pubsub.EventBus) {
	playerState := dataman.GetCharacterState(id.CharacterStateID(defs.PlayerID))
	if _, exists := playerState.Knowledge[topicID]; exists {
		// knowledge already exists, so do nothing
		logz.Println("AddKnowledge", "topic already known:", topicID)
		return
	}

	logz.Warnln("AddKnowledge", "New knowledge topic:", topicID)

	playerState.Knowledge[topicID] = true

	topicDef := dataman.GetDialogTopic(topicID)

	eventBus.Publish(defs.Event{
		Type: pubsub.EventNewKnowledgeTopic,
		Data: map[string]any{
			"topicID":          topicID,
			"topicDisplayName": topicDef.Prompt,
		},
	})
}

func ActivateItem(itemState *state.ItemState, dataman *datamanager.DataManager, eventBus *pubsub.EventBus) {
	if itemState == nil {
		logz.Panic("itemState was nil")
	}
	itemDef := dataman.GetItemDef(itemState.DefID)

	switch itemDef.Type {
	case defs.TypeBook:
		if itemDef.BookID == "" {
			logz.Panicln("ActivateItem", "book item didn't have a bookID:", itemState.DefID)
		}
		bookDef := dataman.GetBookDef(itemDef.BookID)
		for _, topic := range bookDef.KnowledgeTopics {
			AddKnowledge(topic, dataman, eventBus)
		}
	default:
		logz.Warnln("ActivateItem", "item was activated, but no logic is assigned to its type. item ID:", itemState.DefID, "item type:", itemDef.Type)
	}
}

func CalculateOpinion(opinionHolder, subject *state.CharacterState, currentTime clock.GameTime, dataman *datamanager.DataManager) ([]defs.OpinionModifier, int) {
	mods := []defs.OpinionModifier{}
	opinion := 0

	// just to be safe, since some old character states from old load files might not have this
	if opinionHolder.OpinionMods == nil {
		opinionHolder.OpinionMods = make(map[id.CharacterStateID][]defs.OpinionModifier)
	}
	if subject.OpinionMods == nil {
		subject.OpinionMods = make(map[id.CharacterStateID][]defs.OpinionModifier)
	}

	// first, count up existing modifiers if they exist
	if opinionMods, exists := opinionHolder.OpinionMods[subject.ID]; exists {
		activeOpinionMods := []defs.OpinionModifier{}

		for _, opinionMod := range opinionMods {
			// first, make sure this opinion mod isn't expired
			if opinionMod.Until != nil {
				if currentTime.IsAfter(*opinionMod.Until) {
					// this one is already expired, so skip it and remove it from character state opinion mods
					continue
				}
			}
			activeOpinionMods = append(activeOpinionMods, opinionMod)

			mods = append(mods, opinionMod)
			opinion += opinionMod.Mod
		}

		opinionHolder.OpinionMods[subject.ID] = activeOpinionMods
	}

	// next, check traits for opinion mods
	// trait opinion mods are based on what the subject has; so we look at the subject's traits, not our own
	for _, traitID := range subject.Traits {
		traitDef := dataman.GetTraitDef(traitID)

		if traitDef.SameTraitOpinionModifier.Mod != 0 {
			if slices.Contains(opinionHolder.Traits, traitID) {
				// both characters share this trait; apply same trait mod
				mod := traitDef.SameTraitOpinionModifier
				mods = append(mods, mod)
				opinion += mod.Mod
			}
		}

		if traitDef.OtherTraitOpinionModifiers != nil {
			for _, myTraitID := range opinionHolder.Traits {
				mod := traitDef.OtherTraitOpinionModifiers[myTraitID]
				mods = append(mods, mod)
				opinion += mod.Mod
			}
		}
	}

	// check for culture opinion mods
	holderCharDef := dataman.GetCharacterDef(opinionHolder.DefID)
	cultureDef := dataman.GetCultureDef(holderCharDef.CultureID)
	subjectCharDef := dataman.GetCharacterDef(subject.DefID)
	subjectCultureDef := dataman.GetCultureDef(subjectCharDef.CultureID)

	if cultureDef.ID == subjectCultureDef.ID {
		// TODO: should this go somewhere in the culture def or config? probably shouldn't be hardcoded here
		mods = append(mods, defs.OpinionModifier{
			Mod:    20,
			Reason: "Same culture",
		})
		opinion += 20
	} else {
		if mod, exists := cultureDef.OtherCultureOpinions[subjectCultureDef.ID]; exists {
			mods = append(mods, mod)
			opinion += mod.Mod
		}
	}

	return mods, opinion
}

func AddOpinionModifier(holder, subject id.CharacterStateID, mod defs.OpinionModifier, dataman *datamanager.DataManager) {
	holderState := dataman.GetCharacterState(holder)
	if holderState.OpinionMods == nil {
		holderState.OpinionMods = make(map[id.CharacterStateID][]defs.OpinionModifier)
	}
	holderState.OpinionMods[subject] = append(holderState.OpinionMods[subject], mod)
}

func CalculateSkillsAndAttributes(charStateID id.CharacterStateID, dataman *datamanager.DataManager) (skills map[defs.SkillID]int, attrs map[defs.AttributeID]int) {
	// get base skill levels
	characterState := dataman.GetCharacterState(charStateID)

	skillLevels := maps.Clone(characterState.BaseSkills)
	attrLevels := maps.Clone(characterState.BaseAttributes)

	// factor in trait modifiers
	for _, traitID := range characterState.Traits {
		traitDef := dataman.GetTraitDef(traitID)
		for attrID, mod := range traitDef.AttributeChanges {
			attrLevels[attrID] += mod
		}
		for skillID, mod := range traitDef.SkillChanges {
			skillLevels[skillID] += mod
		}
	}

	// factor in culture modifiers
	charDef := dataman.GetCharacterDef(characterState.DefID)
	if charDef.CultureID != "" {
		cultureDef := dataman.GetCultureDef(charDef.CultureID)
		for attrID, mod := range cultureDef.AttrMods {
			attrLevels[attrID] += mod
		}
		for skillID, mod := range cultureDef.SkillMods {
			skillLevels[skillID] += mod
		}
	}

	return skillLevels, attrLevels
}
