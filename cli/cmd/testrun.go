package cmd

import (
	"fmt"
	"image/color"
	"log"
	"os"
	"path/filepath"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/spf13/cobra"
	"github.com/webbben/2d-game-engine/dialog"
	"github.com/webbben/2d-game-engine/entity"
	"github.com/webbben/2d-game-engine/game"
	"github.com/webbben/2d-game-engine/internal/audio"
	"github.com/webbben/2d-game-engine/internal/config"
	"github.com/webbben/2d-game-engine/internal/display"
	"github.com/webbben/2d-game-engine/internal/image"
	"github.com/webbben/2d-game-engine/internal/lights"
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
		display.SetupGameDisplay("Ancient Rome!", false)

		err := lights.LoadShaders()
		if err != nil {
			log.Fatal("error loading shaders: ", err)
		}

		tiled.InitFileStructure()

		// get our testrun game state
		game := setupGameState()

		// set config
		config.ShowPlayerCoords = true
		config.ShowGameDebugInfo = true
		config.DrawGridLines = true
		config.ShowEntityPositions = true
		//config.TrackMemoryUsage = true
		//config.HourSpeed = time.Second * 20

		config.DefaultFont = image.LoadFont("assets/fonts/ashlander-pixel.ttf", 0, 0)

		if err := ebiten.RunGame(game); err != nil {
			panic(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(testrunCmd)
}

func setupGameState() *game.Game {
	g := game.NewGame(10)
	g.SetupMap("assets/tiled/maps/surano.tmj", game.OpenMapOptions{
		RunNPCManager: true,
	})

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
	playerEnt.LoadFootstepSFX(audio.FootstepSFX{
		StepDefaultSrc: []string{
			"/Users/benwebb/dev/personal/ancient-rome/assets/audio/sfx/footsteps/footstep_stone_01_A.mp3",
			"/Users/benwebb/dev/personal/ancient-rome/assets/audio/sfx/footsteps/footstep_stone_01_B.mp3",
		},
	})
	p := player.Player{
		Entity: &playerEnt,
	}

	g.MapInfo.AddPlayerToMap(&p, model.Coords{X: 5, Y: 5})

	// make NPCs
	legionaryEnt, err := entity.OpenEntity(filepath.Join(config.GameDefsPath(), "ent", "ent_6ef9b0ec-8e34-4ebf-a9da-e04ef154e80b.json"))
	if err != nil {
		log.Fatal(err)
	}

	for i := 0; i < 0; i++ {
		npcEnt := legionaryEnt.Duplicate()
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

		g.MapInfo.AddNPCToMap(&n, model.Coords{X: i, Y: 0})
	}

	// setup the game struct
	g.Player = p

	// add my test key bindings
	addCustomKeyBindings(g)

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
			fmt.Println("getting title screen")
			s := GetTitleScreen()
			gg.CurrentScreen = &s
		}()
	})

	g.SetGlobalKeyBinding(ebiten.KeyEscape, func(gg *game.Game) {
		os.Exit(0)
	})
}

func GetDialog() dialog.Dialog {
	d := dialog.Dialog{
		BoxTilesetSource: "assets/tiled/tilesets/boxes/box1.tsj",
		TextFont: dialog.Font{
			Source: "assets/fonts/ashlander-pixel.ttf",
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
