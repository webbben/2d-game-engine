package screen

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text"
	"github.com/webbben/2d-game-engine/image"
	"github.com/webbben/2d-game-engine/rendering"
	"golang.org/x/image/font"
)

// Screen represents a screen in the game
//
// A screen is a full-screen display that is not part of the game world, and is used for things like:
//
// * title screens
//
// * options screens
//
// * load save screens
//
// * etc.
//
// If a screen is set in the game state, the game world is not active or being displayed.
type Screen struct {
	ID         string
	screenInit bool // if the screen has been initialized and is ready for rendering

	Menus               []Menu
	Background          *ebiten.Image
	BackgroundImagePath string
	Title               string
	TitleFontName       string
	titleFont           font.Face
	TitleFontColor      color.Color
	BodyFontName        string
	bodyFont            font.Face
	BodyFontColor       color.Color
}

// Menu represents a group of buttons that represent options for the user
//
// These options are likely for navigating to a screen, starting a game, loading a game, etc.
//
// Buttons in a menu use the same borders/images and are the same size
type Menu struct {
	Buttons        []Button
	BoxImage       *ebiten.Image
	BoxTilesetPath string
}

type Button struct {
	Text     string
	Callback func()
}

func (s *Screen) init() {
	img, err := image.LoadImage(s.BackgroundImagePath)
	if err != nil {
		panic(err)
	}
	s.Background = img

	// create shadow backdrop
	shadow := image.NewRadialGradientImage(s.Background.Bounds().Dy())
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(s.Background.Bounds().Dx()-shadow.Bounds().Dx())/2, 0)
	s.Background.DrawImage(shadow, op)

	// load fonts
	s.titleFont = image.LoadFont(s.TitleFontName)
	s.bodyFont = image.LoadFont(s.BodyFontName)

	// create button images
	for i := range s.Menus {
		// first, find out the max width of the buttons
		width := 80
		height := 30
		for _, button := range s.Menus[i].Buttons {
			buttonWidth := rendering.TextWidth(button.Text, s.bodyFont) + 50
			if buttonWidth > width {
				width = buttonWidth
			}
		}
		boxTileset, err := image.LoadBoxTileSet(s.Menus[i].BoxTilesetPath)
		if err != nil {
			panic(err)
		}
		tilesize := boxTileset.Top.Bounds().Dx()
		s.Menus[i].BoxImage = image.CreateBox(width/tilesize, height/tilesize, boxTileset, 1, 1)
	}

	s.screenInit = true
}

func (s *Screen) DrawScreen(screen *ebiten.Image) {
	if !s.screenInit {
		return
	}
	screen.DrawImage(s.Background, nil)

	// draw the title
	titleX, _ := rendering.CenterTextOnImage(screen, s.Title, s.titleFont)
	text.Draw(screen, s.Title, s.titleFont, titleX, 20, s.TitleFontColor)

	// draw the first menu only for now
	menu := s.Menus[0]
	for i, button := range menu.Buttons {
		buttonX, _ := rendering.CenterImageOnImage(screen, menu.BoxImage)
		buttonY := (menu.BoxImage.Bounds().Dy()+30)*(i+1) + 150
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(float64(buttonX), float64(buttonY))
		screen.DrawImage(menu.BoxImage, op)
		offX, offY := rendering.CenterTextOnImage(menu.BoxImage, button.Text, s.bodyFont)
		text.Draw(screen, button.Text, s.bodyFont, buttonX+offX, buttonY+offY, s.BodyFontColor)
	}
}

func (s *Screen) UpdateScreen() {
	if !s.screenInit {
		s.init()
	}
}
