package definitions

import (
	"github.com/webbben/2d-game-engine/item"
)

// represents a shopkeeper; has all the items, the shopkeeper's money, etc.
type Shopkeeper struct {
	ShopInventory []*item.InventoryItem
	baseInventory []item.InventoryItem
	gold          int    // current amount of gold this shopkeeper has
	baseGold      int    // the base amount of gold this shopkeeper has
	shopName      string // the name of this shop; shows in the trade screen
}

func NewShopKeeper(baseGold int, shopName string, baseInventory []item.InventoryItem) Shopkeeper {
	sk := Shopkeeper{
		baseGold:      baseGold,
		gold:          baseGold,
		shopName:      shopName,
		baseInventory: baseInventory,
	}

	for _, inv := range baseInventory {
		sk.ShopInventory = append(sk.ShopInventory, &item.InventoryItem{
			Instance: inv.Instance,
			Def:      inv.Def,
			Quantity: inv.Quantity,
		})
	}

	sk.Validate()

	return sk
}

func (sk Shopkeeper) Validate() {
	if len(sk.baseInventory) == 0 {
		panic("base inventory is empty")
	}
	if len(sk.ShopInventory) == 0 {
		panic("shop inventory has no size")
	}
	for _, invItem := range sk.ShopInventory {
		if invItem == nil {
			continue
		}
		invItem.Validate()
	}
}

func (sk *Shopkeeper) SetShopInventory(newItems []*item.InventoryItem) {
	sk.ShopInventory = make([]*item.InventoryItem, 0)
	for _, newItem := range newItems {
		if newItem == nil {
			sk.ShopInventory = append(sk.ShopInventory, nil)
			continue
		}
		sk.ShopInventory = append(sk.ShopInventory, &item.InventoryItem{
			Instance: newItem.Instance,
			Def:      newItem.Def,
			Quantity: newItem.Quantity,
		})
	}
}
