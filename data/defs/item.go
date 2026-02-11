package defs

import (
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/webbben/2d-game-engine/internal/logz"
)

type (
	ItemType string
	ItemID   string
)

type ItemDef interface {
	GetID() ItemID          // the internal ID of this item.
	GetName() string        // the display name of this item.
	GetDescription() string // the description of the item
	GetValue() int          // the full value of this item, if it were sold at maximum price.
	GetWeight() float64     // the weight of this item, which factors into the player's inventory weight.
	GetMaxDurability() int  // the full durability of this item. a higher value means it takes longer to break.

	GetTileImg() *ebiten.Image        // gets the "tile image", i.e. the image used in places like the inventory slots
	GetEquipedTiles() []*ebiten.Image // gets the tiles for the equiped view of this item (if equipable)

	GetBodyPartDef() *SelectedPartDef // gets the body part set for wearing when this item is equiped, if it exists. Only exists for visible equipable items.
	GetLegsPartDef() *SelectedPartDef // gets a legs body part set, if it exists (should only exist for bodywear items)

	// if true, item can be grouped together with other same items in inventories
	// this tends to be true for items that don't have durability, like potions or arrows, but not for weapons and armor.
	IsGroupable() bool
	// if true, this item is treated as an equipable weapon that is held
	IsWeapon() bool
	// if true, this item is treated as an equipable piece of armor that is worn
	IsHeadwear() bool
	// if true, this item is equipable on the body
	IsBodywear() bool
	// if true, this item is equipable on the feet
	IsFootwear() bool
	// if true, this item is an amulet
	IsAmulet() bool
	// if true, this item is a ring
	IsRing() bool
	// if true, this item is treated as an equipable ammunition type (arrows, etc)
	IsAmmunition() bool
	// if true, this item is an "auxiliary" item which is held in the left hand (torches, shields, etc)
	IsAuxiliary() bool
	// if true, this item is treated as an item that can be consumed (food, potions, etc)
	IsConsumable() bool
	// if true, this item has no specific use or utility; it just exists in your inventory and may have value or weight
	IsMiscItem() bool
	// if true, this item is a piece of currency (gold, coins, etc) used in transactions
	IsCurrencyItem() bool
	// determines if this item can be equiped
	IsEquipable() bool

	GetItemType() ItemType // returns the item type; to simply confirm a single item type, use the specific Is<itemType> functions instead.

	Load() // load things like images

	Validate() // checks if item def is properly defined
}

type ItemInstance struct {
	DefID      ItemID // ID of the ItemDef that defines this item
	Durability int    // the current condition of this item
}

type InventoryItem struct {
	Instance ItemInstance
	Def      ItemDef `json:"-"` // since this interface can't be properly loaded from JSON, lets exclude it from JSONs
	Quantity int
}

type StandardInventory struct {
	CoinPurse      []*InventoryItem
	InventoryItems []*InventoryItem
	Equipment      EquipedItems
}

func (inv StandardInventory) Validate() {
	for _, i := range inv.CoinPurse {
		if i != nil {
			i.Validate()
		}
	}
	for _, i := range inv.InventoryItems {
		if i != nil {
			i.Validate()
		}
	}

	allEquipment := []*InventoryItem{}
	allEquipment = append(allEquipment, inv.Equipment.EquipedBodywear)
	allEquipment = append(allEquipment, inv.Equipment.EquipedHeadwear)
	allEquipment = append(allEquipment, inv.Equipment.EquipedFootwear)
	allEquipment = append(allEquipment, inv.Equipment.EquipedAuxiliary)
	allEquipment = append(allEquipment, inv.Equipment.EquipedWeapon)
	allEquipment = append(allEquipment, inv.Equipment.EquipedRing1)
	allEquipment = append(allEquipment, inv.Equipment.EquipedRing2)
	allEquipment = append(allEquipment, inv.Equipment.EquipedAmulet)
	allEquipment = append(allEquipment, inv.Equipment.EquipedAmmo)

	for _, i := range allEquipment {
		if i != nil {
			i.Validate()
		}
	}
}

