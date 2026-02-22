// Package cmd - just some commands to run for testing
package cmd

import (
	"fmt"
	"os"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/spf13/cobra"
	"github.com/webbben/2d-game-engine/definitions"
	"github.com/webbben/2d-game-engine/entity"
	"github.com/webbben/2d-game-engine/entity/npc"
	"github.com/webbben/2d-game-engine/entity/player"
	"github.com/webbben/2d-game-engine/game"
	"github.com/webbben/2d-game-engine/internal/display"
	"github.com/webbben/2d-game-engine/internal/model"
	"github.com/webbben/2d-game-engine/internal/ui/textwindow"
	"github.com/webbben/2d-game-engine/inventory"
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

		SetConfig()

		err := game.InitialStartUp()
		if err != nil {
			panic(err)
		}

		// get our testrun game state
		gameState := SetupGameState(gameParams{
			startMapID: "village_surano",
			startHour:  12,
		})

		// add a test NPC

		charStateID := entity.CreateNewCharacterState("character_02", entity.NewCharacterStateParams{}, gameState.DefinitionManager)

		n := npc.NewNPC(npc.NPCParams{
			CharStateID: charStateID,
		},
			gameState.DefinitionManager, gameState.AudioManager, gameState.EventBus)

		gameState.MapInfo.AddNPCToMap(n, model.Coords{X: 0, Y: 0})

		// Load player inventory page

		if err := gameState.RunGame(); err != nil {
			panic(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(testrunCmd)
}

func addCustomKeyBindings(g *game.Game) {
	g.SetGlobalKeyBinding(ebiten.KeyMinus, func(gg *game.Game) {
		go func() {
			fmt.Println("toggle player menu")
			showPlayerMenu := !gg.ShowPlayerMenu
			if showPlayerMenu {
				gg.PlayerMenu.InventoryPage.LoadPlayerItemsIn()
			} else {
				gg.PlayerMenu.InventoryPage.SaveAndClose()
			}
			gg.ShowPlayerMenu = showPlayerMenu
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
