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

func NewVec2(x, y float64) Vec2 {
	return Vec2{X: x, Y: y}
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
	if w <= 0 {
		panic("width is <= 0")
	}
	if h <= 0 {
		panic("height is <= 0")
	}
	return Rect{X: x, Y: y, W: w, H: h}
}

func (r Rect) GetCenter() (x, y float64) {
	return r.X + (r.W / 2), r.Y + (r.H / 2)
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

func (r Rect) VecWithin(v Vec2) bool {
	return r.Within(int(v.X), int(v.Y))
}

type IntersectionResult struct {
	Intersects bool
	Dx, Dy     float64

	// where the intersection came from, from the perspective of "r" (the main rect the Intersection function was called on)
	FromTL, FromTR, FromBL, FromBR bool
}

func (ir *IntersectionResult) MergeIntersectionResult(other IntersectionResult) {
	if !other.Intersects {
		return
	}
	if !ir.Intersects {
		ir.Intersects = other.Intersects
		ir.Dx = other.Dx
		ir.Dy = other.Dy
		return
	}
	ir.Dx = max(ir.Dx, other.Dx)
	ir.Dy = max(ir.Dy, other.Dy)
}

func (ir IntersectionResult) String() string {
	if !ir.Intersects {
		return ""
	}
	return fmt.Sprintf("dx: %v dy: %v", ir.Dx, ir.Dy)
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

// merges the two collision results, basically by taking the "maximum" collision for each corner.
func (cr *CollisionResult) MergeOtherCollisionResult(other CollisionResult) {
	if !other.Collides() {
		return
	}
	cr.TopLeft.MergeIntersectionResult(other.TopLeft)
	cr.TopRight.MergeIntersectionResult(other.TopRight)
	cr.BottomLeft.MergeIntersectionResult(other.BottomLeft)
	cr.BottomRight.MergeIntersectionResult(other.BottomRight)
	cr.Other.MergeIntersectionResult(other.Other)
}

func (c CollisionResult) String() string {
	if !c.Collides() {
		return "(No Collisions)"
	}
	coll := []string{}
	if c.TopLeft.Intersects {
		coll = append(coll, fmt.Sprintf("TL (%s)", c.TopLeft))
	}
	if c.TopRight.Intersects {
		coll = append(coll, fmt.Sprintf("TR (%s)", c.TopRight))
	}
	if c.BottomLeft.Intersects {
		coll = append(coll, fmt.Sprintf("BL (%s)", c.BottomLeft))
	}
	if c.BottomRight.Intersects {
		coll = append(coll, fmt.Sprintf("BR (%s)", c.BottomRight))
	}
	if c.Other.Intersects {
		coll = append(coll, fmt.Sprintf("O (%s)", c.Other))
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
		// for some reason these calculations come out 1 short, so adding 1 here seems to fix it (<- err, I guess this was undone? no 1's are being added)

		var left, above bool // "R is left/above Other"
		if r.X < other.X {
			res.Dx = (r.X + r.W) - other.X
			left = true
		} else {
			res.Dx = (other.X + other.W) - r.X
			left = false
		}
		if r.Y < other.Y {
			above = true
			res.Dy = (r.Y + r.H) - other.Y
		} else {
			above = false
			res.Dy = (other.Y + other.H) - r.Y
		}
		if res.Dx == 0 && res.Dy == 0 {
			panic("intersection, but no overlap found")
		}
		// record the corner of r where the intersection happened
		/*
			A quick diagram to explain how this works, in case the wording of "FromTL, FromBL" etc gets confused in the future.
			For R, Other is intersecting "FromBR" (From Bottom Right)
			So, the "From" directions are from the perspective of R (the Rect this function is called directly on).

			 -----------
			|           |
			|           |
			|    R      |
			|           |
			|     ------|-----
			 ----|------      |
			     |            |
			     |   Other    |
			     |            |
			      ------------
		*/
		corners := 0
		if left {
			if above {
				res.FromBR = true
				corners++
			} else {
				res.FromTR = true
				corners++
			}
		} else {
			if above {
				res.FromBL = true
				corners++
			} else {
				res.FromTL = true
				corners++
			}
		}
		if corners == 0 {
			panic("intersection, but no relative corner set")
		}
		if corners > 2 {
			panic("intersection somehow has more than 2 intersecting corners set. is R inside Other?")
		}
	}
	return res
}
