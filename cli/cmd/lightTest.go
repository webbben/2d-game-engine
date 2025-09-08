/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/webbben/2d-game-engine/internal/display"
)

// lightTestCmd represents the lightTest command
var lightTestCmd = &cobra.Command{
	Use:   "lightTest",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("lightTest called")
	},
}

func init() {
	rootCmd.AddCommand(lightTestCmd)

	display.SetupGameDisplay("Ancient Rome!", false)

}
