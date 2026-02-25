package textwindow

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/webbben/2d-game-engine/config"
	"github.com/webbben/2d-game-engine/ui/overlay"
	"github.com/webbben/2d-game-engine/imgutil/rendering"
	"github.com/webbben/2d-game-engine/ui/text"
	"golang.org/x/image/font"
)

type CustomHoverWindow struct {
	placeHolderImage *ebiten.Image
	hoverBox         textWindowBox
	title            string
	f                font.Face

	customBodyContent *ebiten.Image
	boxTilesetSrc     string
	originTileIndex   int
}

// NewCustomHoverWindow creates a new custom hover window UI component. After creating, make sure to set the body content image with SetCustomBodyContent.
func NewCustomHoverWindow(title string, titleFont font.Face, boxTilesetSrc string, originTileIndex int) CustomHoverWindow {
	if title == "" {
		panic("title must not be empty")
	}

	hw := CustomHoverWindow{
		title:           title,
		f:               titleFont,
		boxTilesetSrc:   boxTilesetSrc,
		originTileIndex: originTileIndex,
	}

	return hw
}

func (hw *CustomHoverWindow) SetCustomBodyContent(bodyContent *ebiten.Image) {
	bodyBounds := bodyContent.Bounds()
	bodyDx, bodyDy := bodyBounds.Dx(), bodyBounds.Dy()

	hw.hoverBox = newTextWindowBox(hw.boxTilesetSrc, hw.originTileIndex)

	tileSize := int(config.TileSize * config.UIScale)

	// adjust to fit the entire window (with its title section, borders, etc)
	bodyDx += tileSize
	bodyDx -= bodyDx % tileSize
	bodyDy += tileSize * 2
	bodyDy -= bodyDy % tileSize

	hw.hoverBox.buildWindowImage(bodyDx, bodyDy)

	// this is the slate everything is drawn on each Draw
	hw.placeHolderImage = ebiten.NewImage(bodyDx, bodyDy)

	hw.customBodyContent = bodyContent
}

func (hw CustomHoverWindow) Draw(om *overlay.OverlayManager) {
	if hw.placeHolderImage == nil {
		panic("placeholderimage is nil")
	}
	if hw.hoverBox.windowImage == nil {
		panic("window image is nil")
	}
	if hw.customBodyContent == nil {
		panic("customBodyContent is nil. ensure it is initialized with SetCustomBodyContent")
	}
	if hw.title == "" {
		panic("title is nil")
	}

	hw.placeHolderImage.Clear()

	rendering.DrawImage(hw.placeHolderImage, hw.hoverBox.windowImage, 0, 0, 0)

	tileSize := config.GetScaledTilesize()
	// int(tw.x)+(tileSize/2), int(tw.y)+(tileSize)-5,
	tx := int(tileSize / 2)
	ty := int(tileSize - 5)

	text.DrawShadowText(hw.placeHolderImage, hw.title, hw.f, tx, ty, nil, nil, 0, 0)

	rendering.DrawImage(hw.placeHolderImage, hw.customBodyContent, float64(tx), tileSize*1.5, 0)

	bounds := hw.placeHolderImage.Bounds()
	drawX, drawY := getPosNearMouse(15, bounds.Dx(), bounds.Dy())

	om.AddOverlay(hw.placeHolderImage, float64(drawX), float64(drawY))
}
