package screen

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text"
	"github.com/webbben/2d-game-engine/internal/display"
	"github.com/webbben/2d-game-engine/internal/general_util"
	"github.com/webbben/2d-game-engine/internal/image"
	"github.com/webbben/2d-game-engine/internal/model"
	"github.com/webbben/2d-game-engine/internal/rendering"
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
	Text      string
	Callback  func()
	isHovered bool
	pos       model.Coords
	bounds    model.Coords
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
	s.titleFont = image.LoadFont(s.TitleFontName, 48, 72)
	s.bodyFont = image.LoadFont(s.BodyFontName, 0, 0)

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
		imageWidth := s.Menus[i].BoxImage.Bounds().Dx()
		imageHeight := s.Menus[i].BoxImage.Bounds().Dy()

		// set button positions
		for j := range s.Menus[i].Buttons {
			buttonX := (display.SCREEN_WIDTH - s.Menus[i].BoxImage.Bounds().Dx()) / 2
			buttonY := (s.Menus[i].BoxImage.Bounds().Dy()+30)*(j+1) + 150
			s.Menus[i].Buttons[j].pos = model.Coords{
				X: buttonX,
				Y: buttonY,
			}
			s.Menus[i].Buttons[j].bounds = model.Coords{X: imageWidth, Y: imageHeight}
		}
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
	text.Draw(screen, s.Title, s.titleFont, titleX, 80, s.TitleFontColor)

	// draw the first menu only for now
	menu := s.Menus[0]
	for _, button := range menu.Buttons {
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(float64(button.pos.X), float64(button.pos.Y))
		if button.isHovered {
			op.ColorScale.Scale(1.2, 1.2, 1.2, 1)
		}
		screen.DrawImage(menu.BoxImage, op)
		offX, offY := rendering.CenterTextOnImage(menu.BoxImage, button.Text, s.bodyFont)
		text.Draw(screen, button.Text, s.bodyFont, button.pos.X+offX, button.pos.Y+offY, s.BodyFontColor)
	}
}

func (s *Screen) UpdateScreen() {
	if !s.screenInit {
		s.init()
	}

	// update buttons
	for i := range s.Menus {
		for j := range s.Menus[i].Buttons {
			s.Menus[i].Buttons[j].UpdateButton()
		}
	}
}

func (b *Button) UpdateButton() {
	// check for mouse hover and click
	hover, click := general_util.DetectMouse(b.pos.X, b.pos.Y, b.pos.X+b.bounds.X, b.pos.Y+b.bounds.Y)
	if hover {
		b.isHovered = true
		if click {
			b.Callback()
		}
	} else {
		b.isHovered = false
	}
}
