package main

import (
	"fmt"
	"image"
	"image/png"
	"log"
	"os"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/webbben/2d-game-engine/tileset"
)

type Game struct {
	sprite   *ebiten.Image
	exported bool
}

func (g *Game) Update() error {
	if g.sprite != nil && !g.exported {
		err := SaveImageToFile(g.sprite, "_testing/build_sprite.png")
		if err != nil {
			log.Fatal("error saving png:", err)
		}
		log.Println("exported")
		g.exported = true
	}
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	if g.sprite != nil {
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(0, 0)
		op.GeoM.Scale(10, 10)
		screen.DrawImage(g.sprite, op)
	} else {
		log.Println("no sprite...")
	}
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return 500, 500
}

func main() {
	baseDir := "/Users/benwebb/Desktop/game_art/character"
	spriteComponents := tileset.SpriteComponentPaths{
		Skin: tileset.SpriteComponent{
			ImagePath: fmt.Sprintf("%s/skin/skin_01_down.png", baseDir),
		},
		Head: tileset.SpriteComponent{
			ImagePath: fmt.Sprintf("%s/head/head_01_down.png", baseDir),
			Dy:        2,
		},
		Body: tileset.SpriteComponent{
			ImagePath: fmt.Sprintf("%s/body/body_01_down.png", baseDir),
			Dy:        -1,
		},
		Legs: tileset.SpriteComponent{
			ImagePath: fmt.Sprintf("%s/legs/legs_01_down.png", baseDir),
		},
		Shadow: tileset.SpriteComponent{
			ImagePath: fmt.Sprintf("%s/shadow/shadow_01_down.png", baseDir),
		},
	}
	img, err := tileset.BuildSpriteFrameImage(spriteComponents)
	if err != nil {
		fmt.Println("error building sprite:", err)
		return
	}

	// err = SaveImageToFile(img, "build_sprite.png")
	// if err != nil {
	// 	fmt.Println("error saving png:", err)
	// 	return
	// }

	g := Game{sprite: img}

	ebiten.SetWindowSize(500, 500)
	ebiten.SetWindowTitle("sprite test build")
	if err := ebiten.RunGame(&g); err != nil {
		log.Fatal(err)
	}
}

// SaveImageToFile saves an ebiten.Image to a PNG file.
func SaveImageToFile(img *ebiten.Image, filename string) error {
	// Get the size of the image
	width, height := img.Bounds().Dx(), img.Bounds().Dy()

	// Create a new RGBA image
	rgba := image.NewRGBA(image.Rect(0, 0, width, height))

	// Copy pixel data from the ebiten.Image
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			rgba.Set(x, y, img.At(x, y))
		}
	}

	// Create the file
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	// Encode the image as PNG and save it
	return png.Encode(file, rgba)
}
