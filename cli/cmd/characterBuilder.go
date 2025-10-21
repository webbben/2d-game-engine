/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/spf13/cobra"
	"github.com/webbben/2d-game-engine/internal/config"
	"github.com/webbben/2d-game-engine/internal/display"
	"github.com/webbben/2d-game-engine/internal/image"
	"github.com/webbben/2d-game-engine/internal/rendering"
	"github.com/webbben/2d-game-engine/internal/text"
	"github.com/webbben/2d-game-engine/internal/tiled"
	"github.com/webbben/2d-game-engine/internal/ui/button"
	"github.com/webbben/2d-game-engine/internal/ui/dropdown"
	"github.com/webbben/2d-game-engine/internal/ui/slider"
)

// characterBuilderCmd represents the characterBuilder command
var characterBuilderCmd = &cobra.Command{
	Use:   "characterBuilder",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("CHARACTER BUILDER")

		display.SetupGameDisplay("CHARACTER BUILDER", false)

		config.GameDataPathOverride = "/Users/benwebb/dev/personal/ancient-rome"
		config.DefaultFont = image.LoadFont("ashlander-pixel.ttf", 22, 0)
		config.DefaultTitleFont = image.LoadFont("ashlander-pixel.ttf", 28, 0)

		err := config.InitFileStructure()
		if err != nil {
			panic(err)
		}
		characterBuilder()
	},
}

