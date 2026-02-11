package defs

type ShopID string

// ShopkeeperDef is a base definition for a shopkeeper.
type ShopkeeperDef struct {
	ID            ShopID
	ShopName      string
	BaseInventory []InventoryItem
	BaseGold      int
}

func (sk ShopkeeperDef) Validate() {
	if len(sk.BaseInventory) == 0 {
		panic("base inventory is empty")
	}
	if len(sk.BaseInventory) == 0 {
		panic("shop inventory has no size")
	}
	for _, invItem := range sk.BaseInventory {
		invItem.Validate()
	}
}
