package model

import (
	"fmt"
	"math"

	"github.com/webbben/2d-game-engine/internal/config"
)

/*

Types related to world layouts and navigating maps, etc.

*/

type directions struct {
	Left  byte
	Right byte
	Up    byte
	Down  byte
}

var Directions directions = directions{
	Left:  'L',
	Right: 'R',
	Up:    'U',
	Down:  'D',
}

func GetOppositeDirection(dir byte) byte {
	switch dir {
	case Directions.Left:
		return Directions.Right
	case Directions.Right:
		return Directions.Left
	case Directions.Up:
		return Directions.Down
	case Directions.Down:
		return Directions.Up
	default:
		panic("invalid direction given!")
	}
}

// gets the direction that b lies in relative to a
func GetRelativeDirection(a, b Coords) byte {
	dx := b.X - a.X
	dy := b.Y - a.Y
	if math.Abs(float64(dx)) > math.Abs(float64(dy)) {
		if dx < 0 {
			return Directions.Left
		} else {
			return Directions.Right
		}
	} else {
		if dy < 0 {
			return Directions.Up
		} else {
			return Directions.Down
		}
	}
}

// tile-based coordinate position in a room
type Coords struct {
	X int `json:"x"`
	Y int `json:"y"`
}

func (c Coords) String() string {
	return fmt.Sprintf("[X: %v, Y: %v]", c.X, c.Y)
}

func (c Coords) Equals(other Coords) bool {
	return c.X == other.X && c.Y == other.Y
}

// returns a copy of the coords struct, to avoid reference ties
func (c Coords) Copy() Coords {
	return Coords{X: c.X, Y: c.Y}
}

func (c Coords) GetAdj(direction byte) Coords {
	adj := c.Copy()
	switch direction {
	case 'L':
		adj.X--
	case 'R':
		adj.X++
	case 'U':
		adj.Y--
	case 'D':
		adj.Y++
	default:
		panic("GetAdj: invalid direction passed")
	}
	return adj
}

func (c Coords) IsAdjacent(otherPos Coords) bool {
	dx := math.Abs(float64(c.X) - float64(otherPos.X))
	dy := math.Abs(float64(c.Y) - float64(otherPos.Y))
	return dx+dy == 1
}

func ConvertPxToTilePos(x, y int) Coords {
	return Coords{
		X: x / config.TileSize,
		Y: y / config.TileSize,
	}
}

func (c Coords) WithinBounds(x1, x2, y1, y2 int) bool {
	return c.X >= x1 && c.X <= x2 && c.Y >= y1 && c.Y <= y2
}

type Vec2 struct {
	X, Y float64
}

func (v Vec2) Add(o Vec2) Vec2      { return Vec2{v.X + o.X, v.Y + o.Y} }
func (v Vec2) Sub(o Vec2) Vec2      { return Vec2{v.X - o.X, v.Y - o.Y} }
func (v Vec2) Scale(s float64) Vec2 { return Vec2{v.X * s, v.Y * s} }
func (v Vec2) Len() float64         { return math.Sqrt(v.X*v.X + v.Y*v.Y) }
func (v Vec2) Equals(o Vec2) bool   { return v.X == o.X && v.Y == o.Y }

func (v Vec2) Normalize() Vec2 {
	l := v.Len()
	if l == 0 {
		return Vec2{0, 0}
	}
	return Vec2{v.X / l, v.Y / l}
}

func (pos Vec2) Dist(target Vec2) float64 {
	dir := target.Sub(pos)
	return dir.Len()
}

type Rect struct {
	X, Y, W, H float64
}

func NewRect(x, y, w, h float64) Rect {
	return Rect{X: x, Y: y, W: w, H: h}
}

func (r Rect) String() string {
	return fmt.Sprintf("x: %v y: %v w: %v h: %v", r.X, r.Y, r.W, r.H)
}

func (r Rect) Intersects(other Rect) bool {
	return r.X < other.X+other.W &&
		r.X+r.W > other.X &&
		r.Y < other.Y+other.H &&
		r.Y+r.H > other.Y
}

func (r Rect) Within(x, y int) bool {
	return x > int(r.X) && x < int(r.X+r.W) && y > int(r.Y) && y < int(r.Y+r.H)
}

type IntersectionResult struct {
	Intersects bool
	Dx, Dy     float64
}

func (ir IntersectionResult) Int() int {
	if ir.Intersects {
		return 1
	}
	return 0
}

func (ir IntersectionResult) assert() {
	// intersections can have 0 intersection area, since we currently do that for map edges
	// but non intersections must have 0 intersection area
	if !ir.Intersects {
		if ir.Dx != 0 || ir.Dy != 0 {
			panic("not an intersection, but there is an intersection area (dx or dy are not 0)")
		}
	}
}

type CollisionResult struct {
	TopLeft     IntersectionResult
	TopRight    IntersectionResult
	BottomLeft  IntersectionResult
	BottomRight IntersectionResult
	Other       IntersectionResult // for general use when not specifically corner related
}

func (c CollisionResult) String() string {
	if !c.Collides() {
		return "(No Collisions)"
	}
	coll := []string{}
	if c.TopLeft.Intersects {
		coll = append(coll, "TL")
	}
	if c.TopRight.Intersects {
		coll = append(coll, "TR")
	}
	if c.BottomLeft.Intersects {
		coll = append(coll, "BL")
	}
	if c.BottomRight.Intersects {
		coll = append(coll, "BR")
	}
	if c.Other.Intersects {
		coll = append(coll, "O")
	}
	return fmt.Sprintf("%v", coll)
}

func (c CollisionResult) Collides() bool {
	return c.TopLeft.Intersects ||
		c.TopRight.Intersects ||
		c.BottomLeft.Intersects ||
		c.BottomRight.Intersects ||
		c.Other.Intersects
}

func (c CollisionResult) Assert() {
	c.TopLeft.assert()
	c.TopRight.assert()
	c.BottomLeft.assert()
	c.BottomRight.assert()
}

// gets the area that intersects between two rects.
// dx and dy area always positive (or 0 if no intersection)
func (r Rect) IntersectionArea(other Rect) IntersectionResult {
	res := IntersectionResult{}
	res.Intersects = r.Intersects(other)
	if res.Intersects {
		// find intersecting area
		// for some reason these calculations come out 1 short, so adding 1 here seems to fix it
		if r.X < other.X {
			res.Dx = (r.X + r.W) - other.X
		} else {
			res.Dx = (other.X + other.W) - r.X
		}
		if r.Y < other.Y {
			res.Dy = (r.Y + r.H) - other.Y
		} else {
			res.Dy = (other.Y + other.H) - r.Y
		}
	}
	return res
}
