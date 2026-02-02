package cmd

import (
	"github.com/webbben/2d-game-engine/skills"
)

const (
	Strength     skills.AttributeID = "STRENGTH"
	Endurance    skills.AttributeID = "ENDURANCE"
	Martial      skills.AttributeID = "MARTIAL"
	Agility      skills.AttributeID = "AGILITY"
	Intelligence skills.AttributeID = "INTELLIGENCE"
	Intrigue     skills.AttributeID = "INTRIGUE"
	Personality  skills.AttributeID = "PERSONALITY"
	Luck         skills.AttributeID = "LUCK"

	Blade        skills.SkillID = "BLADE"
	Blunt        skills.SkillID = "BLUNT"
	Axe          skills.SkillID = "AXE"
	Spear        skills.SkillID = "SPEAR"
	Marksmanship skills.SkillID = "MARKSMANSHIP"
	Repair       skills.SkillID = "REPAIR"
	HeavyArmor   skills.SkillID = "HEAVY_ARMOR"
	LightArmor   skills.SkillID = "LIGHT_ARMOR"
	Security     skills.SkillID = "SECURITY"
	Sneak        skills.SkillID = "SNEAK"
	Speechcraft  skills.SkillID = "SPEECHCRAFT"
	Mercantile   skills.SkillID = "MERCANTILE"
	Alchemy      skills.SkillID = "ALCHEMY"
	Incantation  skills.SkillID = "INCANTATION"
)

func GetAllAttributes() []skills.AttributeDef {
	return []skills.AttributeDef{
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

func GetAllSkills() []skills.SkillDef {
	return []skills.SkillDef{
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
	TraitBrave    skills.TraitID = "BRAVE"
	TraitTimid    skills.TraitID = "TIMID"
	TraitGenerous skills.TraitID = "GENEROUS"
	TraitGreedy   skills.TraitID = "GREEDY"
	TraitPolite   skills.TraitID = "POLITE"
	TraitRude     skills.TraitID = "RUDE"
	TraitCharming skills.TraitID = "CHARMING"
	TraitCreepy   skills.TraitID = "CREEPY"
	TraitAnnoying skills.TraitID = "ANNOYING"
	TraitStoic    skills.TraitID = "STOIC"
	TraitCruel    skills.TraitID = "CRUEL"
	TraitWrathful skills.TraitID = "WRATHFUL"
	TraitArrogant skills.TraitID = "ARROGANT"
	TraitParanoid skills.TraitID = "PARANOID"
	TraitIdiot    skills.TraitID = "IDIOT"
)

func GetAllTraits() []skills.Trait {
	tilesetSrc := "entities/traits.tsj"
	traits := []skills.Trait{
		skills.NewSkillTrait(skills.TraitParams{
			ID:               TraitBrave,
			Name:             "Brave",
			Description:      "This person never backs down from a challenge, and is fearless in the face of danger",
			TilesetSrc:       tilesetSrc,
			TileID:           0,
			ConflictTraitIDs: []string{"timid"},
		}, nil, map[skills.AttributeID]int{
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
		}, map[skills.SkillID]int{
			Sneak: buff2,
		}, map[skills.AttributeID]int{
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
		}, map[skills.SkillID]int{
			Mercantile: debuff3,
		}, map[skills.AttributeID]int{
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
		}, map[skills.SkillID]int{
			Mercantile: buff3,
		}, map[skills.AttributeID]int{
			Personality: debuff2,
		}),
		skills.NewSkillTrait(skills.TraitParams{
			ID:               TraitPolite,
			Name:             "Polite",
			Description:      "This person is well-versed in the ettiquette and decorum of polite society",
			TilesetSrc:       tilesetSrc,
			TileID:           4,
			ConflictTraitIDs: []string{"rude"},
		}, map[skills.SkillID]int{}, map[skills.AttributeID]int{
			Personality: buff2,
		}),
		skills.NewSkillTrait(skills.TraitParams{
			ID:               TraitRude,
			Name:             "Rude",
			Description:      "This person has the vocabulary of a sailor and the table manners of a wild boar",
			TilesetSrc:       tilesetSrc,
			TileID:           5,
			ConflictTraitIDs: []string{"charming", "polite"},
		}, map[skills.SkillID]int{
			Speechcraft: debuff2,
		}, map[skills.AttributeID]int{
			Personality: debuff2,
		}),
		skills.NewSkillTrait(skills.TraitParams{
			ID:               TraitCharming,
			Name:             "Charming",
			Description:      "This person has a way with people, and can warm even the coldest of hearts",
			TilesetSrc:       tilesetSrc,
			TileID:           6,
			ConflictTraitIDs: []string{"rude", "creepy"},
		}, map[skills.SkillID]int{
			Speechcraft: buff3,
		}, map[skills.AttributeID]int{
			Personality: buff2,
		}),
		skills.NewSkillTrait(skills.TraitParams{
			ID:               TraitCreepy,
			Name:             "Creepy",
			Description:      "This person gives people the creeps...",
			TilesetSrc:       tilesetSrc,
			TileID:           7,
			ConflictTraitIDs: []string{"charming"},
		}, map[skills.SkillID]int{
			Sneak: buff2,
		}, map[skills.AttributeID]int{
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
		}, map[skills.SkillID]int{
			Speechcraft: debuff2,
		}, map[skills.AttributeID]int{
			Personality: debuff2,
		}),
		skills.NewSkillTrait(skills.TraitParams{
			ID:               TraitStoic,
			Name:             "Stoic",
			Description:      "This person is a student of Stoicism. It takes a lot to get under their skin.",
			TilesetSrc:       tilesetSrc,
			TileID:           9,
			ConflictTraitIDs: []string{"wrathful"},
		}, map[skills.SkillID]int{}, map[skills.AttributeID]int{
			Intelligence: buff1,
			Endurance:    buff1,
		}),
		skills.NewSkillTrait(skills.TraitParams{
			ID:          TraitCruel,
			Name:        "Cruel",
			Description: "This person takes an odd pleasure in the sufferings of others",
			TilesetSrc:  tilesetSrc,
			TileID:      10,
		}, map[skills.SkillID]int{}, map[skills.AttributeID]int{
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
		}, map[skills.SkillID]int{}, map[skills.AttributeID]int{
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
		}, map[skills.SkillID]int{}, map[skills.AttributeID]int{
			Personality: debuff1,
		}),
		skills.NewSkillTrait(skills.TraitParams{
			ID:          TraitParanoid,
			Name:        "Paranoid",
			Description: "This person thinks the world is out to get him",
			TilesetSrc:  tilesetSrc,
			TileID:      13,
		}, map[skills.SkillID]int{}, map[skills.AttributeID]int{
			Intrigue:    buff3,
			Personality: debuff2,
		}),
		skills.NewSkillTrait(skills.TraitParams{
			ID:          TraitIdiot,
			Name:        "Idiot",
			Description: "This person is, quite simply, an idiot",
			TilesetSrc:  tilesetSrc,
			TileID:      14,
		}, map[skills.SkillID]int{}, map[skills.AttributeID]int{
			Intelligence: debuff4,
			Intrigue:     debuff4,
			Luck:         buff1,
		}),
	}

	return traits
}
