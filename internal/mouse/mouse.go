package mouse

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/webbben/2d-game-engine/internal/config"
)

type MouseBehavior struct {
	IsHovering bool // mouse is hovering
	LeftClick  ClickBehavior
	RightClick ClickBehavior
}

type ClickBehavior struct {
	ClickStart    bool
	ClickHolding  bool
	ClickReleased bool
}

func (mouseBehavior *MouseBehavior) Update(drawX, drawY int, boxWidth, boxHeight int) {
	mouseX, mouseY := ebiten.CursorPosition()

	// adjust to game scale
	boxWidth *= int(config.GameScale)
	boxHeight *= int(config.GameScale)

	// detect hovering
	if mouseX > int(drawX) && mouseX < (drawX+boxWidth) {
		if mouseY > int(drawY) && mouseY < (drawY+boxHeight) {
			mouseBehavior.IsHovering = true

			// detect click behavior
			mouseBehavior.LeftClick.detectClick(ebiten.MouseButtonLeft, mouseBehavior.LeftClick)
			mouseBehavior.RightClick.detectClick(ebiten.MouseButtonRight, mouseBehavior.RightClick)
		}
	}
}

func (c *ClickBehavior) detectClick(mouseButton ebiten.MouseButton, prev ClickBehavior) {
	if ebiten.IsMouseButtonPressed(mouseButton) {
		if prev.ClickStart {
			// was started last tick; now it should be "held"
			c.ClickStart = false
			c.ClickHolding = true
		} else if prev.ClickHolding {
			// still holding; no change needed
		} else if prev.ClickReleased {
			// woah - the user clicked so rapidly that one tick after releasing, a new click was detected
			c.ClickStart = true
			c.ClickReleased = false
		} else {
			// a new click was found
			c.ClickStart = true
		}
	} else {
		// click is not active
		// undo all active clicking state
		c.ClickStart = false
		c.ClickHolding = false

		// check if it is being released
		if prev.ClickStart || prev.ClickHolding {
			c.ClickReleased = true
		} else {
			// last tick was not a mouse click; left click is officially completely done
			c.ClickReleased = false
		}
	}
}
