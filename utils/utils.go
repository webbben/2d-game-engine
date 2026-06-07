// Package utils contains useful utility functions for use anywhere
package utils

import (
	"math"
	"math/rand"

	"github.com/google/uuid"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/webbben/2d-game-engine/display"
	"github.com/webbben/2d-game-engine/logz"
	"github.com/webbben/2d-game-engine/model"
	"golang.org/x/text/message"
)

func EuclideanDist(x1, y1, x2, y2 float64) float64 {
	// 2026-04-08 apparently I can "expand" this by just using regular multiplication (x*x instead of x^2)
	// but, for now just leaving as is (despite the annoying warning) since I don't like having to type out x2-x1 twice or putting it in a variable lol.
	return math.Sqrt(math.Pow(x2-x1, 2) + math.Pow(y2-y1, 2))
}

// EuclideanDistCoords is a euclidean distance function for coords structs, for ease of use
func EuclideanDistCoords(pointA, pointB model.Coords) float64 {
	return EuclideanDist(float64(pointA.X), float64(pointA.Y), float64(pointB.X), float64(pointB.Y))
}

// EuclideanDistCenter calculates euclidean distance based on the center of the given rects.
// gives a more "real" distance compared to getting distance of the top left corner of, say, two entities
func EuclideanDistCenter(r1, r2 model.Rect) float64 {
	r1.X += r1.W / 2
	r1.Y += r1.H / 2
	r2.X += r2.W / 2
	r2.Y += r2.H / 2
	return EuclideanDist(r1.X, r1.Y, r2.X, r2.Y)
}

func RandInt(min, max int) int {
	return rand.Intn(max-min+1) + min
}

func RoundToDecimal(val float64, decimals int) float64 {
	factor := math.Pow(10, float64(decimals))
	return math.Round(val*factor) / factor
}

// IsHovering returns true if the mouse cursor is hovering within the given coordinates box
func IsHovering(x1, y1, x2, y2 int) bool {
	x, y := ebiten.CursorPosition()
	return x >= x1 && x <= x2 && y >= y1 && y <= y2
}

// DetectMouse returns two bools: the first is true if the mouse is hovering over the given coordinates box, the second is true if the left mouse button is clicked within the box
func DetectMouse(x1, y1, x2, y2 int) (bool, bool) {
	if !IsHovering(x1, y1, x2, y2) {
		return false, false
	}
	return true, ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft)
}

func GenerateUUID() string {
	id := uuid.New()
	return id.String()
}

// RemoveIndexUnordered Removes the element at index i of slice s, without preserving order.
// Apparently much faster, but only use this if you don't care about the ordering of the elements
func RemoveIndexUnordered[T any](s []T, i int) []T {
	if i >= len(s) {
		logz.Panicf("index (%v) out of range (len=%v)", i, len(s))
	}
	s[i] = s[len(s)-1]
	return s[:len(s)-1]
}

// RemoveIndex Removes element at index i of slice s, preserving the original order.
// Apparently somewhat slow since it involves moving all the elements.
func RemoveIndex[T any](s []T, i int) []T {
	if i >= len(s) {
		logz.Panicf("index (%v) out of range (len=%v)", i, len(s))
	}
	return append(s[:i], s[i+1:]...)
}

func ConvertIntToCommaString(n int) string {
	p := message.NewPrinter(message.MatchLanguage("en"))
	return p.Sprintf("%d", n)
}

func Int(v int) *int {
	return &v
}

func Bool(v bool) *bool {
	return &v
}

func Str(s string) *string {
	return &s
}

func CenterInScreen(dx, dy int) (x, y float64) {
	screenDx := float64(display.SCREEN_WIDTH)
	screenDy := float64(display.SCREEN_HEIGHT)
	x = screenDx/2 - float64(dx)/2
	y = screenDy/2 - float64(dy)/2
	return x, y
}

// DifferentSigns - if either is 0, then it returns false (0 is considered the "same sign" as both positive and negative here)
func DifferentSigns(a, b float64) bool {
	return a != 0 && b != 0 && ((a < 0) != (b < 0))
}

func RoundDownToTile(v int, tilesize int) int {
	return v - (v % tilesize)
}

func RoundUpToTile(v int, tilesize int) int {
	if v%tilesize == 0 {
		return v
	}
	return RoundDownToTile(v+tilesize, tilesize)
}

func PanicAssert(assertTrue bool, failMsg string) {
	if !assertTrue {
		logz.Panicln("Assert", failMsg)
	}
}

func GetPositionNearMouse(distFromMouse int, dx, dy int) (x, y int) {
	// draw next to the mouse
	mouseX, mouseY := ebiten.CursorPosition()
	// make sure the window doesn't go off screen
	x = mouseX + distFromMouse
	y = mouseY + distFromMouse
	if x+dx > display.SCREEN_WIDTH {
		x = mouseX - distFromMouse - dx
	}
	if y+dy > display.SCREEN_HEIGHT {
		y = mouseY - distFromMouse - dy
	}

	return x, y
}
