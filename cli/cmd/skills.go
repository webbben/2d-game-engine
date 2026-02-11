package cmd

import (
	"github.com/webbben/2d-game-engine/data/defs"
	"github.com/webbben/2d-game-engine/skills"
)

const (
	Strength     defs.AttributeID = "STRENGTH"
	Endurance    defs.AttributeID = "ENDURANCE"
	Martial      defs.AttributeID = "MARTIAL"
	Agility      defs.AttributeID = "AGILITY"
	Intelligence defs.AttributeID = "INTELLIGENCE"
	Intrigue     defs.AttributeID = "INTRIGUE"
	Personality  defs.AttributeID = "PERSONALITY"
	Luck         defs.AttributeID = "LUCK"

	Blade        defs.SkillID = "BLADE"
	Blunt        defs.SkillID = "BLUNT"
	Axe          defs.SkillID = "AXE"
	Spear        defs.SkillID = "SPEAR"
	Marksmanship defs.SkillID = "MARKSMANSHIP"
	Repair       defs.SkillID = "REPAIR"
	HeavyArmor   defs.SkillID = "HEAVY_ARMOR"
	LightArmor   defs.SkillID = "LIGHT_ARMOR"
	Security     defs.SkillID = "SECURITY"
	Sneak        defs.SkillID = "SNEAK"
	Speechcraft  defs.SkillID = "SPEECHCRAFT"
	Mercantile   defs.SkillID = "MERCANTILE"
	Alchemy      defs.SkillID = "ALCHEMY"
	Incantation  defs.SkillID = "INCANTATION"
)

func GetAllAttributes() []defs.AttributeDef {
	return []defs.AttributeDef{
		{
			ID:          Strength,
			DisplayName: "Strength",
		},
		{
			ID:          Endurance,
			DisplayName: "Endurance",
		},
		{
			ID:          Martial,
			DisplayName: "Martial",
		},
		{
			ID:          Agility,
			DisplayName: "Agility",
		},
		{
			ID:          Intelligence,
			DisplayName: "Intelligence",
		},
		{
			ID:          Intrigue,
			DisplayName: "Intrigue",
		},
		{
			ID:          Personality,
			DisplayName: "Personality",
		},
		{
			ID:          Luck,
			DisplayName: "Luck",
		},
	}
}

func GetAllSkills() []defs.SkillDef {
	return []defs.SkillDef{
		{
			ID:          Blade,
			DisplayName: "Blade",
		},
		{
			ID:          Blunt,
			DisplayName: "Blunt",
		},
		{
			ID:          Axe,
			DisplayName: "Axe",
		},
		{
			ID:          Spear,
			DisplayName: "Spear",
		},
		{
			ID:          Marksmanship,
			DisplayName: "Marksmanship",
		},
		{
			ID:          Repair,
			DisplayName: "Repair",
		},
		{
			ID:          HeavyArmor,
			DisplayName: "Heavy Armor",
		},
		{
			ID:          LightArmor,
			DisplayName: "Light Armor",
		},
		{
			ID:          Security,
			DisplayName: "Security",
		},
		{
			ID:          Sneak,
			DisplayName: "Sneak",
		},
		{
			ID:          Speechcraft,
			DisplayName: "Speechcraft",
		},
		{
			ID:          Mercantile,
			DisplayName: "Mercantile",
		},
		{
			ID:          Alchemy,
			DisplayName: "Alchemy",
		},
		{
			ID:          Incantation,
			DisplayName: "Incantation",
		},
	}
}

const (
	buff1   int = 5
	buff2   int = 10
	buff3   int = 15
	buff4   int = 20
	buff5   int = 25
	debuff1 int = -buff1
	debuff2 int = -buff2
	debuff3 int = -buff3
	debuff4 int = -buff4
	debuff5 int = -buff5
)

