// Package mouse
package mouse

import (
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/webbben/2d-game-engine/config"
	"github.com/webbben/2d-game-engine/imgutil/rendering"
	"github.com/webbben/2d-game-engine/logz"
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

type CursorShape int

const (
	DefaultShape CursorShape = iota
	TextShape
	PointerShape
	CrosshairShape

	// we don't support resize cursor shapes for now, since I don't think we'll need them in-game.
)

var (
	CurrentShape CursorShape = DefaultShape
	NextShape    CursorShape = DefaultShape

	// if these are set, they will be drawn (and system cursor will be hidden)

	DefaultShapeCustomImg   *ebiten.Image
	TextShapeCustomImg      *ebiten.Image
	PointerShapeCustomImg   *ebiten.Image
	CrosshairShapeCustomImg *ebiten.Image
)

// SetCursorShape basically signals what next cursor shape should be applied. Then, the main update loop actually calls
// ChangeCursorShape, which handles really changing things. It's good to use this function during update logic just so we aren't
// spamming ebiten.SetCursorMode/SetCursorShape as much.
func SetCursorShape(shape CursorShape) {
	NextShape = shape
}

// UpdateCursorShape is used in the actual game update loop to apply a cursor change.
// This is called at the end of the update loop, so that we aren't constantly switching from default to another shape on every loop.
// Update logic can set a "next shape", and then at the end of the entire update loop, we will check if the "next shape" is different from the current shape,
// and if so, actually apply a change.
//
// I don't think constantly changing the cursor shape on each update loop has any performance impact, but I just don't like the idea of spamming functions
// that make actual changes unnecessarily.
func UpdateCursorShape() {
	if NextShape == CurrentShape {
		return
	}

	logz.Println("UpdateCursorShape", NextShape)

	CurrentShape = NextShape
	ebiten.SetCursorMode(ebiten.CursorModeVisible)

	// check if the shape has an image override, and if so, hide the default system cursor
	switch NextShape {
	case DefaultShape:
		if DefaultShapeCustomImg != nil {
			ebiten.SetCursorMode(ebiten.CursorModeHidden)
		} else {
			ebiten.SetCursorShape(ebiten.CursorShapeDefault)
		}
	case TextShape:
		if TextShapeCustomImg != nil {
			ebiten.SetCursorMode(ebiten.CursorModeHidden)
		} else {
			ebiten.SetCursorShape(ebiten.CursorShapeText)
		}
	case PointerShape:
		if PointerShapeCustomImg != nil {
			ebiten.SetCursorMode(ebiten.CursorModeHidden)
		} else {
			ebiten.SetCursorShape(ebiten.CursorShapePointer)
		}
	case CrosshairShape:
		if CrosshairShapeCustomImg != nil {
			ebiten.SetCursorMode(ebiten.CursorModeHidden)
		} else {
			ebiten.SetCursorShape(ebiten.CursorShapeCrosshair)
		}
	default:
		logz.Panicln("SetCursorShape", "cursor shape not recognized:", NextShape)
	}
}

func DrawCursor(screen *ebiten.Image) {
	x, y := ebiten.CursorPosition()
	switch CurrentShape {
	case DefaultShape:
		if DefaultShapeCustomImg != nil {
			rendering.DrawImage(screen, DefaultShapeCustomImg, float64(x), float64(y), 0)
		}
	case TextShape:
		if TextShapeCustomImg != nil {
			rendering.DrawImage(screen, TextShapeCustomImg, float64(x), float64(y), 0)
		}
	case PointerShape:
		if PointerShapeCustomImg != nil {
			rendering.DrawImage(screen, PointerShapeCustomImg, float64(x), float64(y), 0)
		}
	case CrosshairShape:
		if CrosshairShapeCustomImg != nil {
			rendering.DrawImage(screen, CrosshairShapeCustomImg, float64(x), float64(y), 0)
		}
	default:
		logz.Panicln("DrawCursor", "cursor shape not recognized:", CurrentShape)
	}
}
