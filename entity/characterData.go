package entity

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"os"
	"path/filepath"

	"github.com/webbben/2d-game-engine/definitions"
	"github.com/webbben/2d-game-engine/entity/body"
	"github.com/webbben/2d-game-engine/internal/config"
	"github.com/webbben/2d-game-engine/internal/logz"
	"github.com/webbben/2d-game-engine/item"
)

// CharacterData contains all the info and data about a character, excluding things like mechanics and flags, etc.
//
// Things like the character's identity (name, ID, etc), the character's body state (visible appearance, hair, eyes, etc),
// the items the character has in its inventory, etc. Basically, all the data needed to actually save and load this character.
// (For example, the character builder only saves the data in this struct, and doesn't refer to anything else in an entity)
//
// The inner mechanisms used for things like movement, combat, etc are not included here.
type CharacterData struct {
	// Name, Identity

	DisplayName string // the actual name of the entity, as displayed in game to players
	ID          string // the unique identifier of this entity (not usually seen by players - only by developers)
	IsPlayer    bool   `json:"-"` // flag indicating if this entity is the player

	// Inventory and items
	// Each InventoryItem slot represents an actual individual item slot in the inventory.
	// Equiped armor, weapon, rings, amulets, arrows, etc all have their own specific slots.
	// Other inventory items just go in the array.
	// Note: as of now, we are not doing the "top level scrollable items row" concept that you see in Minecraft or SDV.
	// Could implement that later, but for now there are just specific slots for equipped items, and then an array of items for the rest of your inventory.

	InventoryItems []*item.InventoryItem

	// items in the entity's coin purse. note that coins can also be put in regular inventory slots.
	// the coin purse is effectively disabled if its size is 0.
	CoinPurse []*item.InventoryItem

	EquipedHeadwear  *item.InventoryItem
	EquipedBodywear  *item.InventoryItem
	EquipedFootwear  *item.InventoryItem
	EquipedAmulet    *item.InventoryItem
	EquipedRing1     *item.InventoryItem
	EquipedRing2     *item.InventoryItem
	EquipedAmmo      *item.InventoryItem
	EquipedAuxiliary *item.InventoryItem
	EquipedWeapon    *item.InventoryItem

	// Body

	Body body.EntityBodySet

	// Attributes, Skills

	Vitals     Vitals
	Attributes Attributes

	WalkSpeed float64 `json:"walk_speed"` // value should be a TileSize / NumFrames calculation
	RunSpeed  float64 `json:"run_speed"`
}

