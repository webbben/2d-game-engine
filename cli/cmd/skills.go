package cmd

import (
	"github.com/webbben/2d-game-engine/data/defs"
	"github.com/webbben/2d-game-engine/skills"
)

const (
	// Attributes

	Strength     defs.AttributeID = "STRENGTH"
	Endurance    defs.AttributeID = "ENDURANCE"
	Martial      defs.AttributeID = "MARTIAL"
	Agility      defs.AttributeID = "AGILITY"
	Intelligence defs.AttributeID = "INTELLIGENCE"
	Intrigue     defs.AttributeID = "INTRIGUE"
	Personality  defs.AttributeID = "PERSONALITY"
	Luck         defs.AttributeID = "LUCK"

	// Combat

	LongBlade  defs.SkillID = "LONG_BLADE"
	Blunt      defs.SkillID = "BLUNT"
	Axe        defs.SkillID = "AXE"
	Spear      defs.SkillID = "SPEAR"
	Repair     defs.SkillID = "REPAIR"
	HeavyArmor defs.SkillID = "HEAVY_ARMOR"
	Block      defs.SkillID = "BLOCK"

	// Stealth

	Security     defs.SkillID = "SECURITY"
	Sneak        defs.SkillID = "SNEAK"
	Speechcraft  defs.SkillID = "SPEECHCRAFT"
	Mercantile   defs.SkillID = "MERCANTILE"
	LightArmor   defs.SkillID = "LIGHT_ARMOR"
	Marksmanship defs.SkillID = "MARKSMANSHIP"
	ShortBlade   defs.SkillID = "SHORT_BLADE"

	// Magic

	Alchemy     defs.SkillID = "ALCHEMY"
	Incantation defs.SkillID = "INCANTATION"
)

var AllSkillIDs []defs.SkillID = []defs.SkillID{
	LongBlade, Blunt, Axe, Spear, Repair, HeavyArmor, Block,
	Security, Sneak, Speechcraft, Mercantile, LightArmor, Marksmanship, ShortBlade,
	Alchemy, Incantation,
}

func LevelSystemParameters() defs.LevelSystemParameters {
	var lvl defs.LevelSystemParameters = defs.LevelSystemParameters{
		MajorCount: 4,
		MinorCount: 4,

		MajorWeight: 3,
		MinorWeight: 2,
		MiscWeight:  1,

		MajorBase: 30,
		MinorBase: 20,
		MiscBase:  10,

		AttributeBase:   10,
		AttributeGrowth: 0.3,
		FavoredBonus:    10,
	}

	lvl.MiscCount = len(AllSkillIDs) - lvl.MajorCount - lvl.MinorCount

	lvl.MajorRate = 5.0 / float64(lvl.MajorCount)
	lvl.MinorRate = 3.0 / float64(lvl.MinorCount)
	lvl.MiscRate = 1.0 / float64(lvl.MiscCount)

	return lvl
}