const (
	TraitBrave    defs.TraitID = "BRAVE"
	TraitTimid    defs.TraitID = "TIMID"
	TraitGenerous defs.TraitID = "GENEROUS"
	TraitGreedy   defs.TraitID = "GREEDY"
	TraitPolite   defs.TraitID = "POLITE"
	TraitRude     defs.TraitID = "RUDE"
	TraitCharming defs.TraitID = "CHARMING"
	TraitCreepy   defs.TraitID = "CREEPY"
	TraitAnnoying defs.TraitID = "ANNOYING"
	TraitStoic    defs.TraitID = "STOIC"
	TraitCruel    defs.TraitID = "CRUEL"
	TraitWrathful defs.TraitID = "WRATHFUL"
	TraitArrogant defs.TraitID = "ARROGANT"
	TraitParanoid defs.TraitID = "PARANOID"
	TraitIdiot    defs.TraitID = "IDIOT"
)

func GetAllTraits() []defs.Trait {
	tilesetSrc := "entities/traits.tsj"
	traits := []defs.Trait{
		skills.NewSkillTrait(skills.TraitParams{
			ID:               TraitBrave,
			Name:             "Brave",
			Description:      "This person never backs down from a challenge, and is fearless in the face of danger",
			TilesetSrc:       tilesetSrc,
			TileID:           0,
			ConflictTraitIDs: []string{"timid"},
		}, nil, map[defs.AttributeID]int{
			Martial:  buff2,
			Intrigue: debuff1,
		}),
		skills.NewSkillTrait(skills.TraitParams{
			ID:               TraitTimid,
			Name:             "Timid",
			Description:      "This person avoids danger and confrontation at all costs",
			TilesetSrc:       tilesetSrc,
			TileID:           1,
			ConflictTraitIDs: []string{"brave"},
		}, map[defs.SkillID]int{
			Sneak: buff2,
		}, map[defs.AttributeID]int{
			Martial:     debuff2,
			Personality: debuff1,
		}),
		skills.NewSkillTrait(skills.TraitParams{
			ID:               TraitGenerous,
			Name:             "Generous",
			Description:      "This person has a soft heart and will empty his pockets to help those less fortunate",
			TilesetSrc:       tilesetSrc,
			TileID:           2,
			ConflictTraitIDs: []string{"greedy"},
		}, map[defs.SkillID]int{
			Mercantile: debuff3,
		}, map[defs.AttributeID]int{
			Personality: buff2,
			Intrigue:    debuff1,
		}),
		skills.NewSkillTrait(skills.TraitParams{
			ID:               TraitGreedy,
			Name:             "Greedy",
			Description:      "This miserly fellow wouldn't give a single bronze piece to save his own mother",
			TilesetSrc:       tilesetSrc,
			TileID:           3,
			ConflictTraitIDs: []string{"generous"},
		}, map[defs.SkillID]int{
			Mercantile: buff3,
		}, map[defs.AttributeID]int{
			Personality: debuff2,
		}),
		skills.NewSkillTrait(skills.TraitParams{
			ID:               TraitPolite,
			Name:             "Polite",
			Description:      "This person is well-versed in the ettiquette and decorum of polite society",
			TilesetSrc:       tilesetSrc,
			TileID:           4,
			ConflictTraitIDs: []string{"rude"},
		}, map[defs.SkillID]int{}, map[defs.AttributeID]int{
			Personality: buff2,
		}),
		skills.NewSkillTrait(skills.TraitParams{
			ID:               TraitRude,
			Name:             "Rude",
			Description:      "This person has the vocabulary of a sailor and the table manners of a wild boar",
			TilesetSrc:       tilesetSrc,
			TileID:           5,
			ConflictTraitIDs: []string{"charming", "polite"},
		}, map[defs.SkillID]int{
			Speechcraft: debuff2,
		}, map[defs.AttributeID]int{
			Personality: debuff2,
		}),
		skills.NewSkillTrait(skills.TraitParams{
			ID:               TraitCharming,
			Name:             "Charming",
			Description:      "This person has a way with people, and can warm even the coldest of hearts",
			TilesetSrc:       tilesetSrc,
			TileID:           6,
			ConflictTraitIDs: []string{"rude", "creepy"},
		}, map[defs.SkillID]int{
			Speechcraft: buff3,
		}, map[defs.AttributeID]int{
			Personality: buff2,
		}),
		skills.NewSkillTrait(skills.TraitParams{
			ID:               TraitCreepy,
			Name:             "Creepy",
			Description:      "This person gives people the creeps...",
			TilesetSrc:       tilesetSrc,
			TileID:           7,
			ConflictTraitIDs: []string{"charming"},
		}, map[defs.SkillID]int{
			Sneak: buff2,
		}, map[defs.AttributeID]int{
			Personality: debuff3,
			Intrigue:    buff1,
		}),
		skills.NewSkillTrait(skills.TraitParams{
			ID:               TraitAnnoying,
			Name:             "Annoying",
			Description:      "Oh god, it's this guy again",
			TilesetSrc:       tilesetSrc,
			TileID:           8,
			ConflictTraitIDs: []string{"charming"},
		}, map[defs.SkillID]int{
			Speechcraft: debuff2,
		}, map[defs.AttributeID]int{
			Personality: debuff2,
		}),
		skills.NewSkillTrait(skills.TraitParams{
			ID:               TraitStoic,
			Name:             "Stoic",
			Description:      "This person is a student of Stoicism. It takes a lot to get under their skin.",
			TilesetSrc:       tilesetSrc,
			TileID:           9,
			ConflictTraitIDs: []string{"wrathful"},
		}, map[defs.SkillID]int{}, map[defs.AttributeID]int{
			Intelligence: buff1,
			Endurance:    buff1,
		}),
		skills.NewSkillTrait(skills.TraitParams{
			ID:          TraitCruel,
			Name:        "Cruel",
			Description: "This person takes an odd pleasure in the sufferings of others",
			TilesetSrc:  tilesetSrc,
			TileID:      10,
		}, map[defs.SkillID]int{}, map[defs.AttributeID]int{
			Personality: debuff2,
			Intrigue:    buff1,
			Martial:     buff1,
		}),

		skills.NewSkillTrait(skills.TraitParams{
			ID:               TraitWrathful,
			Name:             "Wrathful",
			Description:      "This person can fly into a rage at the most minor of inconveniences",
			TilesetSrc:       tilesetSrc,
			TileID:           11,
			ConflictTraitIDs: []string{"stoic"},
		}, map[defs.SkillID]int{}, map[defs.AttributeID]int{
			Martial:     buff1,
			Personality: debuff2,
		}),
		skills.NewSkillTrait(skills.TraitParams{
			// TODO: this trait will probably take a prestige and opinion effect
			ID:          TraitArrogant,
			Name:        "Arrogant",
			Description: "This person holds his nose high and feels a sense of superiority over others",
			TilesetSrc:  tilesetSrc,
			TileID:      12,
		}, map[defs.SkillID]int{}, map[defs.AttributeID]int{
			Personality: debuff1,
		}),
		skills.NewSkillTrait(skills.TraitParams{
			ID:          TraitParanoid,
			Name:        "Paranoid",
			Description: "This person thinks the world is out to get him",
			TilesetSrc:  tilesetSrc,
			TileID:      13,
		}, map[defs.SkillID]int{}, map[defs.AttributeID]int{
			Intrigue:    buff3,
			Personality: debuff2,
		}),
		skills.NewSkillTrait(skills.TraitParams{
			ID:          TraitIdiot,
			Name:        "Idiot",
			Description: "This person is, quite simply, an idiot",
			TilesetSrc:  tilesetSrc,
			TileID:      14,
		}, map[defs.SkillID]int{}, map[defs.AttributeID]int{
			Intelligence: debuff4,
			Intrigue:     debuff4,
			Luck:         buff1,
		}),
	}

	return traits
}
