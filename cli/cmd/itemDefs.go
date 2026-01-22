package cmd

import (
	"github.com/webbben/2d-game-engine/entity/body"
	"github.com/webbben/2d-game-engine/internal/tiled"
	"github.com/webbben/2d-game-engine/item"
)

// At some point, these will probably be moved into the actual game repo. For now, defining here for testing.

func offsetInts(input []int, offset int) []int {
	newSlice := []int{}

	for _, v := range input {
		newSlice = append(newSlice, v+offset)
	}

	return newSlice
}

func GetItemDefs() []item.ItemDef {
	equipBodyTileset := "entities/parts/human_entity_parts.tsj"
	equipHeadTileset := "items/equiped_head_01.tsj"
	equipFeetTileset := "items/footwear.tsj"
	equipWeaponTileset := "items/weapon_frames.tsj"
	weaponFxTileset := "items/weapon_fx_frames.tsj"
	auxTileset := "items/equiped_aux.tsj"

	equipBodyOptions := []equipBodyOption{}
	for i := range 2 {
		// first equipment item starts at row 6; each item's first row is the body equipment, second row is legs equipment.
		offset := 73 * ((i * 2) + 5)
		bodyDef := body.NewPartDef(body.PartDefParams{
			Idle: &body.AnimationParams{
				TilesetSrc: equipBodyTileset,
				TilesLeft:  offsetInts([]int{37}, offset),
				AuxLeft:    offsetInts([]int{38}, offset),
				TilesRight: offsetInts([]int{19}, offset),
				AuxRight:   offsetInts([]int{20}, offset),
				TilesUp:    offsetInts([]int{55}, offset),
				AuxUp:      offsetInts([]int{56}, offset),
				TilesDown:  offsetInts([]int{0}, offset),
				AuxDown:    offsetInts([]int{1}, offset),
			},
			Walk: &body.AnimationParams{
				TilesetSrc: equipBodyTileset,
				TilesDown:  offsetInts([]int{4, 0, 6, 0}, offset),
				AuxDown:    offsetInts([]int{4, 1, 6, 1}, offset),
				TilesRight: offsetInts([]int{23, 19, 25, 19}, offset),
				AuxRight:   offsetInts([]int{23, 20, 25, 20}, offset),
				TilesLeft:  offsetInts([]int{41, 37, 43, 37}, offset),
				AuxLeft:    offsetInts([]int{41, 38, 43, 38}, offset),
				TilesUp:    offsetInts([]int{59, 55, 61, 55}, offset),
				AuxUp:      offsetInts([]int{59, 56, 61, 56}, offset),
			},
			Run: &body.AnimationParams{
				TilesetSrc: equipBodyTileset,
				TilesDown:  offsetInts([]int{3, 4, 0, 5, 6, 0}, offset),
				AuxDown:    offsetInts([]int{3, 4, 1, 5, 6, 1}, offset),
				TilesRight: offsetInts([]int{22, 23, 19, 24, 25, 19}, offset),
				AuxRight:   offsetInts([]int{22, 23, 20, 24, 25, 20}, offset),
				TilesLeft:  offsetInts([]int{40, 41, 37, 42, 43, 37}, offset),
				AuxLeft:    offsetInts([]int{40, 41, 38, 42, 43, 38}, offset),
				TilesUp:    offsetInts([]int{58, 59, 55, 60, 61, 55}, offset),
				AuxUp:      offsetInts([]int{58, 59, 56, 60, 61, 56}, offset),
			},
			Slash: &body.AnimationParams{
				TilesetSrc: equipBodyTileset,
				TilesDown:  offsetInts([]int{7, 8, 9, 9}, offset),
				TilesRight: offsetInts([]int{26, 27, 28, 28}, offset),
				TilesLeft:  offsetInts([]int{44, 45, 46, 46}, offset),
				TilesUp:    offsetInts([]int{62, 63, 64, 64}, offset),
			},
			Backslash: &body.AnimationParams{
				TilesetSrc: equipBodyTileset,
				TilesDown:  offsetInts([]int{10, 9, 8, 7}, offset),
				TilesRight: offsetInts([]int{28, 27, 26, 26}, offset),
				TilesLeft:  offsetInts([]int{46, 45, 44, 44}, offset),
				TilesUp:    offsetInts([]int{64, 63, 62, 62}, offset),
			},
			Shield: &body.AnimationParams{
				TilesetSrc: equipBodyTileset,
				TilesDown:  offsetInts([]int{11}, offset),
				TilesRight: offsetInts([]int{29}, offset),
				TilesLeft:  offsetInts([]int{47}, offset),
				TilesUp:    offsetInts([]int{65}, offset),
			},
		})

		offset = 73 * ((i * 2) + 6)

		legsDef := body.NewPartDef(body.PartDefParams{
			Idle: &body.AnimationParams{
				TilesetSrc: equipBodyTileset,
				TilesDown:  offsetInts([]int{0}, offset),
				TilesRight: offsetInts([]int{19}, offset),
				TilesLeft:  offsetInts([]int{37}, offset),
				TilesUp:    offsetInts([]int{55}, offset),
			},
			Walk: &body.AnimationParams{
				TilesetSrc: equipBodyTileset,
				TilesDown:  offsetInts([]int{4, 0, 6, 0}, offset),
				TilesRight: offsetInts([]int{23, 19, 25, 19}, offset),
				TilesLeft:  offsetInts([]int{41, 37, 43, 37}, offset),
				TilesUp:    offsetInts([]int{59, 55, 61, 55}, offset),
			},
			Run: &body.AnimationParams{
				TilesetSrc: equipBodyTileset,
				TilesDown:  offsetInts([]int{3, 4, 0, 5, 6, 0}, offset),
				TilesRight: offsetInts([]int{22, 23, 19, 24, 25, 19}, offset),
				TilesLeft:  offsetInts([]int{40, 41, 37, 42, 43, 37}, offset),
				TilesUp:    offsetInts([]int{58, 59, 55, 60, 61, 55}, offset),
			},
			Slash: &body.AnimationParams{
				TilesetSrc: equipBodyTileset,
				TilesDown:  offsetInts([]int{7, 7, 7, 7}, offset),
				TilesRight: offsetInts([]int{23, 26, 26, 26}, offset),
				TilesLeft:  offsetInts([]int{43, 44, 44, 44}, offset),
				TilesUp:    offsetInts([]int{62, 62, 62, 62}, offset),
			},
			Backslash: &body.AnimationParams{
				TilesetSrc: equipBodyTileset,
				TilesDown:  offsetInts([]int{7, 7, 7, 7}, offset),
				TilesRight: offsetInts([]int{26, 26, 23, 23}, offset),
				TilesLeft:  offsetInts([]int{44, 44, 43, 43}, offset),
				TilesUp:    offsetInts([]int{62, 62, 62, 62}, offset),
			},
			Shield: &body.AnimationParams{
				TilesetSrc: equipBodyTileset,
				TilesDown:  offsetInts([]int{7}, offset),
				TilesRight: offsetInts([]int{26}, offset),
				TilesLeft:  offsetInts([]int{44}, offset),
				TilesUp:    offsetInts([]int{62}, offset),
			},
		})
		defs := equipBodyOption{
			bodyDef: bodyDef,
			legsDef: legsDef,
		}
		equipBodyOptions = append(equipBodyOptions, defs)
	}

	equipFeetOptions := []body.SelectedPartDef{}
	for i := range 1 {
		offset := (i + 1) * 32
		def := body.NewPartDef(body.PartDefParams{
			Idle: &body.AnimationParams{
				TilesetSrc: equipFeetTileset,
				TilesDown:  offsetInts([]int{0}, offset),
				TilesRight: offsetInts([]int{8}, offset),
				TilesLeft:  offsetInts([]int{16}, offset),
				TilesUp:    offsetInts([]int{24}, offset),
			},
			Walk: &body.AnimationParams{
				TilesetSrc: equipFeetTileset,
				TilesDown:  offsetInts([]int{4, 0, 6, 0}, offset),
				TilesRight: offsetInts([]int{12, 8, 14, 8}, offset),
				TilesLeft:  offsetInts([]int{20, 16, 22, 16}, offset),
				TilesUp:    offsetInts([]int{28, 24, 30, 24}, offset),
			},
			Run: &body.AnimationParams{
				TilesetSrc: equipFeetTileset,
				TilesDown:  offsetInts([]int{3, 4, 0, 5, 6, 0}, offset),
				TilesRight: offsetInts([]int{11, 12, 8, 13, 14, 8}, offset),
				TilesLeft:  offsetInts([]int{19, 20, 16, 21, 22, 16}, offset),
				TilesUp:    offsetInts([]int{27, 28, 24, 29, 30, 24}, offset),
			},
			Slash: &body.AnimationParams{
				TilesetSrc: equipFeetTileset,
				TilesDown:  offsetInts([]int{7, 7, 7, 7}, offset),
				TilesRight: offsetInts([]int{14, 15, 15, 15}, offset),
				TilesLeft:  offsetInts([]int{22, 23, 23, 23}, offset),
				TilesUp:    offsetInts([]int{31, 31, 31, 31}, offset),
			},
			Backslash: &body.AnimationParams{
				TilesetSrc: equipFeetTileset,
				TilesDown:  offsetInts([]int{7, 7, 7, 7}, offset),
				TilesRight: offsetInts([]int{15, 15, 15, 14}, offset),
				TilesLeft:  offsetInts([]int{23, 23, 23, 22}, offset),
				TilesUp:    offsetInts([]int{31, 31, 31, 31}, offset),
			},
			Shield: &body.AnimationParams{
				TilesetSrc: equipFeetTileset,
				TilesDown:  offsetInts([]int{7}, offset),
				TilesRight: offsetInts([]int{15}, offset),
				TilesLeft:  offsetInts([]int{23}, offset),
				TilesUp:    offsetInts([]int{31}, offset),
			},
		})

		equipFeetOptions = append(equipFeetOptions, def)
	}

	equipHeadOptions := []body.SelectedPartDef{}
	for i := range 10 {
		index := i * 4
		cropHair, found := tiled.GetTileBoolProperty(equipHeadTileset, index, "COVER_HAIR")

		offset := i * 4
		animParams := body.AnimationParams{
			TilesetSrc: equipHeadTileset,
			TilesDown:  []int{0 + offset},
			TilesRight: []int{1 + offset},
			TilesLeft:  []int{2 + offset},
			TilesUp:    []int{3 + offset},
		}
		def := body.NewPartDef(body.PartDefParams{
			Idle:           &animParams,
			Walk:           &animParams,
			Run:            &animParams,
			Slash:          &animParams,
			Backslash:      &animParams,
			Shield:         &animParams,
			CropHairToHead: found && cropHair,
		})
		equipHeadOptions = append(equipHeadOptions, def)
	}

	weaponOptions := []weaponOption{}
	oneHandedWeapons := []body.SelectedPartDef{}
	for i := range 2 {
		weaponOffset := 73 * (i + 1)
		weaponDef := body.NewPartDef(body.PartDefParams{
			Idle: &body.AnimationParams{
				TilesetSrc: equipWeaponTileset,
				TilesDown:  offsetInts([]int{0}, weaponOffset),
				TilesRight: offsetInts([]int{19}, weaponOffset),
				TilesLeft:  offsetInts([]int{37}, weaponOffset),
				TilesUp:    offsetInts([]int{55}, weaponOffset),
			},
			Walk: &body.AnimationParams{
				TilesetSrc: equipWeaponTileset,
				TilesDown:  offsetInts([]int{4, 0, 6, 0}, weaponOffset),
				TilesRight: offsetInts([]int{23, 20, 25, 20}, weaponOffset),
				TilesLeft:  offsetInts([]int{41, 37, 43, 37}, weaponOffset),
				TilesUp:    offsetInts([]int{59, 55, 61, 55}, weaponOffset),
			},
			Run: &body.AnimationParams{
				TilesetSrc: equipWeaponTileset,
				TilesDown:  offsetInts([]int{3, 4, 0, 5, 6, 0}, weaponOffset),
				TilesRight: offsetInts([]int{22, 23, 20, 24, 25, 20}, weaponOffset),
				TilesLeft:  offsetInts([]int{40, 41, 37, 42, 43, 37}, weaponOffset),
				TilesUp:    offsetInts([]int{58, 59, 55, 60, 61, 55}, weaponOffset),
			},
			Slash: &body.AnimationParams{
				TilesetSrc: equipWeaponTileset,
				TilesDown:  offsetInts([]int{7, 8, 9, 9}, weaponOffset),
				TilesRight: offsetInts([]int{26, 27, 28, 28}, weaponOffset),
				TilesLeft:  offsetInts([]int{44, 45, 46, 46}, weaponOffset),
				TilesUp:    offsetInts([]int{62, 63, 64, 64}, weaponOffset),
			},
			Backslash: &body.AnimationParams{
				TilesetSrc: equipWeaponTileset,
				TilesDown:  offsetInts([]int{10, 9, 8, 7}, weaponOffset),
				TilesRight: offsetInts([]int{28, 27, 26, 26}, weaponOffset),
				TilesLeft:  offsetInts([]int{46, 45, 44, 44}, weaponOffset),
				TilesUp:    offsetInts([]int{64, 63, 62, 62}, weaponOffset),
			},
			Shield: &body.AnimationParams{
				TilesetSrc: equipWeaponTileset,
				TilesDown:  offsetInts([]int{11}, weaponOffset),
				TilesRight: offsetInts([]int{29}, weaponOffset),
				TilesLeft:  offsetInts([]int{47}, weaponOffset),
				TilesUp:    offsetInts([]int{65}, weaponOffset),
			},
		})
		oneHandedWeapons = append(oneHandedWeapons, weaponDef)
	}
	swordFx01 := body.NewPartDef(body.PartDefParams{
		Idle:   &body.AnimationParams{Skip: true},
		Walk:   &body.AnimationParams{Skip: true},
		Run:    &body.AnimationParams{Skip: true},
		Shield: &body.AnimationParams{Skip: true},
		Slash: &body.AnimationParams{
			TilesetSrc: weaponFxTileset,
			TilesDown:  []int{-1, 0, 1, 2},
			TilesRight: []int{-1, 9, 10, 11},
			TilesLeft:  []int{-1, 18, 19, 20},
			TilesUp:    []int{-1, 27, 28, 29},
		},
		Backslash: &body.AnimationParams{
			TilesetSrc: weaponFxTileset,
			TilesDown:  []int{-1, 3, 4, 5},
			TilesRight: []int{-1, 12, 13, 14},
			TilesLeft:  []int{-1, 21, 22, 23},
			TilesUp:    []int{-1, 30, 31, 32},
		},
	})

	for _, weaponDef := range oneHandedWeapons {
		weaponOptions = append(weaponOptions, weaponOption{
			weaponPartDef: weaponDef,
			weaponFxDef:   swordFx01,
		})
	}

	auxAnimated := []body.SelectedPartDef{}
	auxStatic := []body.SelectedPartDef{}

	for i := range 1 {
		offset := i * 76
		def := body.NewPartDef(body.PartDefParams{
			Idle: &body.AnimationParams{
				TilesetSrc: auxTileset,
				TilesDown:  offsetInts([]int{0, 1, 2, 3}, offset),
				TilesRight: offsetInts([]int{19, 20, 21, 22}, offset),
				TilesLeft:  offsetInts([]int{38, 39, 40, 41}, offset),
				TilesUp:    offsetInts([]int{57, 58, 59, 60}, offset),
			},
			Walk: &body.AnimationParams{
				TilesetSrc: auxTileset,
				TilesDown:  offsetInts([]int{5, 0, 7, 0}, offset),
				TilesRight: offsetInts([]int{24, 19, 26, 19}, offset),
				TilesLeft:  offsetInts([]int{43, 38, 46, 38}, offset),
				TilesUp:    offsetInts([]int{62, 57, 64, 57}, offset),
			},
			Run: &body.AnimationParams{
				TilesetSrc: auxTileset,
				TilesDown:  offsetInts([]int{4, 5, 0, 6, 7, 0}, offset),
				TilesRight: offsetInts([]int{23, 24, 19, 25, 26, 19}, offset),
				TilesLeft:  offsetInts([]int{42, 43, 38, 44, 45, 38}, offset),
				TilesUp:    offsetInts([]int{61, 62, 57, 63, 64, 57}, offset),
			},
			Slash: &body.AnimationParams{
				TilesetSrc: auxTileset,
				TilesDown:  offsetInts([]int{8, 9, 10, 10}, offset),
				TilesRight: offsetInts([]int{27, 28, 29, 30}, offset),
				TilesLeft:  offsetInts([]int{46, 47, 48, 49}, offset),
				TilesUp:    offsetInts([]int{65, 66, 67, 68}, offset),
			},
			Backslash: &body.AnimationParams{
				TilesetSrc: auxTileset,
				TilesDown:  offsetInts([]int{11, 12, 13, 14}, offset),
				TilesRight: offsetInts([]int{30, 31, 32, 33}, offset),
				TilesLeft:  offsetInts([]int{49, 50, 51, 52}, offset),
				TilesUp:    offsetInts([]int{68, 69, 70, 71}, offset),
			},
			Shield: &body.AnimationParams{Skip: true},
		})
		auxAnimated = append(auxAnimated, def)
	}

	for i := range 1 {
		offset := (i + 1) * 76
		def := body.NewPartDef(body.PartDefParams{
			Idle: &body.AnimationParams{
				TilesetSrc: auxTileset,
				TilesDown:  offsetInts([]int{0}, offset),
				TilesRight: offsetInts([]int{19}, offset),
				TilesLeft:  offsetInts([]int{38}, offset),
				TilesUp:    offsetInts([]int{57}, offset),
			},
			Walk: &body.AnimationParams{
				TilesetSrc: auxTileset,
				TilesDown:  offsetInts([]int{5, 0, 7, 0}, offset),
				TilesRight: offsetInts([]int{24, 19, 26, 19}, offset),
				TilesLeft:  offsetInts([]int{43, 38, 46, 38}, offset),
				TilesUp:    offsetInts([]int{62, 57, 64, 57}, offset),
			},
			Run: &body.AnimationParams{
				TilesetSrc: auxTileset,
				TilesDown:  offsetInts([]int{4, 5, 0, 6, 7, 0}, offset),
				TilesRight: offsetInts([]int{23, 24, 19, 25, 26, 19}, offset),
				TilesLeft:  offsetInts([]int{42, 43, 38, 44, 45, 38}, offset),
				TilesUp:    offsetInts([]int{61, 62, 57, 63, 64, 57}, offset),
			},
			Slash: &body.AnimationParams{
				TilesetSrc: auxTileset,
				TilesDown:  offsetInts([]int{8, 9, 10, 10}, offset),
				TilesRight: offsetInts([]int{27, 28, 29, 30}, offset),
				TilesLeft:  offsetInts([]int{46, 47, 48, 49}, offset),
				TilesUp:    offsetInts([]int{65, 66, 67, 68}, offset),
			},
			Backslash: &body.AnimationParams{
				TilesetSrc: auxTileset,
				TilesDown:  offsetInts([]int{11, 12, 13, 14}, offset),
				TilesRight: offsetInts([]int{30, 31, 32, 33}, offset),
				TilesLeft:  offsetInts([]int{49, 50, 51, 52}, offset),
				TilesUp:    offsetInts([]int{68, 69, 70, 71}, offset),
			},
			Shield: &body.AnimationParams{Skip: true},
		})
		auxStatic = append(auxStatic, def)
	}

	return []item.ItemDef{
		&item.WeaponDef{
			ItemBase: *item.NewItemBase(item.ItemBaseParams{
				ID:                "longsword_01",
				Name:              "Iron Longsword",
				Description:       "An iron longsword forged by blacksmiths in Gaul.",
				Value:             100,
				Weight:            25,
				MaxDurability:     250,
				TileImgTilesetSrc: "items/items_01.tsj",
				TileImgIndex:      0,
				Type:              item.TypeWeapon,
				BodyPartDef:       &weaponOptions[0].weaponPartDef,
			}),
			Damage:        10,
			HitsPerSecond: 1,
			FxPartDef:     &weaponOptions[0].weaponFxDef,
		},
		item.NewItemBase(
			item.ItemBaseParams{
				ID:                "potion_herculean_strength",
				Name:              "Potion of Herculean Strength",
				Description:       "This potion invigorates the drinker and gives him strength only matched by Hercules himself.",
				Value:             200,
				Weight:            3,
				TileImgTilesetSrc: "items/items_01.tsj",
				TileImgIndex:      129,
				Groupable:         true,
				Type:              item.TypeConsumable,
			},
		),
		item.NewItemBase(
			item.ItemBaseParams{
				ID:                "currency_value_1",
				Name:              "Aes",
				Description:       "A Roman bronze coin",
				Value:             1,
				Weight:            0.05,
				TileImgTilesetSrc: "items/items_01.tsj",
				TileImgIndex:      64,
				Groupable:         true,
				Type:              item.TypeCurrency,
			},
		),
		item.NewItemBase(
			item.ItemBaseParams{
				ID:                "currency_value_5",
				Name:              "Dupondius",
				Description:       "A Roman brass coin",
				Value:             5,
				Weight:            0.05,
				TileImgTilesetSrc: "items/items_01.tsj",
				TileImgIndex:      65,
				Groupable:         true,
				Type:              item.TypeCurrency,
			},
		),
		item.NewItemBase(
			item.ItemBaseParams{
				ID:                "currency_value_10",
				Name:              "Sestertius",
				Description:       "A Roman brass coin",
				Value:             10,
				Weight:            0.05,
				TileImgTilesetSrc: "items/items_01.tsj",
				TileImgIndex:      66,
				Groupable:         true,
				Type:              item.TypeCurrency,
			},
		),
		item.NewItemBase(
			item.ItemBaseParams{
				ID:                "currency_value_50",
				Name:              "Quinarius",
				Description:       "A Roman silver coin",
				Value:             50,
				Weight:            0.05,
				TileImgTilesetSrc: "items/items_01.tsj",
				TileImgIndex:      67,
				Groupable:         true,
				Type:              item.TypeCurrency,
			},
		),
		item.NewItemBase(
			item.ItemBaseParams{
				ID:                "currency_value_100",
				Name:              "Denarius",
				Description:       "A Roman silver coin",
				Value:             100,
				Weight:            0.05,
				TileImgTilesetSrc: "items/items_01.tsj",
				TileImgIndex:      68,
				Groupable:         true,
				Type:              item.TypeCurrency,
			},
		),
		item.NewItemBase(
			item.ItemBaseParams{
				ID:                "currency_value_1000",
				Name:              "Aureus",
				Description:       "A Roman gold coin",
				Value:             1000,
				Weight:            0.05,
				TileImgTilesetSrc: "items/items_01.tsj",
				TileImgIndex:      69,
				Groupable:         true,
				Type:              item.TypeCurrency,
			},
		),
		&item.ArmorDef{
			ItemBase: *item.NewItemBase(item.ItemBaseParams{
				ID:                "legionary_helm",
				Name:              "Legionary Helm",
				Description:       "A standard issue steel helmet for Roman legionaries.",
				Value:             250,
				Weight:            15,
				TileImgTilesetSrc: "items/items_01.tsj",
				TileImgIndex:      32,
				Type:              item.TypeHeadwear,
				BodyPartDef:       &equipHeadOptions[2],
			}),
			Protection: 10,
		},
		&item.ArmorDef{
			ItemBase: *item.NewItemBase(item.ItemBaseParams{
				ID:                "legionary_cuirass",
				Name:              "Legionary Cuirass",
				Description:       "A set of Lorica Segmentata body armor, used by Roman legionaries.",
				Value:             700,
				Weight:            25,
				TileImgTilesetSrc: "items/items_01.tsj",
				TileImgIndex:      33,
				Type:              item.TypeBodywear,
				BodyPartDef:       &equipBodyOptions[0].bodyDef,
				LegsPartDef:       &equipBodyOptions[0].legsDef,
			}),
			Protection: 18,
		},
		&item.ArmorDef{
			ItemBase: *item.NewItemBase(item.ItemBaseParams{
				ID:                "caligae_boots",
				Name:              "Caligae",
				Description:       "A pair of caligae, heavy leather sandals commonly worn by Roman soldiers.",
				Value:             15,
				Weight:            4,
				TileImgTilesetSrc: "items/items_01.tsj",
				TileImgIndex:      34,
				Type:              item.TypeFootwear,
				BodyPartDef:       &equipFeetOptions[0],
			}),
			Protection: 5,
		},
		item.NewItemBase(item.ItemBaseParams{
			ID:                "torch",
			Name:              "Torch",
			Description:       "A torch to light your way in dark places.",
			Value:             5,
			Weight:            2.5,
			TileImgTilesetSrc: "items/items_01.tsj",
			TileImgIndex:      97,
			Type:              item.TypeAuxiliary,
			BodyPartDef:       &auxAnimated[0],
		}),
		&item.ArmorDef{
			ItemBase: *item.NewItemBase(item.ItemBaseParams{
				ID:                "legionary_shield",
				Name:              "Legionary Shield",
				Description:       "A standard-issue shield used by Roman legionaries.",
				Value:             80,
				Weight:            10,
				TileImgTilesetSrc: "items/items_01.tsj",
				TileImgIndex:      35,
				Type:              item.TypeAuxiliary,
				BodyPartDef:       &auxStatic[0],
			}),
			Protection: 8,
		},
		item.NewItemBase(item.ItemBaseParams{
			ID:                "toga_white",
			Name:              "Toga",
			Description:       "A Roman Toga of fine quality, often worn by the Senatorial and elite classes of Roman society.",
			Value:             500,
			Weight:            8,
			TileImgTilesetSrc: "items/items_01.tsj",
			TileImgIndex:      192,
			Type:              item.TypeBodywear,
			BodyPartDef:       &equipBodyOptions[0].bodyDef,
			LegsPartDef:       &equipBodyOptions[0].legsDef,
		}),
		item.NewItemBase(item.ItemBaseParams{
			ID:                "laurel_wreath",
			Name:              "Laurel Wreath",
			Description:       "A wreath of laurel branches, ceremonially worn by triumphant generals, athletes, and poets alike. It's just leaves and branches though.",
			Value:             50,
			Weight:            0.5,
			TileImgTilesetSrc: "items/items_01.tsj",
			TileImgIndex:      70,
			Type:              item.TypeHeadwear,
			BodyPartDef:       &equipHeadOptions[0],
		}),
	}
}
