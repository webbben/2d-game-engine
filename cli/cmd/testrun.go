package cmd

import (
	"fmt"
	"os"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/spf13/cobra"
	"github.com/webbben/2d-game-engine/definitions"
	"github.com/webbben/2d-game-engine/dialog"
	"github.com/webbben/2d-game-engine/entity"
	"github.com/webbben/2d-game-engine/entity/npc"
	"github.com/webbben/2d-game-engine/entity/player"
	"github.com/webbben/2d-game-engine/game"
	"github.com/webbben/2d-game-engine/internal/audio"
	"github.com/webbben/2d-game-engine/internal/config"
	"github.com/webbben/2d-game-engine/internal/display"
	"github.com/webbben/2d-game-engine/internal/image"
	"github.com/webbben/2d-game-engine/internal/model"
	"github.com/webbben/2d-game-engine/internal/ui/textwindow"
	"github.com/webbben/2d-game-engine/inventory"
	"github.com/webbben/2d-game-engine/item"
	playermenu "github.com/webbben/2d-game-engine/playerMenu"
	"github.com/webbben/2d-game-engine/trade"
)

// testrunCmd represents the testrun command
var testrunCmd = &cobra.Command{
	Use:   "testrun",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		display.SetupGameDisplay("Ancient Rome!", false)

		// set config
		config.ShowPlayerCoords = true
		config.ShowGameDebugInfo = true
		//config.DrawGridLines = true
		//config.ShowEntityPositions = true
		//config.TrackMemoryUsage = true
		//config.HourSpeed = time.Second * 20
		//config.ShowCollisions = true
		//config.ShowNPCPaths = true

		config.GameDataPathOverride = "/Users/benwebb/dev/personal/ancient-rome"

		config.DefaultFont = image.LoadFont("ashlander-pixel.ttf", 22, 0)
		config.DefaultTitleFont = image.LoadFont("ashlander-pixel.ttf", 28, 0)

		config.DefaultTooltipBox = config.DefaultBox{
			TilesetSrc:  "boxes/boxes.tsj",
			OriginIndex: 132,
		}
		config.DefaultUIBox = config.DefaultBox{
			TilesetSrc:  "boxes/boxes.tsj",
			OriginIndex: 16,
		}

		err := game.InitialStartUp()
		if err != nil {
			panic(err)
		}

		// get our testrun game state
		gameState := setupGameState()

		LoadItems(gameState)

		gameState.Player.AddItemToInventory(gameState.DefinitionManager.NewInventoryItem("longsword_01", 1))
		gameState.Player.AddItemToInventory(gameState.DefinitionManager.NewInventoryItem("potion_herculean_strength", 2))
		gameState.Player.AddItemToInventory(gameState.DefinitionManager.NewInventoryItem("currency_value_1", 4))
		gameState.Player.AddItemToInventory(gameState.DefinitionManager.NewInventoryItem("currency_value_10", 6))
		gameState.Player.AddItemToInventory(gameState.DefinitionManager.NewInventoryItem("currency_value_100", 1))
		gameState.Player.AddItemToInventory(gameState.DefinitionManager.NewInventoryItem("currency_value_1000", 1))

		fmt.Println("player inventory:", gameState.Player.Entity.InventoryItems)

		gameState.PlayerMenu.InventoryPage.SyncPlayerItems()

		shopKeeperInventory := []item.InventoryItem{}
		shopKeeperInventory = append(shopKeeperInventory, gameState.DefinitionManager.NewInventoryItem("longsword_01", 1))
		shopkeeper := definitions.NewShopKeeper(1200, "Aurelius' Tradehouse", shopKeeperInventory)
		gameState.DefinitionManager.LoadShopkeeper("aurelius_tradehouse", shopkeeper)

		gameState.DefinitionManager.LoadDialog("dialog1", GetDialog())

		if err := gameState.RunGame(); err != nil {
			panic(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(testrunCmd)
}

func setupGameState() *game.Game {
	g := game.NewGame(10)
	err := g.SetupMap("village_surano", &game.OpenMapOptions{
		RunNPCManager:    true,
		RegenerateImages: true,
	})
	if err != nil {
		panic(err)
	}

	footstepSFX := entity.AudioProps{
		FootstepSFX: audio.FootstepSFX{
			StepDefaultSrc: []string{
				"sfx/footsteps/footstep_stone_01_A.mp3",
				"sfx/footsteps/footstep_stone_01_B.mp3",
			},
			StepWoodSrc: []string{
				"sfx/footsteps/footstep_wood_01_A.mp3",
				"sfx/footsteps/footstep_wood_01_B.mp3",
			},
			StepGrassSrc: []string{
				"sfx/footsteps/footstep_grass_01_A.mp3",
				"sfx/footsteps/footstep_grass_01_B.mp3",
			},
		},
	}

	// make the player
	playerEnt := entity.NewEntity(entity.GeneralProps{
		DisplayName:   "Caius Cosades",
		EntityBodySrc: "/Users/benwebb/dev/personal/ancient-rome/src/data/characters/json/character_01.json",
		IsPlayer:      true,
		InventorySize: 18,
	}, entity.MovementProps{
		WalkSpeed: 0,
	}, footstepSFX)

	p := player.NewPlayer(g.DefinitionManager, &playerEnt)

	g.PlacePlayerAtSpawnPoint(&p, 0)

	npcEnt := entity.NewEntity(entity.GeneralProps{
		DisplayName:   "Legionary",
		EntityBodySrc: "/Users/benwebb/dev/personal/ancient-rome/src/data/characters/json/character_02.json",
		InventorySize: 18,
	}, entity.MovementProps{
		WalkSpeed: 0,
	}, footstepSFX)
	n := npc.New(npc.NPC{
		Entity: &npcEnt,
		NPCInfo: npc.NPCInfo{
			DisplayName: npcEnt.DisplayName,
		},
		DialogID: "dialog1",
	})

	err = n.SetFightTask(&playerEnt, false)
	if err != nil {
		panic(err)
	}

	g.MapInfo.AddNPCToMap(&n, model.Coords{X: 0, Y: 0})

	// setup the game struct
	g.Player = &p

	// add my test key bindings
	addCustomKeyBindings(g)

	g.PlayerMenu = GetPlayerMenu(g.Player, g.DefinitionManager)
	g.TradeScreen = GetTradeScreen(g.Player, g.DefinitionManager)

	return g
}

func addCustomKeyBindings(g *game.Game) {
	// open a test dialog
	g.SetGlobalKeyBinding(ebiten.KeyEqual, func(gg *game.Game) {
		// doing this async since we are loading an image file
		go func() {
			fmt.Println("getting dialog")
			d := GetDialog()
			gg.Dialog = &d
		}()
	})
	g.SetGlobalKeyBinding(ebiten.KeyMinus, func(gg *game.Game) {
		go func() {
			fmt.Println("toggle player menu")
			showPlayerMenu := !gg.ShowPlayerMenu
			if showPlayerMenu {
				gg.PlayerMenu.InventoryPage.SyncPlayerItems()
			} else {
				gg.PlayerMenu.InventoryPage.SavePlayerInventory()
			}
			gg.ShowPlayerMenu = showPlayerMenu
			fmt.Println(gg.Player.Entity.InventoryItems)
		}()
	})
	g.SetGlobalKeyBinding(ebiten.Key0, func(gg *game.Game) {
		go func() {
			fmt.Println("toggle trade screen")
			g.SetupTradeSession("aurelius_tradehouse")
		}()
	})

	g.SetGlobalKeyBinding(ebiten.KeyEscape, func(gg *game.Game) {
		os.Exit(0)
	})
}

func GetPlayerMenu(p *player.Player, defMgr *definitions.DefinitionManager) playermenu.PlayerMenu {
	pm := playermenu.PlayerMenu{
		BoxTilesetSource:      "boxes/boxes.tsj",
		PageTabsTilesetSource: "ui/ui-components.tsj",
		BoxOriginIndex:        16,
		BoxTitleOriginIndex:   111,
	}

	invParams := inventory.InventoryParams{
		ItemSlotTilesetSource:    "ui/ui-components.tsj",
		SlotEnabledTileID:        0,
		SlotDisabledTileID:       1,
		SlotEquipedBorderTileID:  3,
		SlotSelectedBorderTileID: 4,
		HoverWindowParams: textwindow.TextWindowParams{
			TilesetSource:   "boxes/boxes.tsj",
			OriginTileIndex: 20,
		},
		RowCount:          10,
		ColCount:          9,
		EnabledSlotsCount: 18,
	}

	pm.Load(p, defMgr, invParams)

	return pm
}

func GetTradeScreen(p *player.Player, defMgr *definitions.DefinitionManager) trade.TradeScreen {
	invParams := inventory.InventoryParams{
		ItemSlotTilesetSource:    "ui/ui-components.tsj",
		SlotEnabledTileID:        0,
		SlotDisabledTileID:       1,
		SlotEquipedBorderTileID:  3,
		SlotSelectedBorderTileID: 4,
		HoverWindowParams: textwindow.TextWindowParams{
			TilesetSource:   "boxes/boxes.tsj",
			OriginTileIndex: 20,
		},
		RowCount:          10,
		ColCount:          9,
		EnabledSlotsCount: 18,
	}
	shopKeeperInvParams := invParams
	shopKeeperInvParams.EnabledSlotsCount = 90

	ts := trade.NewTradeScreen(trade.TradeScreenParams{
		BoxTilesetSrc:             "boxes/boxes.tsj",
		BoxTilesetOrigin:          16,
		BoxTitleOrigin:            111,
		ShopkeeperInventoryParams: shopKeeperInvParams,
		PlayerInventoryParams:     invParams,
		TextBoxTilesetSrc:         "boxes/boxes.tsj",
		TextBoxOrigin:             135,
	}, defMgr, p)

	return ts
}

func GetDialog() dialog.Dialog {
	d := dialog.Dialog{
		BoxTilesetSource:   "boxes/boxes.tsj",
		BoxOriginTileIndex: 16,
		TextFont: dialog.Font{
			Source: "ashlander-pixel.ttf",
		},
		TopicsEnabled: true,
	}
	rootTopic := dialog.Topic{
		TopicText: "Root",
		MainText: dialog.TextBranch{
			Text: "Hello! Welcome to the Magical Goods Emporium. All of these items were acquired in distant lands such as Aegyptus or Indus. I assure you that you'll find nothing like this anywhere else in Rome.",
		},
		ReturnText: "Anything else I can help you with?",
	}
	rootTopic.SubTopics = append(rootTopic.SubTopics, dialog.Topic{
		TopicText: "Rumors",
		MainText: dialog.TextBranch{
			Text: "They say if you go to the Forum past midnight, you might find a group of shady individuals hanging around in the dark. Not sure what for, but I'd also imagine it's a bad idea to go snooping around for them.",
		},
	})
	rootTopic.SubTopics = append(rootTopic.SubTopics, dialog.Topic{
		TopicText: "The Empire",
		MainText: dialog.TextBranch{
			Text: "The Empire spans the world over - they say all the peoples from the foggy isles of Britain to the Nile of Egypt all are under Imperial rule.",
		},
	})
	rootTopic.SubTopics = append(rootTopic.SubTopics, dialog.Topic{
		TopicText:    "Trade",
		ShopkeeperID: "aurelius_tradehouse",
	})

	jokeTopic := dialog.Topic{
		TopicText: "Tell me a joke",
		MainText: dialog.TextBranch{
			Text: "A joke? Alright, how about this one:\nWhy did the chicken cross the road?",
			Options: []dialog.TextBranch{
				{
					OptionText: "To get to the other side?",
					Text:       "No stupid! He was running away from a Yakitori chef!",
				},
				{
					OptionText: "I don't know, why?",
					Text:       "Come on, not even a guess?",
				},
			},
		},
	}
	rootTopic.SubTopics = append(rootTopic.SubTopics, jokeTopic)

	rootTopic.SubTopics = append(rootTopic.SubTopics, dialog.Topic{
		TopicText: "Lorem Ipsum",
		MainText: dialog.TextBranch{
			Text: `
Lorem ipsum dolor sit amet, consectetur adipiscing elit. Vestibulum suscipit urna ex, laoreet gravida risus blandit consectetur. Nullam pulvinar, enim et commodo fringilla, nulla magna tempor enim, quis elementum est mauris at mauris. Nam non ligula a enim sollicitudin luctus. Sed aliquet maximus erat aliquam iaculis. In tempus sapien nisi. Etiam tortor massa, tristique nec ex in, imperdiet dignissim nisi. Vivamus id mi at dolor suscipit luctus. In nec lacus et elit rhoncus cursus. Sed porttitor, dui eu ornare fringilla, dui risus placerat eros, nec porta sem justo sit amet neque. Lorem ipsum dolor sit amet, consectetur adipiscing elit. Cras vel congue tortor. Mauris aliquet molestie massa, venenatis volutpat justo convallis vestibulum.
Integer nisi ligula, volutpat feugiat eros ut, cursus eleifend est. Quisque gravida sit amet dui vitae pellentesque. Morbi interdum facilisis tellus aliquam egestas. Nunc posuere nunc neque, a sagittis elit ultricies eget. Aliquam vel dignissim dui. Quisque mollis massa nibh, id dignissim ante semper eu. Class aptent taciti sociosqu ad litora torquent per conubia nostra, per inceptos himenaeos. Vivamus sed accumsan felis.
Sed velit neque, eleifend quis arcu sed, fringilla varius augue. Aenean interdum ornare consectetur. Mauris eleifend mauris erat, non luctus nunc venenatis at. Aliquam ultricies dolor sed odio iaculis, id gravida ipsum faucibus. Morbi vitae rutrum nisl. Praesent lorem leo, tincidunt eu felis quis, ornare blandit nisl. Class aptent taciti sociosqu ad litora torquent per conubia nostra, per inceptos himenaeos. Nulla convallis diam sed elementum posuere. Morbi vulputate urna vitae quam gravida pellentesque. Pellentesque habitant morbi tristique senectus et netus et malesuada fames ac turpis egestas. Maecenas a tincidunt turpis, sed luctus enim. Sed venenatis quam non velit viverra ultricies. Maecenas tristique lacinia mauris id interdum. Suspendisse tempus enim arcu, id fringilla mauris consectetur quis. Vestibulum placerat ante lacus, sed tempor mi pharetra ut. Aenean metus augue, vestibulum in tincidunt sed, tempus nec lacus.
Quisque sollicitudin auctor magna, a pharetra justo consequat sed. Nulla mi leo, ultricies in neque ac, elementum posuere neque. Praesent et maximus est. Etiam pulvinar velit a felis bibendum molestie. Donec faucibus mi in elit dapibus fermentum. Quisque vestibulum libero quis lacus tincidunt volutpat. Nullam posuere mauris odio, vitae venenatis tortor sodales ac. Donec porta massa eu vehicula dapibus. Phasellus vulputate placerat urna, nec feugiat est porta sed. Curabitur ac turpis sem. Morbi nisi turpis, dignissim eu nisi at, mollis posuere lorem. Maecenas pretium congue lectus, ut tempus dolor ornare vitae.
In et aliquet orci. Curabitur pharetra sit amet felis et faucibus. Morbi vitae massa quam. Aliquam porta, nulla quis egestas lobortis, magna diam ultrices felis, a scelerisque justo diam a est. Vestibulum nisi leo, placerat ut laoreet vel, iaculis id sapien. Ut et placerat lectus. Aliquam erat volutpat. Fusce finibus sapien quis justo lobortis feugiat.
			`,
		},
	})

	d.RootTopic = rootTopic

	return d
}

func LoadItems(g *game.Game) {
	itemDefs := []item.ItemDef{
		&item.WeaponDef{
			ItemBase: item.ItemBase{
				ID:                   "longsword_01",
				Name:                 "Iron Longsword",
				Description:          "An iron longsword forged by blacksmiths in Gaul.",
				Value:                100,
				Weight:               25,
				MaxDurability:        250,
				TilesetSourceTileImg: "items/items_01.tsj",
				TileImgIndex:         0,
				MiscItem:             true,
			},
			Damage:        10,
			HitsPerSecond: 1,
		},
		&item.ItemBase{
			ID:                   "potion_herculean_strength",
			Name:                 "Potion of Herculean Strength",
			Description:          "This potion invigorates the drinker and gives him strength only matched by Hercules himself.",
			Value:                200,
			Weight:               3,
			TilesetSourceTileImg: "items/items_01.tsj",
			TileImgIndex:         129,
			Groupable:            true,
			Consumable:           true,
		},
		&item.ItemBase{
			ID:                   "currency_value_1",
			Name:                 "Aes",
			Description:          "A Roman bronze coin",
			Value:                1,
			Weight:               0.05,
			TilesetSourceTileImg: "items/items_01.tsj",
			TileImgIndex:         64,
			Groupable:            true,
			Currency:             true,
		},
		&item.ItemBase{
			ID:                   "currency_value_5",
			Name:                 "Dupondius",
			Description:          "A Roman brass coin",
			Value:                5,
			Weight:               0.05,
			TilesetSourceTileImg: "items/items_01.tsj",
			TileImgIndex:         65,
			Groupable:            true,
			Currency:             true,
		},
		&item.ItemBase{
			ID:                   "currency_value_10",
			Name:                 "Sestertius",
			Description:          "A Roman brass coin",
			Value:                10,
			Weight:               0.05,
			TilesetSourceTileImg: "items/items_01.tsj",
			TileImgIndex:         66,
			Groupable:            true,
			Currency:             true,
		},
		&item.ItemBase{
			ID:                   "currency_value_50",
			Name:                 "Quinarius",
			Description:          "A Roman silver coin",
			Value:                50,
			Weight:               0.05,
			TilesetSourceTileImg: "items/items_01.tsj",
			TileImgIndex:         67,
			Groupable:            true,
			Currency:             true,
		},
		&item.ItemBase{
			ID:                   "currency_value_100",
			Name:                 "Denarius",
			Description:          "A Roman silver coin",
			Value:                100,
			Weight:               0.05,
			TilesetSourceTileImg: "items/items_01.tsj",
			TileImgIndex:         68,
			Groupable:            true,
			Currency:             true,
		},
		&item.ItemBase{
			ID:                   "currency_value_1000",
			Name:                 "Aureus",
			Description:          "A Roman gold coin",
			Value:                1000,
			Weight:               0.05,
			TilesetSourceTileImg: "items/items_01.tsj",
			TileImgIndex:         69,
			Groupable:            true,
			Currency:             true,
		},
		// &item.ArmorDef{
		// 	ItemBase: item.ItemBase{
		// 		ID:                   "legionary_helm",
		// 		Name:                 "Legionary Helm",
		// 		Description:          "A standard issue steel helmet for Roman legionaries.",
		// 		Value:                250,
		// 		Weight:               15,
		// 		TilesetSourceTileImg: "assets/tiled/tilesets/items_01.tsj",
		// 		TileImgIndex:         32,
		// 		Armor:                true,
		// 	},
		// 	Protection: 10,
		// },
		// &item.ArmorDef{
		// 	ItemBase: item.ItemBase{
		// 		ID:                   "legionary_cuirass",
		// 		Name:                 "Legionary Cuirass",
		// 		Description:          "A set of Lorica Segmentata body armor, used by Roman legionaries.",
		// 		Value:                700,
		// 		Weight:               25,
		// 		TilesetSourceTileImg: "assets/tiled/tilesets/items_01.tsj",
		// 		TileImgIndex:         33,
		// 		Armor:                true,
		// 	},
		// 	Protection: 18,
		// },
		// &item.ArmorDef{
		// 	ItemBase: item.ItemBase{
		// 		ID:                   "caligae_boots",
		// 		Name:                 "Caligae",
		// 		Description:          "A pair of caligae, heavy leather sandals commonly worn by Roman soldiers.",
		// 		Value:                15,
		// 		Weight:               4,
		// 		TilesetSourceTileImg: "assets/tiled/tilesets/items_01.tsj",
		// 		TileImgIndex:         33,
		// 		Armor:                true,
		// 	},
		// 	Protection: 5,
		// },
	}

	g.DefinitionManager.LoadItemDefs(itemDefs)
}