func DefaultClasses() []defs.ClassDef {
	classes := []defs.ClassDef{
		{
			ID:                "legionary",
			Name:              "Legionary",
			FavoredAttributes: []defs.AttributeID{Endurance, Martial},
			SkillCategories: map[defs.SkillID]defs.SkillCategory{
				LongBlade:  skills.SkillCategoryMajor,
				HeavyArmor: skills.SkillCategoryMajor,
				Repair:     skills.SkillCategoryMajor,
				Block:      skills.SkillCategoryMajor,

				LightArmor:   skills.SkillCategoryMinor,
				Marksmanship: skills.SkillCategoryMinor,
				Spear:        skills.SkillCategoryMinor,
				ShortBlade:   skills.SkillCategoryMinor,
			},
			AboutMe: `
			I'm a soldier in the Roman Legion. We drill our formations day in and day out, and are quite handy with a gladius.
			I don't think you can find a finer armed force in all the lands under the Sun's chariot.
			`,
		},
		{
			ID:                "hoplite",
			Name:              "Hoplite",
			FavoredAttributes: []defs.AttributeID{Endurance, Martial},
			SkillCategories: map[defs.SkillID]defs.SkillCategory{
				Spear:      skills.SkillCategoryMajor,
				HeavyArmor: skills.SkillCategoryMajor,
				Block:      skills.SkillCategoryMajor,
				LightArmor: skills.SkillCategoryMajor,

				ShortBlade:  skills.SkillCategoryMinor,
				Alchemy:     skills.SkillCategoryMinor,
				Repair:      skills.SkillCategoryMinor,
				Speechcraft: skills.SkillCategoryMinor,
			},
			AboutMe: `
			I am Hoplite, armed with a spear and oft to put it to deadly effect. We train in the Greek style, same as our ancestors always have back to the sacking of Troy.
			When you think about it, what else do you really need than a spear? Nothing, if you ask me.
			`,
		},
		{
			ID:                "barbarian_raider",
			Name:              "Barbarian Raider",
			FavoredAttributes: []defs.AttributeID{Agility, Strength},
			SkillCategories: map[defs.SkillID]defs.SkillCategory{
				Axe:        skills.SkillCategoryMajor,
				Blunt:      skills.SkillCategoryMajor,
				LightArmor: skills.SkillCategoryMajor,
				Sneak:      skills.SkillCategoryMajor,

				Mercantile: skills.SkillCategoryMinor,
				HeavyArmor: skills.SkillCategoryMinor,
				Block:      skills.SkillCategoryMinor,
				Spear:      skills.SkillCategoryMinor,
			},
			AboutMe: `
			My people know little of the soft ways of the city dwellers. But they have gold, wine, and many things nice! So I say "don't mind if I do?" and smash their 
			doors down from time to time.
			`,
		},
		{
			ID:                "woodland_rogue",
			Name:              "Woodland Rogue",
			FavoredAttributes: []defs.AttributeID{Agility, Personality},
			SkillCategories: map[defs.SkillID]defs.SkillCategory{
				Marksmanship: skills.SkillCategoryMajor,
				LightArmor:   skills.SkillCategoryMajor,
				Alchemy:      skills.SkillCategoryMajor,
				Sneak:        skills.SkillCategoryMajor,

				ShortBlade:  skills.SkillCategoryMinor,
				Mercantile:  skills.SkillCategoryMinor,
				Security:    skills.SkillCategoryMinor,
				Speechcraft: skills.SkillCategoryMinor,
			},
			AboutMe: `
			You ask what I do? Well, if you were a rich traveler loaded down with sacks of gold and wandering cluelessly in a forest, you might well know the answer already.
			Ha! I'm joking of course.
			`,
		},
		{
			ID:                "hunter",
			Name:              "Hunter",
			FavoredAttributes: []defs.AttributeID{Agility, Endurance},
			SkillCategories: map[defs.SkillID]defs.SkillCategory{
				Marksmanship: skills.SkillCategoryMajor,
				LightArmor:   skills.SkillCategoryMajor,
				Sneak:        skills.SkillCategoryMajor,
				Spear:        skills.SkillCategoryMajor,

				Repair:     skills.SkillCategoryMinor,
				Axe:        skills.SkillCategoryMinor,
				Alchemy:    skills.SkillCategoryMinor,
				ShortBlade: skills.SkillCategoryMinor,
			},
			AboutMe: `
			I'm a hunter by trade; I stalk the forests and have taken down many a fearsome beast. There have been some close calls - you don't want to go toe to toe with a wild
			boar if you can avoid it. Luckily, my spear hasn't missed its mark yet.
			`,
		},
		{
			ID:                "frumentarius",
			Name:              "Frumentarius",
			FavoredAttributes: []defs.AttributeID{Intrigue, Personality},
			SkillCategories: map[defs.SkillID]defs.SkillCategory{
				Speechcraft: skills.SkillCategoryMajor,
				Security:    skills.SkillCategoryMajor,
				Sneak:       skills.SkillCategoryMajor,
				LongBlade:   skills.SkillCategoryMajor,

				Mercantile:   skills.SkillCategoryMinor,
				Block:        skills.SkillCategoryMinor,
				HeavyArmor:   skills.SkillCategoryMinor,
				Marksmanship: skills.SkillCategoryMinor,
			},
			// TODO: I'm guessing we will want different answers depending on opinion levels.
			// for example, for low opinion, they would give evasive or short answers; for higher opinions, a more real answer.
			AboutMe: `
			You're a very curious sort, aren't you? Let's just say, it's my business to know the business of others, and not to answer questions like yours.
			`,
		},
		{
			ID:                "watchman",
			Name:              "Watchman",
			FavoredAttributes: []defs.AttributeID{Martial, Intrigue},
			SkillCategories: map[defs.SkillID]defs.SkillCategory{
				Spear:      skills.SkillCategoryMajor,
				Blunt:      skills.SkillCategoryMajor,
				HeavyArmor: skills.SkillCategoryMajor,
				Block:      skills.SkillCategoryMajor,

				Marksmanship: skills.SkillCategoryMinor,
				Mercantile:   skills.SkillCategoryMinor,
				LightArmor:   skills.SkillCategoryMinor,
				Security:     skills.SkillCategoryMinor,
			},
			AboutMe: `
			I'm the only one around here making sure this place doesn't get overrun with criminals or burn to the ground!
			I deal with the crooks, drunks, thieves, and most of the undesirable elements society has to offer.
			`,
		},
	}

	for _, class := range classes {
		// fill in misc skills
		for _, skillID := range AllSkillIDs {
			if _, exists := class.SkillCategories[skillID]; !exists {
				class.SkillCategories[skillID] = skills.SkillCategoryMisc
			}
		}
	}

	return classes
}

