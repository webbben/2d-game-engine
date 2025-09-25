package item

import "time"

type ItemDef interface {
	GetID() string         // the internal ID of this item.
	GetName() string       // the display name of this item.
	GetValue() int         // the full value of this item, if it were sold at maximum price.
	GetWeight() float64    // the weight of this item, which factors into the player's inventory weight.
	GetMaxDurability() int // the full durability of this item. a higher value means it takes longer to break.
}

// includes the basic functions required for an item to be recognizable under the ItemDef interface.
// embed into a struct to make it effectively an item type.
type ItemBase struct {
	ID            string
	Name          string
	Value         int
	Weight        float64
	MaxDurability int
}

func (ib ItemBase) GetID() string {
	return ib.ID
}
func (ib ItemBase) GetName() string {
	return ib.Name
}
func (ib ItemBase) GetValue() int {
	return ib.Value
}
func (ib ItemBase) GetWeight() float64 {
	return ib.Weight
}
func (ib ItemBase) GetMaxDurability() int {
	return ib.MaxDurability
}

// just for the compiler to automatically confirm each item type implements ItemDef interface
func _confirmItemTypes() {
	_ = append([]ItemDef{}, WeaponDef{}, PotionDef{}, ArmorDef{})
}

type ItemState struct {
	DefID      string // ID of the ItemDef that defines this item
	Durability int    // the current condition of this item
}

type WeaponDef struct {
	ItemBase
	Damage        int     // damage per attack
	HitsPerSecond float64 // speed of attacks, in terms of number of attacks possible per second
}

type ArmorDef struct {
	ItemBase
	Protection int // amount of protection this piece of armor gives
}

type PotionDef struct {
	ItemBase
	EffectDuration time.Duration
	// TODO: add an Effect concept, which will encompass potion effects and enchantments on weapons or items
	// for now, just going to make a "heal amount" value.
	HealAmount int // how much health will be healed per second
}
