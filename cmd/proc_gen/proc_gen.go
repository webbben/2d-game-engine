package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/webbben/2d-game-engine/proc_gen"
)

func main() {
	width := 100
	height := 100

	genNoiseMapCmd := flag.NewFlagSet("genNoise", flag.ExitOnError)
	noiseTypeArg := genNoiseMapCmd.String("type", "", "type of noise map to generate")

	if len(os.Args) < 2 {
		fmt.Println("Error: no command provided")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "genNoise":
		genNoiseMapCmd.Parse(os.Args[2:])
		if *noiseTypeArg == "" {
			fmt.Println("Error: noiseType argument is required for genNoise command")
			genNoiseMapCmd.Usage()
			os.Exit(1)
		}
		fmt.Println("Generating noise map of type:", *noiseTypeArg)
		switch *noiseTypeArg {
		case "forest":
			proc_gen.GenerateForest(width, height)
		case "town_elev":
			proc_gen.GenerateTownElevation(width, height)
		case "mountain":
			proc_gen.GenerateMountain(width, height)
		}
	default:
		fmt.Println("command not recognized.")
		genNoiseMapCmd.Usage()
	}
}
