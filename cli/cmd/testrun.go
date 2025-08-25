package cmd

import (
	"fmt"
	"image/color"
	"log"
	"os"
	"path/filepath"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/spf13/cobra"
	"github.com/webbben/2d-game-engine/entity"
	g "github.com/webbben/2d-game-engine/game"
	"github.com/webbben/2d-game-engine/internal/config"
	"github.com/webbben/2d-game-engine/internal/dialog"
	"github.com/webbben/2d-game-engine/internal/model"
	"github.com/webbben/2d-game-engine/internal/tiled"
	"github.com/webbben/2d-game-engine/npc"
	"github.com/webbben/2d-game-engine/player"
	"github.com/webbben/2d-game-engine/screen"
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
		ebiten.SetWindowSize(config.ScreenWidth, config.ScreenHeight)
		ebiten.SetWindowTitle(config.WindowTitle)

		tiled.InitFileStructure()

		// get our testrun game state
		game := setupGameState()

		// set config
		config.ShowPlayerCoords = true

		if err := ebiten.RunGame(game); err != nil {
			panic(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(testrunCmd)
}

func setupGameState() *g.Game {
	currentMap, err := tiled.OpenMap("assets/tiled/maps/testmap.tmj")
	if err != nil {
		log.Fatal(err)
	}

	// setup the map
	err = currentMap.Load()
	if err != nil {
		log.Fatal(err)
	}
	mapInfo := g.MapInfo{
		Map: currentMap,
	}

	// make the player
	playerEnt, err := entity.OpenEntity(filepath.Join(config.GameDefsPath(), "ent", "ent_750fde30-4e5a-41ce-96e3-0105e0064a4d.json"))
	if err != nil {
		log.Fatal(err)
	}
	playerEnt.IsPlayer = true
	err = playerEnt.Load()
	if err != nil {
		log.Fatal(err)
	}
	p := player.Player{
		Entity: &playerEnt,
	}

	mapInfo.AddPlayerToMap(&p, model.Coords{X: 5, Y: 5})

	// make NPCs

	for i := 0; i < 1; i++ {
		npcEnt := playerEnt.Duplicate()
		npcEnt.DisplayName = fmt.Sprintf("NPC_%v", i)
		err = npcEnt.Load()
		if err != nil {
			log.Fatal(err)
		}
		n := npc.New(npc.NPC{
			Entity: &npcEnt,
			NPCInfo: npc.NPCInfo{
				DisplayName: npcEnt.DisplayName,
			},
		})

		n.SetFollowTask(&playerEnt, 0)

		mapInfo.AddNPCToMap(&n, model.Coords{X: i, Y: 0})
	}

	// setup the game struct
	game := &g.Game{
		MapInfo: mapInfo,
		Player:  p,
	}

	// add my test key bindings
	addCustomKeyBindings(game)

	return game
}

func addCustomKeyBindings(game *g.Game) {
	// open a test dialog
	game.SetGlobalKeyBinding(ebiten.KeyEqual, func(g *g.Game) {
		// doing this async since we are loading an image file
		go func() {
			fmt.Println("getting dialog")
			c := GetConversation()
			g.Conversation = &c
		}()
	})
	game.SetGlobalKeyBinding(ebiten.KeyMinus, func(g *g.Game) {
		go func() {
			fmt.Println("getting title screen")
			s := GetTitleScreen()
			g.CurrentScreen = &s
		}()
	})

	game.SetGlobalKeyBinding(ebiten.KeyEscape, func(g *g.Game) {
		os.Exit(0)
	})
}

func GetConversation() dialog.Conversation {
	d := dialog.Dialog{
		Steps: []dialog.DialogStep{
			{Text: "Greetings, what can I do for you?"},
		},
	}

	c := dialog.Conversation{
		Greeting: d,
		Font: dialog.Font{
			FontName: "Planewalker",
		},
		Topics: map[string]dialog.Dialog{
			"rumors":        {Steps: []dialog.DialogStep{{Text: "I heard there are goblins in the forest."}}},
			"little advice": {Steps: []dialog.DialogStep{{Text: "Don't go into the forest alone."}}},
			"joke":          {Steps: []dialog.DialogStep{{Text: "Why did the chicken cross the road?"}, {Text: "To get to the other side!"}}},
			"the empire":    {Steps: []dialog.DialogStep{{Text: "The empire is a vast and powerful entity."}}},
		},
	}
	c.SetDialogTiles("tileset/borders/dialog_1")

	return c
}

func GetTitleScreen() screen.Screen {
	s := screen.Screen{
		Title:               "Ancient Rome!",
		TitleFontName:       "Herculanum",
		TitleFontColor:      color.White,
		BodyFontName:        "Herculanum",
		BodyFontColor:       color.White,
		BackgroundImagePath: "image/bg/dark_cistern.png",
	}

	// add a menu
	m := screen.Menu{
		Buttons: []screen.Button{
			{Text: "New Game", Callback: func() {}},
			{Text: "Load Game", Callback: func() {}},
			{Text: "Options", Callback: func() {}},
			{Text: "Quit", Callback: func() { os.Exit(0) }},
		},
		BoxTilesetPath: "tileset/borders/stone_1",
	}
	s.Menus = append(s.Menus, m)

	return s
}