func (cd CharacterData) WriteToJSON(outputFilePath string) error {
	if !filepath.IsAbs(outputFilePath) {
		return fmt.Errorf("given path is not abs (%s); please pass an absolute path", outputFilePath)
	}

	// Note: ItemDefs are not allowed to be written to JSONs since they can't be loaded from JSON properly.
	// That is handled by setting the Def field of inventoryItems to json:"-"

	data, err := json.MarshalIndent(cd, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	return os.WriteFile(outputFilePath, data, 0o644)
}

func LoadCharacterDataJSON(src string, defMgr *definitions.DefinitionManager) (CharacterData, error) {
	if !config.FileExists(src) {
		return CharacterData{}, errors.New("no file found at path: " + src)
	}

	data, err := os.ReadFile(src)
	if err != nil {
		return CharacterData{}, fmt.Errorf("failed to read file data: %w", err)
	}

	var cd CharacterData
	err = json.Unmarshal(data, &cd)
	if err != nil {
		return CharacterData{}, fmt.Errorf("failed to unmarshal data: %w", err)
	}

	// Load actual ItemDefs from DefinitionManager
	if cd.EquipedAmmo != nil {
		cd.EquipedAmmo.Def = defMgr.GetItemDef(cd.EquipedAmmo.Instance.DefID)
	}
	if cd.EquipedAmulet != nil {
		cd.EquipedAmulet.Def = defMgr.GetItemDef(cd.EquipedAmulet.Instance.DefID)
	}
	if cd.EquipedAuxiliary != nil {
		cd.EquipedAuxiliary.Def = defMgr.GetItemDef(cd.EquipedAuxiliary.Instance.DefID)
	}
	if cd.EquipedBodywear != nil {
		cd.EquipedBodywear.Def = defMgr.GetItemDef(cd.EquipedBodywear.Instance.DefID)
	}
	if cd.EquipedFootwear != nil {
		cd.EquipedFootwear.Def = defMgr.GetItemDef(cd.EquipedFootwear.Instance.DefID)
	}
	if cd.EquipedHeadwear != nil {
		cd.EquipedHeadwear.Def = defMgr.GetItemDef(cd.EquipedHeadwear.Instance.DefID)
	}
	if cd.EquipedRing1 != nil {
		cd.EquipedRing1.Def = defMgr.GetItemDef(cd.EquipedRing1.Instance.DefID)
	}
	if cd.EquipedRing2 != nil {
		cd.EquipedRing2.Def = defMgr.GetItemDef(cd.EquipedRing2.Instance.DefID)
	}
	if cd.EquipedWeapon != nil {
		cd.EquipedWeapon.Def = defMgr.GetItemDef(cd.EquipedWeapon.Instance.DefID)
	}

	for _, i := range cd.InventoryItems {
		if i == nil {
			continue
		}
		i.Def = defMgr.GetItemDef(i.Instance.DefID)
	}

	// Load body "skin" parts
	if cd.Body.BodySet.PartSrc.ID == "" {
		logz.Panicln(cd.DisplayName, "failed to load body set; id is empty")
	}
	cd.Body.BodySet.PartSrc = defMgr.GetBodyPartDef(cd.Body.BodySet.PartSrc.ID)
	if cd.Body.ArmsSet.PartSrc.ID == "" {
		logz.Panicln(cd.DisplayName, "failed to load arms set; id is empty")
	}
	cd.Body.ArmsSet.PartSrc = defMgr.GetBodyPartDef(cd.Body.ArmsSet.PartSrc.ID)
	if cd.Body.LegsSet.PartSrc.ID == "" {
		logz.Panicln(cd.DisplayName, "failed to load legs set; id is empty")
	}
	cd.Body.LegsSet.PartSrc = defMgr.GetBodyPartDef(cd.Body.LegsSet.PartSrc.ID)
	if cd.Body.EyesSet.PartSrc.ID == "" {
		logz.Panicln(cd.DisplayName, "failed to load eyes set; id is empty")
	}
	cd.Body.EyesSet.PartSrc = defMgr.GetBodyPartDef(cd.Body.EyesSet.PartSrc.ID)
	if cd.Body.HairSet.PartSrc.ID == "" {
		logz.Panicln(cd.DisplayName, "failed to load hair set; id is empty")
	}
	cd.Body.HairSet.PartSrc = defMgr.GetBodyPartDef(cd.Body.HairSet.PartSrc.ID)

	// Load equiped items
	if cd.EquipedHeadwear == nil {
		cd.Body.EquipHeadSet.PartSrc = body.SelectedPartDef{None: true}
	} else {
		cd.Body.EquipHeadSet.PartSrc = *cd.EquipedHeadwear.Def.GetBodyPartDef()
	}
	if cd.EquipedBodywear == nil {
		cd.Body.EquipBodySet.PartSrc = body.SelectedPartDef{None: true}
		cd.Body.EquipLegsSet.PartSrc = body.SelectedPartDef{None: true}
	} else {
		cd.Body.EquipBodySet.PartSrc = *cd.EquipedBodywear.Def.GetBodyPartDef()
		cd.Body.EquipLegsSet.PartSrc = *cd.EquipedBodywear.Def.GetLegsPartDef()
	}
	if cd.EquipedFootwear == nil {
		cd.Body.EquipFeetSet.PartSrc = body.SelectedPartDef{None: true}
	} else {
		cd.Body.EquipFeetSet.PartSrc = *cd.EquipedFootwear.Def.GetBodyPartDef()
	}
	if cd.EquipedAuxiliary == nil {
		cd.Body.AuxItemSet.PartSrc = body.SelectedPartDef{None: true}
	} else {
		cd.Body.AuxItemSet.PartSrc = *cd.EquipedAuxiliary.Def.GetBodyPartDef()
	}
	if cd.EquipedWeapon == nil {
		cd.Body.WeaponSet.PartSrc = body.SelectedPartDef{None: true}
		cd.Body.WeaponFxSet.PartSrc = body.SelectedPartDef{None: true}
	} else {
		weaponPart, fxPart := item.GetWeaponParts(cd.EquipedWeapon.Def)
		cd.Body.WeaponSet.PartSrc = weaponPart
		cd.Body.WeaponFxSet.PartSrc = fxPart
	}

	return cd, nil
}

// SetInventoryItems sets all the inventory items of an entity
func (cd *CharacterData) SetInventoryItems(invItems []*item.InventoryItem) {
	cd.InventoryItems = make([]*item.InventoryItem, 0)

	for _, newItem := range invItems {
		if newItem == nil {
			cd.InventoryItems = append(cd.InventoryItems, nil)
			continue
		}
		newItem.Validate()
		cd.InventoryItems = append(cd.InventoryItems, &item.InventoryItem{
			Instance: newItem.Instance,
			Def:      newItem.Def,
			Quantity: newItem.Quantity,
		})
	}
}

func (cd *CharacterData) SetCoinPurseItems(invItems []*item.InventoryItem) {
	cd.CoinPurse = make([]*item.InventoryItem, 0)

	for _, newItem := range invItems {
		if newItem == nil {
			cd.CoinPurse = append(cd.CoinPurse, nil)
			continue
		}
		if !newItem.Def.IsCurrencyItem() {
			logz.Panicln(cd.DisplayName, "SetCoinPurseItems: tried to add item to coin purse that is not a currency item:", newItem)
		}

		newItem.Validate()
		cd.CoinPurse = append(cd.CoinPurse, &item.InventoryItem{
			Instance: newItem.Instance,
			Def:      newItem.Def,
			Quantity: newItem.Quantity,
		})
	}
}

func (cd *CharacterData) AddItemToInventory(invItem item.InventoryItem) (bool, item.InventoryItem) {
	invItem.Validate()

	// if currency, first try to place it in the coin purse
	if len(cd.CoinPurse) > 0 {
		if invItem.Def.IsCurrencyItem() {
			success, remaining := item.AddItemToInventory(invItem, cd.CoinPurse)
			if success {
				return true, item.InventoryItem{}
			}
			invItem = remaining
		}
	}

	success, remaining := item.AddItemToInventory(invItem, cd.InventoryItems)
	cd.Validate()
	return success, remaining
}

func (cd *CharacterData) RemoveItemFromInventory(itemToRemove item.InventoryItem) (bool, item.InventoryItem) {
	itemToRemove.Validate()

	if itemToRemove.Def.IsCurrencyItem() && len(cd.CoinPurse) > 0 {
		// first try the coin purse
		success, remaining := item.RemoveItemFromInventory(itemToRemove, cd.CoinPurse)
		if success {
			if remaining.Quantity != 0 {
				logz.Panicln("RemoveItemFromInventory", "item removal was supposedly successful, but remaining is not 0:", remaining.Quantity)
			}
			return true, remaining
		}
		if remaining.Quantity == 0 {
			panic("why is Quantity 0 if no success?")
		}
		// if there is still some left to remove, use the `remaining` value
		// e.g. if some coins are not in the coin purse for some reason
		itemToRemove.Quantity = remaining.Quantity
	}

	success, remaining := item.RemoveItemFromInventory(itemToRemove, cd.InventoryItems)
	if success {
		if remaining.Quantity != 0 {
			logz.Panicln("RemoveItemFromInventory", "item removal was supposedly successful, but remaining is not 0:", remaining.Quantity)
		}
	}
	cd.Validate()
	return success, remaining
}

func (cd CharacterData) CountMoney() int {
	sum := 0
	for _, coinItem := range cd.CoinPurse {
		if coinItem == nil {
			continue
		}
		if coinItem.Def.IsCurrencyItem() {
			sum += coinItem.Def.GetValue() * coinItem.Quantity
		}
	}

	// also check for coins not in coin purse
	for _, coinItem := range cd.InventoryItems {
		if coinItem == nil {
			continue
		}
		if coinItem.Def.IsCurrencyItem() {
			sum += coinItem.Def.GetValue() * coinItem.Quantity
		}
	}

	return sum
}

// SpendMoney spends the given amount of money from the entity's coin purse and/or inventory
func (cd *CharacterData) SpendMoney(value int, defMgr *definitions.DefinitionManager) {
	if defMgr == nil {
		logz.Panicln(cd.DisplayName, "SpendMoney: defMgr passed was nil")
	}
	// first, calculate our wallet
	wallet := map[int]int{}
	for _, coin := range append(cd.CoinPurse, cd.InventoryItems...) {
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
		coinsToRemove := defMgr.NewInventoryItem(itemID, numCoins)
		success, remaining := cd.RemoveItemFromInventory(coinsToRemove)
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
			success, _ := cd.AddItemToInventory(defMgr.NewInventoryItem(itemID, numCoins))
			if !success {
				fmt.Println("failed to add coin to inventory")
			}
		}
	}
}