func GetAllAttributes() []defs.AttributeDef {
	return []defs.AttributeDef{
		{
			ID:          Strength,
			DisplayName: "Strength",
			Description: "Governs how strong you are, how much you can carry in your inventory, and how mighty a weapon you can wield.",
		},
		{
			ID:          Endurance,
			DisplayName: "Endurance",
			Description: "Governs your total health and how much hardship you can endure.",
		},
		{
			ID:          Martial,
			DisplayName: "Martial",
			Description: "Governs how skilled you are in the arts of combat and the strategy of war.",
		},
		{
			ID:          Agility,
			DisplayName: "Agility",
			Description: "Governs how swiftly you move on your feet and how carefully you can maneuver.",
		},
		{
			ID:          Intelligence,
			DisplayName: "Intelligence",
			Description: "Governs how learned you are in worldly and academic disciplines, and also your knowledge of things arcane and otherworldly.",
		},
		{
			ID:          Intrigue,
			DisplayName: "Intrigue",
			Description: "Governs your ability to move quietly and unseen, uncover plots, and form plots of your own.",
		},
		{
			ID:          Personality,
			DisplayName: "Personality",
			Description: "Governs how likeable you are to the people around you.",
		},
		{
			ID:          Luck,
			DisplayName: "Luck",
			Description: "Governs how likely you are to narrowly avoid death, defy the will of the gods, or find a gold coin.",
		},
	}
}

func GetAllSkills() []defs.SkillDef {
	return []defs.SkillDef{
		// Combat
		{
			ID:                  LongBlade,
			DisplayName:         "Long Blade",
			GoverningAttributes: []defs.AttributeID{Martial, Strength},
			Description:         "Your swordsmanship with a long blade, including anything from a gladius up to a two-handed claymore.",
		},
		{
			ID:                  Blunt,
			DisplayName:         "Blunt",
			GoverningAttributes: []defs.AttributeID{Strength, Endurance},
			Description:         "How well you can wield a large club or mace.",
		},
		{
			ID:                  Axe,
			DisplayName:         "Axe",
			GoverningAttributes: []defs.AttributeID{Strength, Endurance},
			Description:         "Your skill with an axe, for wood-chopping or head-splitting.",
		},
		{
			ID:                  Spear,
			DisplayName:         "Spear",
			GoverningAttributes: []defs.AttributeID{Martial, Endurance},
			Description:         "How talented you are with a spear or lance.",
		},
		{
			ID:                  Repair, // TODO: rename this to Smithing? could be interesting to make this skill more applicable to other things.
			DisplayName:         "Repair",
			GoverningAttributes: []defs.AttributeID{Martial, Intelligence},
			Description:         "Your ability to mend damaged armor, weapons, or just about anything a hammer, anvil and prongs can fix.",
		},
		{
			ID:                  HeavyArmor,
			DisplayName:         "Heavy Armor",
			GoverningAttributes: []defs.AttributeID{Endurance, Strength},
			Description:         "How well you wear a heavy suit of bronze, iron, or steel plate armor.",
		},
		{
			ID:                  Block,
			DisplayName:         "Block",
			GoverningAttributes: []defs.AttributeID{Endurance, Martial},
			Description:         "Your discipline with a shield in tactical formations, as well as your ability to withstand a crushing enemy blow.",
		},

		// Stealth
		{
			ID:                  Security,
			DisplayName:         "Security",
			GoverningAttributes: []defs.AttributeID{Intrigue, Intelligence},
			Description:         "Your skill in tinkering with and disarming locks, traps, and defensive measures of all kinds.",
		},
		{
			ID:                  Sneak,
			DisplayName:         "Sneak",
			GoverningAttributes: []defs.AttributeID{Intrigue, Agility},
			Description:         "Your ability to blend into the shadows and slip by unseen.",
		},
		{
			ID:                  Speechcraft,
			DisplayName:         "Speechcraft",
			GoverningAttributes: []defs.AttributeID{Personality, Intrigue},
			Description:         "How well you command words and inspire, persuade, or convince others.",
		},
		{
			ID:                  Mercantile,
			DisplayName:         "Mercantile",
			GoverningAttributes: []defs.AttributeID{Personality, Intelligence},
			Description:         "Your knowledge of money and finance, and your skill at bargaining to gain the best prices and deals.",
		},
		{
			ID:                  LightArmor,
			DisplayName:         "Light Armor",
			GoverningAttributes: []defs.AttributeID{Agility, Luck},
			Description:         "How well you maneuver in lighter armor, trading protection for mobility.",
		},
		{
			ID:                  Marksmanship,
			DisplayName:         "Marksmanship",
			GoverningAttributes: []defs.AttributeID{Agility, Martial},
			Description:         "Your skill with ranged weaponry, from the bow, to the javelin, to the sling.",
		},
		{
			ID:                  ShortBlade,
			DisplayName:         "Short Blade",
			GoverningAttributes: []defs.AttributeID{Agility, Intrigue},
			Description:         "How well you can handle a knife or dagger at close range.",
		},

		// Magic
		{
			ID:                  Alchemy, // TODO: Should this have a connotation with medicine, too? Like, your knowledge can help you heal yourself or others (not just with potions)
			DisplayName:         "Alchemy",
			GoverningAttributes: []defs.AttributeID{Intelligence, Intrigue},
			Description:         "Your familiarity with ingredients and herbs of all manner, and how to put them to good use.",
		},
		{
			ID:                  Incantation,
			DisplayName:         "Incantation",
			GoverningAttributes: []defs.AttributeID{Intelligence, Luck},
			Description:         "How in tune you are with the supernatural and the divine, and your knowledge of the world of magic.",
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
