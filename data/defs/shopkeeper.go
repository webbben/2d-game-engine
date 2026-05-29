package defs

type ShopID string

// ShopkeeperDef is a base definition for a shopkeeper.
type ShopkeeperDef struct {
	ID            ShopID
	ShopName      string
	BaseInventory []ItemInitialStateDef
	BaseGold      int
}

func (sk ShopkeeperDef) Validate() {
	if sk.ID == "" {
		panic("shop has no ID")
	}
	if len(sk.BaseInventory) == 0 {
		panic("base inventory is empty")
	}
	if sk.ShopName == "" {
		panic("shop has no name")
	}
	if sk.BaseGold < 0 {
		panic("shop has negative gold")
	}
}
