package dialogv2

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/webbben/2d-game-engine/config"
	"github.com/webbben/2d-game-engine/data/defs"
	"github.com/webbben/2d-game-engine/display"
	"github.com/webbben/2d-game-engine/imgutil/rendering"
	"github.com/webbben/2d-game-engine/ui/button"
	"github.com/webbben/2d-game-engine/ui/text"
)

type replyBox struct {
	replyBoxImage *ebiten.Image
	replyGap      float64
	replyButtons  []*button.MultilineButton
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
		btn := button.NewMultilineButton(reply.Text, maxWidth, config.DefaultFont, text.MultilineParams{})
		ds.replyBox.replyButtons = append(ds.replyBox.replyButtons, btn)
		_, dy := btn.Dimensions()
		totalHeight += dy
	}

	boxWidth := maxWidth + int(tileSize)
	totalHeight += int(tileSize * 2)
	boxWidth -= boxWidth % int(tileSize)
	totalHeight -= totalHeight % int(tileSize)

	// create the big replies box
	// we add the extra padding here since we want the reply box to be wider, but not the buttons themselves
	ds.replyBox.replyBoxImage = ds.boxSrc.BuildBoxImage(boxWidth, totalHeight)
}

func (ds *DialogSession) updateReplyBox() {
	if ds.replyBox == nil {
		panic("replybox was nil")
	}
	if len(ds.replyBox.replyButtons) == 0 {
		panic("there are no replybox buttons")
	}
	for i, b := range ds.replyBox.replyButtons {
		if b.Update().Clicked {
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
	replyX := float64(bigReplyBoxX) + (tileSize / 2)
	replyY := float64(bigReplyBoxY) + (tileSize / 2)

	for _, b := range rb.replyButtons {
		b.Draw(screen, replyX, replyY)
		_, dy := b.Dimensions()
		replyY += float64(dy) + rb.replyGap
	}
}
