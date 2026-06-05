package dialogv2

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/webbben/2d-game-engine/audio"
	"github.com/webbben/2d-game-engine/config"
	"github.com/webbben/2d-game-engine/data/defs"
	"github.com/webbben/2d-game-engine/display"
	"github.com/webbben/2d-game-engine/imgutil/rendering"
	"github.com/webbben/2d-game-engine/logz"
	"github.com/webbben/2d-game-engine/ui/button"
	"github.com/webbben/2d-game-engine/ui/text"
	"github.com/webbben/2d-game-engine/utils"
)

// the box that holds reply options; used when replies are too big to fit in the topic box
type replyBox struct {
	replyBoxImage *ebiten.Image
	replyGap      float64
	// replyButtons  []*button.MultilineButton
	replyButtons []*ReplyComponent
}

func (ds *DialogSession) setupReplyBox(maxReplyWidth int, replies []defs.DialogReply) {
	if len(ds.replyList) == 0 {
		panic("replylist was empty")
	}
	ds.replyBox = &replyBox{}

	tileSize := config.GetScaledTilesize()

	// the max width that any reply button should ever have
	maxWidth := display.SCREEN_WIDTH * 2 / 3
	maxWidth -= maxWidth % int(tileSize)

	// max width of a reply button.
	maxWidth = min(maxReplyWidth, maxWidth)

	totalHeight := 0

	for _, reply := range replies {
		replyComponent := NewReplyComponent(reply, maxWidth, ds.audioman, ds.Ctx)
		ds.replyBox.replyButtons = append(ds.replyBox.replyButtons, &replyComponent)
		_, dy := replyComponent.Dimensions()
		totalHeight += dy
	}

	boxWidth := maxWidth + int(tileSize)
	totalHeight += int(tileSize * 2)
	boxWidth = utils.RoundUpToTile(boxWidth, int(tileSize))
	totalHeight = utils.RoundUpToTile(totalHeight, int(tileSize))

	// create the big replies box
	// we add the extra padding here since we want the reply box to be wider, but not the buttons themselves
	ds.replyBox.replyBoxImage = ds.boxSrc.BuildBoxImage(boxWidth, totalHeight, config.UIScale)
}

func (ds *DialogSession) updateReplyBox() {
	if ds.replyBox == nil {
		panic("replybox was nil")
	}
	if len(ds.replyBox.replyButtons) == 0 {
		panic("there are no replybox buttons")
	}
	for i, b := range ds.replyBox.replyButtons {
		if b.Update() {
			r := ds.replyList[i]
			ds.ApplyReply(r)
			return
		}
	}
}

func (rb *replyBox) draw(screen *ebiten.Image, textBoxY int) {
	tileSize := config.GetScaledTilesize()
	replyBoxDx, replyBoxDy := rb.replyBoxImage.Bounds().Dx(), rb.replyBoxImage.Bounds().Dy()
	bigReplyBoxX := (display.SCREEN_WIDTH / 2) - (replyBoxDx / 2)
	bigReplyBoxY := (textBoxY / 2) - (replyBoxDy / 2)
	rendering.DrawImage(screen, rb.replyBoxImage, float64(bigReplyBoxX), float64(bigReplyBoxY), 0)
	btnDx, _ := rb.replyButtons[0].Dimensions()
	replyX := float64(bigReplyBoxX + (replyBoxDx / 2) - (btnDx / 2))
	replyY := float64(bigReplyBoxY) + (tileSize / 2)

	for _, b := range rb.replyButtons {
		b.Draw(screen, replyX, replyY)
		_, dy := b.Dimensions()
		replyY += float64(dy) + rb.replyGap
	}
}

type ReplyComponent struct {
	Text       string // text of the reply
	Decoration defs.ReplyDecoration
	InfoText   string

	ml  text.Multiline
	btn *button.Button
}

func (rc ReplyComponent) Dimensions() (dx, dy int) {
	return rc.btn.Width, rc.btn.Height
}

func NewReplyComponent(reply defs.DialogReply, maxWidth int, audioman *audio.AudioManager, ctx defs.ConditionContext) ReplyComponent {
	rc := ReplyComponent{
		Text:       reply.Text,
		Decoration: reply.Decoration,
	}

	if reply.InfoText != nil {
		rc.InfoText = reply.InfoText.GetInfoText(ctx)
	}

	padding := 20 // padding on each side of inner content

	totalDy := padding * 2
	rc.ml = text.NewMultiline(reply.Text, maxWidth-(padding*2), config.DefaultFont, text.MultilineParams{})
	_, mlDy := rc.ml.Dimensions()
	logz.Println("mlDy", mlDy)
	totalDy += mlDy

	_, dsc := text.GetRealisticFontMetrics(config.DefaultFont)
	totalDy += dsc

	if rc.InfoText != "" {
		dy, _ := text.GetRealisticFontMetrics(config.DefaultInfoFont)
		logz.Println("infoDy", dy)
		totalDy += dy
		totalDy += dsc
	}

	logz.Println("totalDy:", totalDy)

	rc.btn = button.NewButton("", nil, maxWidth, totalDy, audioman)

	return rc
}

func (rc *ReplyComponent) Draw(screen *ebiten.Image, x, y float64) {
	// position of the decoration/border
	decoWidth := rc.btn.Width - 20
	decoHeight := rc.btn.Height - 20
	decoX := x + 10
	decoY := y + 10

	drawX := decoX + 10
	drawY := decoY + 10

	rc.ml.Draw(screen, drawX, drawY)

	if rc.InfoText != "" {
		// draw below the multiline
		infoTextDx, _, _ := text.GetStringSize(rc.InfoText, config.DefaultInfoFont)
		drawX += float64(decoWidth) - 14 - float64(infoTextDx) // subtract 10 and a little extra to keep text from running into deco

		_, dsc := text.GetRealisticFontMetrics(config.DefaultInfoFont)
		// text draws from bottom, but ml draws from top
		// since drawX is decoX + 10, we need to subtract 10 from decoHeight. also subtract descent
		drawY += float64(decoHeight) - 10 - float64(dsc)
		infoColor := color.RGBA{0, 0, 0, 120}
		text.DrawText(screen, rc.InfoText, config.DefaultInfoFont, int(drawX), int(drawY), infoColor)
	}

	switch rc.Decoration {
	case defs.ReplyDecoGood:
		c := color.RGBA{0, 255, 0, 0}
		vector.StrokeRect(screen, float32(decoX), float32(decoY), float32(decoWidth), float32(decoHeight), 1, c, false)
	case defs.ReplyDecoBad:
		c := color.RGBA{255, 0, 0, 0}
		vector.StrokeRect(screen, float32(decoX), float32(decoY), float32(decoWidth), float32(decoHeight), 1, c, false)
	}

	rc.btn.Draw(screen, int(x), int(y))
}

func (rc *ReplyComponent) Update() bool {
	return rc.btn.Update().Clicked
}