func init() {
	rootCmd.AddCommand(characterBuilderCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// characterBuilderCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// characterBuilderCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

type builderGame struct {
	animation          string
	animationTickCount int
	ticks              int
	currentDirection   byte // L R U D
	bodyImg            *ebiten.Image
	eyesImg            *ebiten.Image
	hairImg            *ebiten.Image
	armsImg            *ebiten.Image
	nonBodyYOffset     int

	bodySet bodyPartSet
	eyesSet bodyPartSet
	hairSet bodyPartSet
	armsSet bodyPartSet

	turnLeft          *button.Button
	turnRight         *button.Button
	speedSlider       slider.Slider
	scaleSlider       slider.Slider
	animationSelector dropdown.OptionSelect
}

func characterBuilder() {
	bodyTileset := "entities/parts/body.tsj"
	eyesTileset := "entities/parts/eyes.tsj"
	hairTileset := "entities/parts/hair.tsj"

	g := builderGame{
		animation:          "run",
		animationTickCount: 15,
		currentDirection:   'D',
		bodySet: bodyPartSet{
			TilesetSrc: bodyTileset,
			DStart:     32,
			RStart:     37,
			UStart:     42,
			WalkAnimation: Animation{
				TileSteps:    []int{1, 0, 2},
				StepsOffsetY: []int{1, 0, 1},
			},
			RunAnimation: Animation{
				// TileSteps:    []int{1, 3, 0, 2, 4},
				// StepsOffsetY: []int{1, 0, 0, 1, 0},
				TileSteps:    []int{3, 1, 0, 4, 2},
				StepsOffsetY: []int{0, 1, 0, 0, 1},
			},
			HasUp:     true,
			FlipRForL: true,
		},
		armsSet: bodyPartSet{
			TilesetSrc: bodyTileset,
			DStart:     32 + 32,
			RStart:     37 + 32,
			UStart:     42 + 32,
			WalkAnimation: Animation{
				TileSteps:    []int{1, 0, 2},
				StepsOffsetY: []int{1, 0, 1},
			},
			RunAnimation: Animation{
				// TileSteps:    []int{1, 3, 0, 2, 4},
				// StepsOffsetY: []int{1, 0, 0, 1, 0},
				TileSteps:    []int{3, 1, 0, 4, 2},
				StepsOffsetY: []int{0, 1, 0, 0, 1},
			},
			HasUp:     true,
			FlipRForL: true,
		},
		eyesSet: bodyPartSet{
			TilesetSrc: eyesTileset,
			DStart:     0,
			RStart:     1,
			FlipRForL:  true,
		},
		hairSet: bodyPartSet{
			TilesetSrc: hairTileset,
			DStart:     0,
			RStart:     1,
			LStart:     2,
			UStart:     3,
			HasUp:      true,
		},
	}

	g.bodySet.Load()
	g.armsSet.Load()
	g.eyesSet.Load()
	g.hairSet.Load()

	g.setDirection('D')

	turnLeftImg := tiled.GetTileImage("ui/ui-components.tsj", 224)
	turnLeftImg = rendering.ScaleImage(turnLeftImg, config.UIScale)
	turnRightImg := tiled.GetTileImage("ui/ui-components.tsj", 225)
	turnRightImg = rendering.ScaleImage(turnRightImg, config.UIScale)
	g.turnLeft = button.NewImageButton("", nil, turnLeftImg)
	g.turnRight = button.NewImageButton("", nil, turnRightImg)

	g.speedSlider = slider.NewSlider(slider.SliderParams{
		TilesetSrc:    "ui/ui-components.tsj",
		TilesetOrigin: 256,
		TileWidth:     4,
		MinVal:        5,
		MaxVal:        20,
		InitialValue:  8,
		StepSize:      1,
	})

	g.scaleSlider = slider.NewSlider(slider.SliderParams{
		TilesetSrc:    "ui/ui-components.tsj",
		TilesetOrigin: 256,
		TileWidth:     4,
		MinVal:        3,
		MaxVal:        10,
		InitialValue:  8,
		StepSize:      1,
	})

	g.animationSelector = dropdown.NewOptionSelect(dropdown.OptionSelectParams{
		Font:                  config.DefaultFont,
		Options:               []string{"", "walk", "run"},
		InitialOptionIndex:    0,
		TilesetSrc:            "ui/ui-components.tsj",
		OriginIndex:           288,
		DropDownBoxTilesetSrc: "boxes/boxes.tsj",
		DropDownBoxOrigin:     128,
	})

	if err := ebiten.RunGame(&g); err != nil {
		panic(err)
	}
}

/*
each body part:
- has 4 directions (LRUD)
- may have movement animations (walking/running in a direction)
*/

type Animation struct {
	L            []*ebiten.Image
	R            []*ebiten.Image
	U            []*ebiten.Image
	D            []*ebiten.Image
	TileSteps    []int
	StepsOffsetY []int
}

// represents either the head, body, eyes, or hair
type bodyPartSet struct {
	animIndex                      int
	TilesetSrc                     string
	RStart, LStart, UStart, DStart int
	FlipRForL                      bool // if true, instead of using an L source, we just flip the frames for right
	WalkAnimation                  Animation
	RunAnimation                   Animation
	HasUp                          bool
}

func (set bodyPartSet) getCurrentFrame(dir byte, animationName string) *ebiten.Image {
	switch dir {
	case 'L':
		switch animationName {
		case "walk":
			return set.WalkAnimation.L[set.animIndex]
		case "run":
			return set.RunAnimation.L[set.animIndex]
		case "":
			return set.WalkAnimation.L[0]
		}
	case 'R':
		switch animationName {
		case "walk":
			return set.WalkAnimation.R[set.animIndex]
		case "run":
			return set.RunAnimation.R[set.animIndex]
		case "":
			return set.WalkAnimation.R[0]
		}
	case 'U':
		if !set.HasUp {
			return nil
		}
		switch animationName {
		case "walk":
			return set.WalkAnimation.U[set.animIndex]
		case "run":
			return set.RunAnimation.U[set.animIndex]
		case "":
			return set.WalkAnimation.U[0]
		}
	case 'D':
		switch animationName {
		case "walk":
			return set.WalkAnimation.D[set.animIndex]
		case "run":
			return set.RunAnimation.D[set.animIndex]
		case "":
			return set.WalkAnimation.D[0]
		}
	default:
		panic("invalid direction")
	}

	panic("invalid animation name")
}

func (set bodyPartSet) getCurrentYOffset(animationName string) int {
	if set.animIndex == 0 {
		return 0
	}
	switch animationName {
	case "walk":
		if len(set.WalkAnimation.StepsOffsetY) > 0 {
			return set.WalkAnimation.StepsOffsetY[set.animIndex-1]
		}
	case "run":
		if len(set.RunAnimation.StepsOffsetY) > 0 {
			return set.RunAnimation.StepsOffsetY[set.animIndex-1]
		}
	}

	return 0
}

func (set *bodyPartSet) nextFrame(animationName string) {
	set.animIndex++
	switch animationName {
	case "walk":
		if set.animIndex > len(set.WalkAnimation.TileSteps) {
			set.animIndex = 0
		}
	case "run":
		if set.animIndex > len(set.RunAnimation.TileSteps) {
			set.animIndex = 0
		}
	}
}

func (set *bodyPartSet) Load() {
	set.WalkAnimation.L = make([]*ebiten.Image, 0)
	set.WalkAnimation.R = make([]*ebiten.Image, 0)
	set.WalkAnimation.U = make([]*ebiten.Image, 0)
	set.WalkAnimation.D = make([]*ebiten.Image, 0)
	set.RunAnimation.L = make([]*ebiten.Image, 0)
	set.RunAnimation.R = make([]*ebiten.Image, 0)
	set.RunAnimation.U = make([]*ebiten.Image, 0)
	set.RunAnimation.D = make([]*ebiten.Image, 0)

	// walk animation
	if set.FlipRForL {
		set.WalkAnimation.L = getAnimationFrames(set.TilesetSrc, set.RStart, set.WalkAnimation.TileSteps, true)
	} else {
		set.WalkAnimation.L = getAnimationFrames(set.TilesetSrc, set.LStart, set.WalkAnimation.TileSteps, false)
	}
	set.WalkAnimation.R = getAnimationFrames(set.TilesetSrc, set.RStart, set.WalkAnimation.TileSteps, false)
	if set.HasUp {
		set.WalkAnimation.U = getAnimationFrames(set.TilesetSrc, set.UStart, set.WalkAnimation.TileSteps, false)
	}
	set.WalkAnimation.D = getAnimationFrames(set.TilesetSrc, set.DStart, set.WalkAnimation.TileSteps, false)

	// run animation
	if set.FlipRForL {
		set.RunAnimation.L = getAnimationFrames(set.TilesetSrc, set.RStart, set.RunAnimation.TileSteps, true)
	} else {
		set.RunAnimation.L = getAnimationFrames(set.TilesetSrc, set.LStart, set.RunAnimation.TileSteps, false)
	}
	set.RunAnimation.R = getAnimationFrames(set.TilesetSrc, set.RStart, set.RunAnimation.TileSteps, false)
	if set.HasUp {
		set.RunAnimation.U = getAnimationFrames(set.TilesetSrc, set.UStart, set.RunAnimation.TileSteps, false)
	}
	set.RunAnimation.D = getAnimationFrames(set.TilesetSrc, set.DStart, set.RunAnimation.TileSteps, false)
}

func getAnimationFrames(tilesetSrc string, startIndex int, indexSteps []int, flip bool) []*ebiten.Image {
	frames := []*ebiten.Image{}
	img := tiled.GetTileImage(tilesetSrc, startIndex)
	if flip {
		img = rendering.FlipHoriz(img)
	}
	frames = append(frames, img)
	for _, step := range indexSteps {
		img := tiled.GetTileImage(tilesetSrc, startIndex+step)
		if flip {
			img = rendering.FlipHoriz(img)
		}
		frames = append(frames, img)
	}
	return frames
}

func (bg *builderGame) Draw(screen *ebiten.Image) {
	var characterScale float64 = float64(bg.scaleSlider.GetValue())

	tileSize := int(config.TileSize * config.UIScale)

	bodyWidth := float64(bg.bodyImg.Bounds().Dx()) * characterScale
	bodyHeight := float64(bg.bodyImg.Bounds().Dy()) * characterScale

	bodyX := float64(display.SCREEN_WIDTH/2) - (bodyWidth / 2)
	bodyY := float64(display.SCREEN_HEIGHT/2) - (bodyHeight / 2)
	rendering.DrawImage(screen, bg.bodyImg, bodyX, bodyY, characterScale)
	rendering.DrawImage(screen, bg.armsImg, bodyX, bodyY, characterScale)
	eyesX := bodyX
	eyesY := bodyY + (float64(bg.nonBodyYOffset) * characterScale)
	if bg.eyesImg != nil {
		rendering.DrawImage(screen, bg.eyesImg, eyesX, eyesY, characterScale)
	}

	hairY := bodyY + (float64(bg.nonBodyYOffset) * characterScale)
	rendering.DrawImage(screen, bg.hairImg, bodyX, hairY, characterScale)

	buttonsY := bodyY + (bodyHeight) + 20
	buttonLX := (display.SCREEN_WIDTH / 2) - bg.turnLeft.Width - 20
	buttonRX := (display.SCREEN_WIDTH / 2) + 20
	bg.turnLeft.Draw(screen, buttonLX, int(buttonsY))
	bg.turnRight.Draw(screen, buttonRX, int(buttonsY))

	sliderX := 100
	sliderY := 100
	text.DrawShadowText(screen, "Ticks Per Frame", config.DefaultTitleFont, sliderX, sliderY, color.White, nil, 0, 0)
	text.DrawShadowText(screen, fmt.Sprintf("%v", bg.speedSlider.GetValue()), config.DefaultFont, sliderX-40, sliderY+(tileSize*2/3), color.White, nil, 0, 0)
	bg.speedSlider.Draw(screen, float64(sliderX), float64(sliderY))

	scaleSliderY := 200
	text.DrawShadowText(screen, "Scale", config.DefaultTitleFont, sliderX, scaleSliderY, color.White, nil, 0, 0)
	text.DrawShadowText(screen, fmt.Sprintf("%v", bg.scaleSlider.GetValue()), config.DefaultFont, sliderX-40, scaleSliderY+(tileSize*2/3), color.White, nil, 0, 0)
	bg.scaleSlider.Draw(screen, float64(sliderX), float64(scaleSliderY))

	animationSelectorY := 300
	text.DrawShadowText(screen, "Animation", config.DefaultTitleFont, sliderX, animationSelectorY, color.White, nil, 0, 0)
	bg.animationSelector.Draw(screen, float64(sliderX), float64(animationSelectorY), nil)
}

func (bg *builderGame) Update() error {
	if bg.turnLeft.Update().Clicked {
		bg.rotateLeft()
	} else if bg.turnRight.Update().Clicked {
		bg.rotateRight()
	}
	bg.speedSlider.Update()
	bg.scaleSlider.Update()

	bg.animationSelector.Update()
	selectorValue := bg.animationSelector.GetCurrentValue()
	if selectorValue != bg.animation {
		bg.setAnimation(selectorValue)
	}

	bg.animationTickCount = bg.speedSlider.GetValue()

	if bg.animation != "" {
		bg.ticks++
		if bg.ticks > bg.animationTickCount {
			bg.ticks = 0
			bg.bodySet.nextFrame(bg.animation)
			bg.armsSet.nextFrame(bg.animation)
		}
	}

	bg.bodyImg = bg.bodySet.getCurrentFrame(bg.currentDirection, bg.animation)
	bg.eyesImg = bg.eyesSet.getCurrentFrame(bg.currentDirection, bg.animation)
	bg.hairImg = bg.hairSet.getCurrentFrame(bg.currentDirection, bg.animation)
	bg.armsImg = bg.armsSet.getCurrentFrame(bg.currentDirection, bg.animation)

	bg.nonBodyYOffset = bg.bodySet.getCurrentYOffset(bg.animation)

	return nil
}

func (bg *builderGame) setAnimation(animation string) {
	bg.animation = animation
	bg.bodySet.animIndex = 0
	bg.eyesSet.animIndex = 0
	bg.hairSet.animIndex = 0
	bg.armsSet.animIndex = 0
}

func (bg *builderGame) rotateLeft() {
	switch bg.currentDirection {
	case 'L':
		bg.setDirection('U')
	case 'U':
		bg.setDirection('R')
	case 'R':
		bg.setDirection('D')
	case 'D':
		bg.setDirection('L')
	}
}

func (bg *builderGame) rotateRight() {
	switch bg.currentDirection {
	case 'L':
		bg.setDirection('D')
	case 'D':
		bg.setDirection('R')
	case 'R':
		bg.setDirection('U')
	case 'U':
		bg.setDirection('L')
	}
}

func (bg *builderGame) setDirection(dir byte) {
	bg.bodySet.animIndex = 0
	bg.eyesSet.animIndex = 0
	bg.hairSet.animIndex = 0
	bg.armsSet.animIndex = 0

	switch dir {
	case 'U':
		bg.currentDirection = 'U'
		bg.bodyImg = bg.bodySet.WalkAnimation.U[0]
		bg.eyesImg = nil
		bg.hairImg = bg.hairSet.WalkAnimation.U[0]
		bg.armsImg = bg.armsSet.WalkAnimation.U[0]
	case 'R':
		bg.currentDirection = 'R'
		bg.bodyImg = bg.bodySet.WalkAnimation.R[0]
		bg.eyesImg = bg.eyesSet.WalkAnimation.R[0]
		bg.hairImg = bg.hairSet.WalkAnimation.R[0]
		bg.armsImg = bg.armsSet.WalkAnimation.R[0]
	case 'D':
		bg.currentDirection = 'D'
		bg.bodyImg = bg.bodySet.WalkAnimation.D[0]
		bg.eyesImg = bg.eyesSet.WalkAnimation.D[0]
		bg.hairImg = bg.hairSet.WalkAnimation.D[0]
		bg.armsImg = bg.armsSet.WalkAnimation.D[0]
	case 'L':
		bg.currentDirection = 'L'
		bg.bodyImg = bg.bodySet.WalkAnimation.L[0]
		bg.eyesImg = bg.eyesSet.WalkAnimation.L[0]
		bg.hairImg = bg.hairSet.WalkAnimation.L[0]
		bg.armsImg = bg.armsSet.WalkAnimation.L[0]
	default:
		panic("direction not recognized")
	}
}

func (bg *builderGame) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return display.SCREEN_WIDTH, display.SCREEN_HEIGHT
}
