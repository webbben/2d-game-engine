package player

import (
	"fmt"
	"math"

	"github.com/webbben/2d-game-engine/definitions"
	"github.com/webbben/2d-game-engine/entity"
	"github.com/webbben/2d-game-engine/entity/npc"
	"github.com/webbben/2d-game-engine/internal/logz"
	"github.com/webbben/2d-game-engine/item"
)

type Player struct {
	Entity *entity.Entity // the entity that represents this player in a map

	InventoryItems []*item.InventoryItem // the regular items (not equiped) that are in the player's inventory
	CoinPurse      []*item.InventoryItem // items in the player's coin purse. note that coins can also be put in regular inventory slots.

	// equiped items
	EquipedHeadwear  *item.InventoryItem
	EquipedBodywear  *item.InventoryItem
	EquipedFootwear  *item.InventoryItem
	EquipedAmulet    *item.InventoryItem
	EquipedRing1     *item.InventoryItem
	EquipedRing2     *item.InventoryItem
	EquipedAmmo      *item.InventoryItem
	EquipedAuxiliary *item.InventoryItem

	defMgr *definitions.DefinitionManager

	World WorldContext
}

type WorldContext interface {
	GetNearbyNPCs(x, y, radius float64) []*npc.NPC
}

// needed for sorting renderables
func (p Player) Y() float64 {
	return p.Entity.Y
}

func NewPlayer(defMgr *definitions.DefinitionManager) Player {
	return Player{
		InventoryItems: make([]*item.InventoryItem, 18),
		CoinPurse:      make([]*item.InventoryItem, 6),
		defMgr:         defMgr,
	}
}

func (p *Player) SetInventoryItems(invItems []*item.InventoryItem) {
	p.InventoryItems = make([]*item.InventoryItem, 0)

	for _, newItem := range invItems {
		if newItem == nil {
			p.InventoryItems = append(p.InventoryItems, nil)
			continue
		}
		p.InventoryItems = append(p.InventoryItems, &item.InventoryItem{
			Instance: newItem.Instance,
			Def:      newItem.Def,
			Quantity: newItem.Quantity,
		})
	}
}

func (p *Player) AddItemToInventory(invItem item.InventoryItem) (bool, item.InventoryItem) {
	// if currency, first try to place it in the coin purse
	if invItem.Def.IsCurrencyItem() {
		success, remaining := item.AddItemToInventory(invItem, p.CoinPurse)
		if success {
			return true, item.InventoryItem{}
		}
		invItem = remaining
	}

	return item.AddItemToInventory(invItem, p.InventoryItems)
}

func (p *Player) RemoveItemFromInventory(itemToRemove item.InventoryItem) (bool, item.InventoryItem) {
	if itemToRemove.Def.IsCurrencyItem() {
		// first try the coin purse
		success, remaining := item.RemoveItemFromInventory(itemToRemove, p.CoinPurse)
		if success {
			return true, remaining
		}
	}
	return item.RemoveItemFromInventory(itemToRemove, p.InventoryItems)
}

func (p Player) CountMoney() int {
	sum := 0
	for _, coinItem := range p.CoinPurse {
		if coinItem == nil {
			continue
		}
		if coinItem.Def.IsCurrencyItem() {
			sum += coinItem.Def.GetValue() * coinItem.Quantity
		}
	}

	// also check for coins not in coin purse
	for _, coinItem := range p.InventoryItems {
		if coinItem == nil {
			continue
		}
		if coinItem.Def.IsCurrencyItem() {
			sum += coinItem.Def.GetValue() * coinItem.Quantity
		}
	}

	return sum
}

// spends the given amount of money from the player's coin purse and/or inventory
func (p *Player) SpendMoney(value int) {
	// first, calculate our wallet
	wallet := map[int]int{}
	for _, coin := range append(p.CoinPurse, p.InventoryItems...) {
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
		totalPaid += denom * numCoins
	}
	overpaid := totalPaid - value

	// remove payment coins and add change coins
	for denom, numCoins := range payment {
		if numCoins == 0 {
			continue
		}
		itemID := fmt.Sprintf("currency_value_%v", denom)
		coinsToRemove := p.defMgr.NewInventoryItem(itemID, numCoins)
		success, remaining := p.RemoveItemFromInventory(coinsToRemove)
		if !success {
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
			success, _ := p.AddItemToInventory(p.defMgr.NewInventoryItem(itemID, numCoins))
			if !success {
				fmt.Println("failed to add coin to inventory")
			}
		}
	}
}

func (p *Player) EarnMoney(value int) {
	coins := CalculateCoins(value)
	for denom, numCoins := range coins {
		if numCoins == 0 {
			continue
		}
		itemID := fmt.Sprintf("currency_value_%v", denom)
		success, _ := p.AddItemToInventory(p.defMgr.NewInventoryItem(itemID, numCoins))
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

	return coins
}