func (cd *CharacterData) EarnMoney(value int, defMgr *definitions.DefinitionManager) {
	coins := CalculateCoins(value)
	for denom, numCoins := range coins {
		if numCoins == 0 {
			continue
		}
		itemID := fmt.Sprintf("currency_value_%v", denom)
		success, _ := cd.AddItemToInventory(defMgr.NewInventoryItem(itemID, numCoins))
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
func (cd *CharacterData) EquipItem(i item.InventoryItem) (success bool) {
	i.Validate()
	if !i.Def.IsEquipable() {
		logz.Panicln(cd.DisplayName, "tried to equip an inequipable item:", i.Def.GetID())
	}

	switch i.Def.GetItemType() {
	case item.TypeHeadwear:
		if cd.EquipedHeadwear != nil {
			// already equiped; remove it and put it in a regular inventory slot
			succ, _ := cd.AddItemToInventory(*cd.EquipedHeadwear)
			if !succ {
				return false
			}
		}
		cd.EquipedHeadwear = &i
		part := i.Def.GetBodyPartDef()
		if part == nil {
			logz.Panicln(cd.DisplayName, "tried to equip an item with no part def:", i.Def.GetID())
		}
		cd.Body.SetEquipHead(*part)
	case item.TypeFootwear:
		if cd.EquipedFootwear != nil {
			succ, _ := cd.AddItemToInventory(*cd.EquipedFootwear)
			if !succ {
				return false
			}
		}
		cd.EquipedFootwear = &i
		part := i.Def.GetBodyPartDef()
		if part == nil {
			logz.Panicln(cd.DisplayName, "tried to equip an item with no part def:", i.Def.GetID())
		}
		cd.Body.SetEquipFeet(*part)
	case item.TypeBodywear:
		if cd.EquipedBodywear != nil {
			// already equiped; remove it and put it in a regular inventory slot
			succ, _ := cd.AddItemToInventory(*cd.EquipedBodywear)
			if !succ {
				return false
			}
		}
		cd.EquipedBodywear = &i
		bodyPart := i.Def.GetBodyPartDef()
		if bodyPart == nil {
			logz.Panicln(cd.DisplayName, "tried to equip an item with no part def:", i.Def.GetID())
		}
		legsPart := i.Def.GetLegsPartDef()
		if legsPart == nil {
			logz.Panicln(cd.DisplayName, "tried to equip bodywear with no legs part:", i.Def.GetID())
		}
		cd.Body.SetEquipBody(*bodyPart, *legsPart)
	case item.TypeWeapon:
		if cd.EquipedWeapon != nil {
			succ, _ := cd.AddItemToInventory(*cd.EquipedWeapon)
			if !succ {
				return false
			}
		}
		cd.EquipedWeapon = &i
		// GetWeaponParts handles panicking if anything is missing
		part, fxPart := item.GetWeaponParts(i.Def)

		cd.Body.SetWeapon(part, fxPart)

		// sanity check
		if cd.Body.WeaponSet.PartSrc.None {
			panic("equiped weapon, but weapon set partSrc is none")
		}
	case item.TypeAuxiliary:
		if cd.EquipedAuxiliary != nil {
			// already equiped; remove it and put it in a regular inventory slot
			succ, _ := cd.AddItemToInventory(*cd.EquipedAuxiliary)
			if !succ {
				return false
			}
		}
		cd.EquipedAuxiliary = &i
		part := i.Def.GetBodyPartDef()
		if part == nil {
			logz.Panicln(cd.DisplayName, "tried to equip an item with no part def:", i.Def.GetID())
		}
		cd.Body.SetAuxiliary(*part)
	default:
		logz.Panicln(cd.DisplayName, "tried to equip item, but it's type didn't match in the switch statement... (this probably should be caught by the IsEquipable check)")
	}

	cd.Validate()
	return true
}

func (cd *CharacterData) UnequipHeadwear() {
	if cd.EquipedHeadwear == nil {
		logz.Panicln(cd.DisplayName, "tried to unequip headwear, but equiped headwear is nil")
	}

	cd.EquipedHeadwear = nil
	cd.Body.EquipHeadSet.Remove()
	// reload hair too, since it may have been cropped by the previously equiped headwear
	cd.Body.ReloadHair()
}

func (cd *CharacterData) UnequipFootwear() {
	if cd.EquipedFootwear == nil {
		logz.Panicln(cd.DisplayName, "tried to unequip footwear, but equiped footwear is nil")
	}

	cd.EquipedFootwear = nil
	cd.Body.EquipFeetSet.Remove()
}

func (cd *CharacterData) UnequipBodywear() {
	if cd.EquipedBodywear == nil {
		logz.Panicln(cd.DisplayName, "tried to unequip bodywear, but equiped bodywear is nil")
	}

	cd.EquipedBodywear = nil
	cd.Body.EquipBodySet.Remove()
	cd.Body.EquipLegsSet.Remove()
	cd.Body.ReloadArms()
}

func (cd *CharacterData) UnequipAuxiliary() {
	if cd.EquipedAuxiliary == nil {
		logz.Panicln(cd.DisplayName, "tried to unequip auxiliary, but equiped auxiliary is nil")
	}

	cd.EquipedAuxiliary = nil
	cd.Body.RemoveAuxiliary()
}

func (cd *CharacterData) UnequipWeapon() {
	if cd.EquipedWeapon == nil {
		logz.Panicln(cd.DisplayName, "tried to unequip weapon, but equiped weapon is nil")
	}

	cd.EquipedWeapon = nil
	cd.Body.RemoveWeapon()
}
