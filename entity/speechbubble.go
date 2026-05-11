package entity

import (
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/webbben/2d-game-engine/config"
	"github.com/webbben/2d-game-engine/imgutil/rendering"
	"github.com/webbben/2d-game-engine/logz"
	"github.com/webbben/2d-game-engine/model"
	"github.com/webbben/2d-game-engine/ui/box"
	"github.com/webbben/2d-game-engine/ui/overlay"
	"github.com/webbben/2d-game-engine/ui/text"
	"github.com/webbben/2d-game-engine/utils"
	"golang.org/x/image/font"
)

type SpeechBubble struct {
	speechBubbleFont     font.Face
	speechBubbleImg      *ebiten.Image
	speechBubbleText     string
	speechBubbleDuration time.Duration
	speechBubbleCreation time.Time
	showSpeechBubble     bool
}

func (sb SpeechBubble) Done() bool {
	if sb.speechBubbleCreation.IsZero() {
		return true
	}
	return time.Now().After(sb.speechBubbleCreation.Add(sb.speechBubbleDuration))
}

type SpeechBubbleParams struct {
	Font          font.Face
	Duration      time.Duration
	BoxTileset    string
	BoxTileOrigin int
}

func NewSpeechBubble(s string, params SpeechBubbleParams) *SpeechBubble {
	if params.BoxTileset == "" {
		logz.Panic("Box tileset was empty")
	}
	if params.Font == nil {
		params.Font = config.DefaultFont
	}
	sb := &SpeechBubble{
		speechBubbleFont:     params.Font,
		speechBubbleDuration: params.Duration,
		speechBubbleCreation: time.Now(),
	}

	b := box.NewBox(params.BoxTileset, params.BoxTileOrigin)

	dx, _, _ := text.GetStringSize(s, params.Font)
	tilesize := int(config.TileSize * config.GameScale)
	// round up to minimum tiles to contain the string
	width := utils.RoundUpToTile(dx, tilesize)
	// speech bubbles have slim border tiles, so add extra width so text doesn't spill out
	width += tilesize * 2
	// ensure minimum width
	width = max(tilesize*3, width)
	height := tilesize * 3
	boxImg := b.BuildBoxImage(width, height, config.GameScale)
	sb.speechBubbleText = s
	sb.showSpeechBubble = true

	sb.speechBubbleImg = ebiten.NewImage(width, height)
	rendering.DrawImage(sb.speechBubbleImg, boxImg, 0, 0, 0)

	// center text in bubble
	x, y := text.CenterTextInRect(s, params.Font, model.NewRect(0, 0, float64(width), float64(height)))
	text.DrawShadowText(sb.speechBubbleImg, sb.speechBubbleText, sb.speechBubbleFont, x, y, nil, nil, 0, 0)

	return sb
}

func (sb *SpeechBubble) Draw(screen *ebiten.Image, x, y float64) {
	rendering.DrawImage(screen, sb.speechBubbleImg, x, y, 0)
}

func (sb *SpeechBubble) DrawOverlay(om *overlay.OverlayManager, x, y float64) {
	om.AddOverlay(sb.speechBubbleImg, x, y)
}
