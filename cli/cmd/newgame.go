package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/webbben/2d-game-engine/game"
	"github.com/webbben/2d-game-engine/internal/display"
)

/*
*
*	For now, this is the command that simulates starting a new game.
*	Until I organize things better and/or move game dev code to the game's actual repo, this command
*	effectively represents "the game" that I'm creating.
*	The other command (testrun) is simply for testing.
*
 */

// newgameCmd represents the newgame command
var newgameCmd = &cobra.Command{
	Use:   "newgame",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("newgame called")
		display.SetupGameDisplay("Ancient Rome!", false)

		SetConfig()

		err := game.InitialStartUp()
		if err != nil {
			panic(err)
		}

		g := setupGameState(gameParams{
			startHour:  23,
			startMapID: "prison_ship",
		})

		if err := g.RunGame(); err != nil {
			panic(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(newgameCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// newgameCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// newgameCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
