package cmd

import (
	"fmt"

	"github.com/webbben/2d-game-engine/entity/body"
)

// EntityBodySkinSet represents the body parts that are the base "skin" (i.e. the base creature, without any sort of armor, clothing, weapons, etc)
// Since these will be defined in groups (a given arms set likely will be assigned directly to a specific body set, for example), we have this struct to group them.
// However, in the body struct itself, we won't use this. It's just convenient for the character builder.
type EntityBodySkinSet struct {
	Body, Arms, Legs body.SelectedPartDef
}

func allBodyParts() (bodyDefs, armDefs, legDefs, eyeDefs, hairDefs []body.SelectedPartDef) {
	bodyTileset := "entities/parts/human_entity_parts.tsj"
	armsTileset := "entities/parts/human_entity_parts.tsj"
	eyesTileset := "entities/parts/eyes.tsj"
	hairTileset := "entities/parts/hair.tsj"

	// BODY
	bodyDefs = []body.SelectedPartDef{
		body.NewPartDef(body.PartDefParams{
			ID: "body_0",
			Idle: &body.AnimationParams{
				TilesetSrc:   bodyTileset,
				TilesDown:    []int{73},
				TilesRight:   []int{92},
				TilesLeft:    []int{110},
				TilesUp:      []int{128},
				StepsOffsetY: []int{0},
			},
			Walk: &body.AnimationParams{
				TilesetSrc:   bodyTileset,
				TilesDown:    []int{75, 73, 77, 73},
				TilesRight:   []int{94, 92, 94, 92},
				TilesLeft:    []int{112, 110, 112, 110},
				TilesUp:      []int{130, 128, 130, 128},
				StepsOffsetY: []int{1, 0, 1, 0},
			},
			Run: &body.AnimationParams{
				TilesetSrc:   bodyTileset,
				TilesDown:    []int{74, 75, 73, 76, 77, 73},
				TilesRight:   []int{93, 94, 92, 93, 94, 92},
				TilesLeft:    []int{111, 112, 110, 111, 112, 110},
				TilesUp:      []int{129, 130, 128, 129, 130, 128},
				StepsOffsetY: []int{0, 1, 0, 0, 1, 0},
			},
			Slash: &body.AnimationParams{
				TilesetSrc:     bodyTileset,
				TilesDown:      []int{78, 79, 79, 79},
				TilesRight:     []int{94, 95, 95, 95},
				TilesLeft:      []int{112, 113, 113, 113},
				TilesUp:        []int{128, 131, 131, 131},
				StepsOffsetY:   []int{1, 2, 2, 2},
				StepsOffsetYUp: []int{0, 1, 1, 1},
			},
			Backslash: &body.AnimationParams{
				TilesetSrc:       bodyTileset,
				TilesDown:        []int{79, 79, 79, 78},
				TilesRight:       []int{95, 95, 94, 94},
				TilesLeft:        []int{113, 113, 112, 112},
				TilesUp:          []int{131, 131, 128, 128},
				StepsOffsetY:     []int{2, 2, 1, 1},
				StepsOffsetYDown: []int{2, 2, 2, 1},
				StepsOffsetYUp:   []int{1, 1, 0, 0},
			},
			Shield: &body.AnimationParams{
				TilesetSrc:   bodyTileset,
				TilesDown:    []int{78},
				TilesRight:   []int{94},
				TilesLeft:    []int{112},
				TilesUp:      []int{128},
				StepsOffsetY: []int{1},
			},
		}),
	}

	// ARMS
	armDefs = []body.SelectedPartDef{
		body.NewPartDef(body.PartDefParams{
			ID: "arms_0",
			Idle: &body.AnimationParams{
				TilesetSrc: armsTileset,
				TilesDown:  []int{146},
				AuxDown:    []int{147},
				TilesRight: []int{165},
				AuxRight:   []int{166},
				TilesLeft:  []int{183},
				AuxLeft:    []int{184},
				TilesUp:    []int{201},
				AuxUp:      []int{202},
			},
			Walk: &body.AnimationParams{
				TilesetSrc: armsTileset,
				TilesDown:  []int{150, 146, 152, 146},
				TilesRight: []int{169, 165, 171, 165},
				TilesLeft:  []int{187, 183, 189, 183},
				TilesUp:    []int{205, 201, 207, 201},
			},
			Run: &body.AnimationParams{
				TilesetSrc: armsTileset,
				TilesDown:  []int{149, 150, 146, 151, 152, 146},
				TilesRight: []int{168, 169, 165, 170, 171, 165},
				TilesLeft:  []int{186, 187, 183, 188, 189, 183},
				TilesUp:    []int{204, 205, 201, 206, 207, 201},
			},
			Slash: &body.AnimationParams{
				TilesetSrc: armsTileset,
				TilesDown:  []int{153, 154, 155, 155},
				TilesRight: []int{172, 173, 174, 174},
				TilesLeft:  []int{190, 191, 192, 192},
				TilesUp:    []int{208, 207, 209, 209},
			},
			Backslash: &body.AnimationParams{
				TilesetSrc: armsTileset,
				TilesDown:  []int{156, 155, 154, 153},
				TilesRight: []int{174, 173, 172, 172},
				TilesLeft:  []int{192, 191, 190, 190},
				TilesUp:    []int{209, 207, 208, 208},
			},
			Shield: &body.AnimationParams{
				TilesetSrc: armsTileset,
				TilesDown:  []int{157},
				TilesRight: []int{175},
				TilesLeft:  []int{193},
				TilesUp:    []int{204},
			},
		}),
	}

	// LEGS
	legDefs = []body.SelectedPartDef{
		body.NewPartDef(body.PartDefParams{
			ID: "legs_0",
			Idle: &body.AnimationParams{
				TilesetSrc: bodyTileset,
				TilesDown:  []int{80},
				TilesRight: []int{99},
				TilesLeft:  []int{117},
				TilesUp:    []int{135},
			},
			Walk: &body.AnimationParams{
				TilesetSrc: bodyTileset,
				TilesDown:  []int{84, 80, 86, 80},
				TilesRight: []int{103, 99, 105, 99},
				TilesLeft:  []int{121, 117, 123, 117},
				TilesUp:    []int{139, 135, 141, 135},
			},
			Run: &body.AnimationParams{
				TilesetSrc: bodyTileset,
				TilesDown:  []int{83, 84, 80, 85, 86, 80},
				TilesRight: []int{102, 103, 99, 104, 105, 99},
				TilesLeft:  []int{120, 121, 117, 122, 123, 117},
				TilesUp:    []int{138, 139, 135, 140, 141, 135},
			},
			Slash: &body.AnimationParams{
				TilesetSrc: bodyTileset,
				TilesDown:  []int{87, 87, 87, 87},
				TilesRight: []int{103, 106, 106, 106},
				TilesLeft:  []int{123, 124, 124, 124},
				TilesUp:    []int{142, 142, 142, 142},
			},
			Backslash: &body.AnimationParams{
				TilesetSrc: bodyTileset,
				TilesDown:  []int{87, 87, 87, 87},
				TilesRight: []int{106, 106, 103, 103},
				TilesLeft:  []int{124, 124, 123, 123},
				TilesUp:    []int{142, 142, 142, 142},
			},
			Shield: &body.AnimationParams{
				TilesetSrc: bodyTileset,
				TilesDown:  []int{87},
				TilesRight: []int{106},
				TilesLeft:  []int{124},
				TilesUp:    []int{142},
			},
		}),
	}

	// EYES
	for i := range 14 {
		numCols := 32 // number of colums in the tileset
		animParams := body.AnimationParams{
			TilesetSrc: eyesTileset,
			TilesDown:  []int{numCols * i},
			TilesRight: []int{1 + (numCols * i)},
		}
		eyeDefs = append(eyeDefs, body.NewPartDef(body.PartDefParams{
			ID:        fmt.Sprintf("eyes_%v", i),
			FlipRForL: true,
			Idle:      &animParams,
			Walk:      &animParams,
			Run:       &animParams,
			Slash:     &animParams,
			Backslash: &animParams,
			Shield:    &animParams,
		}))
	}

	// HAIR
	for i := range 7 {
		numCols := 32
		animParams := body.AnimationParams{
			TilesetSrc: hairTileset,
			TilesDown:  []int{numCols * i},
			TilesRight: []int{(numCols * i) + 1},
			TilesLeft:  []int{(numCols * i) + 2},
			TilesUp:    []int{(numCols * i) + 3},
		}
		hairDefs = append(hairDefs, body.NewPartDef(body.PartDefParams{
			ID:        fmt.Sprintf("hair_%v", i),
			Idle:      &animParams,
			Walk:      &animParams,
			Run:       &animParams,
			Slash:     &animParams,
			Backslash: &animParams,
			Shield:    &animParams,
		}))
	}

	return
}

func GetAllEntityBodyPartSets() (skins []EntityBodySkinSet, eyes []body.SelectedPartDef, hair []body.SelectedPartDef) {
	bodySets, armSets, legSets, eyeSets, hairSets := allBodyParts()

	skins = append(skins, EntityBodySkinSet{
		Body: bodySets[0],
		Legs: legSets[0],
		Arms: armSets[0],
	})

	eyes = eyeSets
	hair = hairSets
	return
}
