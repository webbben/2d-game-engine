package cmd

import (
	"github.com/webbben/2d-game-engine/entity/body"
	"github.com/webbben/2d-game-engine/internal/tiled"
	"github.com/webbben/2d-game-engine/item"
)

// At some point, these will probably be moved into the actual game repo. For now, defining here for testing.

func GetItemDefs() []item.ItemDef {
	equipBodyTileset := "items/equiped_body_01.tsj"
	equipHeadTileset := "items/equiped_head_01.tsj"
	equipWeaponTileset := "items/weapon_frames.tsj"
	weaponFxTileset := "items/weapon_fx_frames.tsj"
	auxTileset := "items/equiped_aux.tsj"

	bodyRowLength := 68
	bodyRStart := 17
	bodyLStart := 34
	bodyUStart := 51

	equipBodyOptions := []body.SelectedPartDef{}
	for i := range 4 {
		equipBodyOptions = append(equipBodyOptions, body.SelectedPartDef{
			TilesetSrc:        equipBodyTileset,
			DStart:            (i * bodyRowLength),
			RStart:            (i * bodyRowLength) + bodyRStart,
			LStart:            (i * bodyRowLength) + bodyLStart,
			UStart:            (i * bodyRowLength) + bodyUStart,
			AuxFirstFrameStep: 1,
		})
	}
	equipHeadOptions := []body.SelectedPartDef{}
	for i := range 3 {
		index := i * 4
		cropHair, found := tiled.GetTileBoolProperty(equipHeadTileset, index, "COVER_HAIR")
		equipHeadOptions = append(equipHeadOptions, body.SelectedPartDef{
			TilesetSrc:     equipHeadTileset,
			DStart:         (i * 4),
			RStart:         (i * 4) + 1,
			LStart:         (i * 4) + 2,
			UStart:         (i * 4) + 3,
			CropHairToHead: found && cropHair,
		})
	}

	weaponOptions := []weaponOption{
		{
			weaponPartDef: body.SelectedPartDef{
				TilesetSrc: equipWeaponTileset,
				DStart:     0,
				RStart:     16,
				LStart:     32,
				UStart:     48,
			},
			weaponFxDef: body.SelectedPartDef{
				TilesetSrc: weaponFxTileset,
				DStart:     0,
				RStart:     9,
				LStart:     18,
				UStart:     27,
			},
		},
	}

	auxOp := body.SelectedPartDef{
		TilesetSrc: auxTileset,
		DStart:     0,
		RStart:     19,
		LStart:     38,
		UStart:     57,
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
				BodyPartDef:       &equipBodyOptions[3],
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
				Type:              item.TypeMisc, // TODO - once we have footwear implemented, switch this type
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
			BodyPartDef:       &auxOp,
		}),
		item.NewItemBase(item.ItemBaseParams{
			ID:                "toga_white",
			Name:              "Toga",
			Description:       "A Roman Toga of fine quality, often worn by the Senatorial and elite classes of Roman society.",
			Value:             500,
			Weight:            8,
			TileImgTilesetSrc: "items/items_01.tsj",
			TileImgIndex:      192,
			Type:              item.TypeBodywear,
			BodyPartDef:       &equipBodyOptions[0],
		}),
	}
}