func (inv StandardInventory) CountMoney() int {
	sum := 0
	for _, coinItem := range inv.CoinPurse {
		if coinItem == nil {
			continue
		}
		if coinItem.Def.IsCurrencyItem() {
			sum += coinItem.Def.GetValue() * coinItem.Quantity
		}
	}

	// also check for coins not in coin purse
	for _, coinItem := range inv.InventoryItems {
		if coinItem == nil {
			continue
		}
		if coinItem.Def.IsCurrencyItem() {
			sum += coinItem.Def.GetValue() * coinItem.Quantity
		}
	}

	return sum
}

func (inv *StandardInventory) SetCoinPurseItems(invItems []*InventoryItem) {
	inv.CoinPurse = make([]*InventoryItem, 0)

	for _, newItem := range invItems {
		if newItem == nil {
			inv.CoinPurse = append(inv.CoinPurse, nil)
			continue
		}
		if !newItem.Def.IsCurrencyItem() {
			logz.Panicln("SetCoinPurseItems", "tried to add item to coin purse that is not a currency item:", newItem)
		}

		newItem.Validate()
		inv.CoinPurse = append(inv.CoinPurse, &InventoryItem{
			Instance: newItem.Instance,
			Def:      newItem.Def,
			Quantity: newItem.Quantity,
		})
	}
}

func (inv *StandardInventory) SetInventoryItems(invItems []*InventoryItem) {
	inv.InventoryItems = make([]*InventoryItem, 0)

	for _, newItem := range invItems {
		if newItem == nil {
			inv.InventoryItems = append(inv.InventoryItems, nil)
			continue
		}
		newItem.Validate()
		inv.InventoryItems = append(inv.InventoryItems, &InventoryItem{
			Instance: newItem.Instance,
			Def:      newItem.Def,
			Quantity: newItem.Quantity,
		})
	}
}

func (invItem *InventoryItem) String() string {
	return fmt.Sprintf("{DefID: %s, Name: %s, Quant: %v}", invItem.Instance.DefID, invItem.Def.GetName(), invItem.Quantity)
}

// NewInventoryItem puts the pieces of an inventoryItem together; use the one in definitionManager to actually get an inventory item from the game state.
func NewInventoryItem(def ItemDef, quantity int) InventoryItem {
	if def == nil {
		panic("def is nil")
	}
	invItem := InventoryItem{
		Instance: ItemInstance{
			DefID:      def.GetID(),
			Durability: def.GetMaxDurability(),
		},
		Def:      def,
		Quantity: quantity,
	}

	invItem.Validate()
	return invItem
}

func InventoryToString(inv []*InventoryItem) string {
	s := ""
	for _, invItem := range inv {
		if invItem == nil {
			continue
		}
		s += invItem.String() + ", "
	}
	return s
}

func (invItem InventoryItem) Validate() {
	if invItem.Def == nil {
		panic("item def is nil")
	}
	if invItem.Instance.DefID == "" {
		panic("item instance has no def ID")
	}
	if invItem.Quantity <= 0 {
		panic("item quantity is less than or equal to 0. all items must have a quantity of at least 1")
	}
	if invItem.Def.GetID() != invItem.Instance.DefID {
		panic("def.GetID() does not match instance.defID")
	}
}

type EquipedItems struct {
	EquipedHeadwear  *InventoryItem
	EquipedBodywear  *InventoryItem
	EquipedFootwear  *InventoryItem
	EquipedAmulet    *InventoryItem
	EquipedRing1     *InventoryItem
	EquipedRing2     *InventoryItem
	EquipedAmmo      *InventoryItem
	EquipedAuxiliary *InventoryItem
	EquipedWeapon    *InventoryItem
}
