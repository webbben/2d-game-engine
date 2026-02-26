// Package mouse
package mouse

import (
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/webbben/2d-game-engine/config"
)

type MouseBehavior struct {
	IsHovering       bool // mouse is hovering
	HoverStart       time.Time
	LeftClick        ClickBehavior
	LeftClickOutside ClickBehavior // for detecting if outside clicks occur
	RightClick       ClickBehavior
}

type ClickBehavior struct {
	LastClick     time.Time // time the last click was released
	lastClickMs   int64     // ms since last-last click was done
	ClickStart    bool
	ClickHolding  bool
	ClickReleased bool
}

func (cb ClickBehavior) String() string {
	if cb.ClickStart {
		return "Click Start"
	}
	if cb.ClickHolding {
		return "Click Holding"
	}
	if cb.ClickReleased {
		return "Click Released"
	}
	return "no behavior detected"
}

func (cb *ClickBehavior) Reset() {
	cb.ClickStart = false
	cb.ClickHolding = false
	cb.ClickReleased = false
}

func (cb ClickBehavior) DoubleClicked() bool {
	return cb.ClickReleased && (time.Duration(cb.lastClickMs) < 250)
}

func (mouseBehavior *MouseBehavior) Update(drawX, drawY int, boxWidth, boxHeight int, scaleForGameWorld bool) {
	mouseX, mouseY := ebiten.CursorPosition()

	// adjust to game scale if in world
	if scaleForGameWorld {
		boxWidth = int(float64(boxWidth) * config.GameScale)
		boxHeight = int(float64(boxHeight) * config.GameScale)
		drawX = int(float64(drawX) * config.GameScale)
		drawY = int(float64(drawY) * config.GameScale)
	}

	// detect hovering
	mouseBehavior.IsHovering = false
	if mouseX > int(drawX) && mouseX < (drawX+boxWidth) {
		if mouseY > int(drawY) && mouseY < (drawY+boxHeight) {
			mouseBehavior.IsHovering = true
			mouseBehavior.LeftClickOutside.Reset() // not clicking outside, so reset this

			// detect click behavior
			mouseBehavior.LeftClick.detectClick(ebiten.MouseButtonLeft, mouseBehavior.LeftClick)
			mouseBehavior.RightClick.detectClick(ebiten.MouseButtonRight, mouseBehavior.RightClick)
		}
	}
	// if not hovering, unset any active click states
	if !mouseBehavior.IsHovering {
		mouseBehavior.HoverStart = time.Now() // continuously set the start time to now, until actually hovering
		mouseBehavior.LeftClick.Reset()
		mouseBehavior.RightClick.Reset()

		// detect outside clicks
		mouseBehavior.LeftClickOutside.detectClick(ebiten.MouseButtonLeft, mouseBehavior.LeftClickOutside)
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
			c.lastClickMs = time.Since(c.LastClick).Milliseconds()
			c.LastClick = time.Now()
			c.ClickReleased = true
		} else {
			// last tick was not a mouse click; left click is officially completely done
			c.ClickReleased = false
		}
	}
}
