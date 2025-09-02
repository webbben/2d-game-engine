/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/webbben/2d-game-engine/entity"
	"github.com/webbben/2d-game-engine/internal/config"
	"github.com/webbben/2d-game-engine/internal/general_util"
)

// newEntityCmd represents the newEntity command
var newEntityCmd = &cobra.Command{
	Use:   "newEntity",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		e := entity.Entity{}
		e.ID = general_util.GenerateUUID()
		e.DisplayName = general_util.PromptUserInput("Entity Display Name")
		fmt.Println("Enter the full paths to the tilesets used by this entity. Enter 'q' when done.")
		for {
			input := general_util.PromptUserInput("Tileset path (q to quit)")
			if input == "q" {
				break
			}
			e.FrameTilesetSources = append(e.FrameTilesetSources, input)
		}

		err := e.SaveJSON()
		if err != nil {
			fmt.Println("error saving entity JSON:", err)
			return
		}
		fmt.Printf("entity JSON saved at %s\n", filepath.Join(config.GameDataRootPath(), "ent", fmt.Sprintf("ent_%s.json", e.ID)))
	},
}

func init() {
	rootCmd.AddCommand(newEntityCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// newEntityCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// newEntityCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
