package defs

import "fmt"

// BodyPartID is only for "skin" body parts (not for equiped items)
type BodyPartID string

type HSV struct {
	H, S, V float64
}

// SelectedPartDef represents the currently selected body part and it's individual definition
// Note: these are never hand written: always generated from the character builder
type SelectedPartDef struct {
	// unique id that represents this body part. only really used for body parts (not equipment, since those are tied to items).
	// this is the only field saved to JSON, since we will load the rest of the data using this ID with the definition manager
	ID        BodyPartID
	None      bool `json:"-"` // if true, this part will not be shown
	FlipRForL bool `json:"-"` // if true, instead of using an L source, we just flip the frames for right
	Female    bool // if true, this body part def is specific to female characters. Not used in the engine, but may be needed for game logic in character design.

	// Animation definitions keyed by animation name (e.g. "idle", "walk", "run", etc.)
	// See body.AllAnimations() for the canonical list of animation names.
	Animations map[string]AnimationParams `json:"-"`

	// body-specific props

	// FYI: these are not currently being used anymore (but remain functional) since we don't have other body options (tall, short, fat, etc) anymore.

	StretchX int `json:"-"` // amount to stretch hair and equip body on X axis. Defined here, this represents the value that is applied to ALL (applicable) parts - not to this one.
	StretchY int `json:"-"` // amount to stretch equip body on the Y axis. Defined here, this represents the value that is applied to ALL (applicable) parts - not to this one.
	OffsetY  int `json:"-"` // amount to offset positions of hair, eyes, equip body, etc on the Y axis

	// headwear-specific props

	CropHairToHead bool `json:"-"` // set to have hair not go outside the head image. used for helmets or certain hats.
}

// IsEqual checks if the two are equal. mainly used for validation.
func (def SelectedPartDef) IsEqual(other SelectedPartDef) bool {
	if def.None != other.None {
		fmt.Printf("'None' value is different: %v vs %v\n", def.None, other.None)
		return false
	}
	if def.FlipRForL != other.FlipRForL {
		fmt.Printf("'FlipRForL' value is different: %v vs %v\n", def.FlipRForL, other.FlipRForL)
		return false
	}
	if def.StretchX != other.StretchX || def.StretchY != other.StretchY || def.OffsetY != other.OffsetY {
		fmt.Println("stretch values or offset values are different:", def.StretchX, other.StretchX, def.StretchY, other.StretchY, def.OffsetY, other.OffsetY)
		return false
	}
	if def.CropHairToHead != other.CropHairToHead {
		fmt.Println("cropHairToHead values are different:", def.CropHairToHead, other.CropHairToHead)
		return false
	}
	if len(def.Animations) != len(other.Animations) {
		fmt.Printf("animation map size differs: %d vs %d\n", len(def.Animations), len(other.Animations))
		return false
	}
	for name, anim := range def.Animations {
		otherAnim, ok := other.Animations[name]
		if !ok {
			fmt.Printf("animation %s is missing in the other definition\n", name)
			return false
		}
		if !anim.IsEqual(otherAnim) {
			fmt.Printf("animation %s is not equal\n", name)
			return false
		}
	}
	return true
}

// BodyDef defines the body of the entity; excluding things like equiped items.
type BodyDef struct {
	BodyHSV HSV
	BodyID  BodyPartID
	EyesHSV HSV
	EyesID  BodyPartID
	HairHSV HSV
	HairID  BodyPartID
	ArmsID  BodyPartID
	LegsID  BodyPartID
}

type AnimationParams struct {
	Name                                      string
	Skip                                      bool
	TilesetSrc                                string
	TilesLeft, TilesRight, TilesUp, TilesDown []int // indices of each frame of the animation
	AuxLeft, AuxRight, AuxUp, AuxDown         []int // if aux influeces this animation, define the aux frames here

	StepsOffsetY []int

	// optional overrides for specific directions Y offset

	StepsOffsetYLeft  []int
	StepsOffsetYRight []int
	StepsOffsetYUp    []int
	StepsOffsetYDown  []int
}

func (ap AnimationParams) IsEqual(other AnimationParams) bool {
	if ap.Name != other.Name {
		return false
	}
	if ap.Skip != other.Skip {
		return false
	}
	if ap.TilesetSrc != other.TilesetSrc {
		return false
	}
	slicesEqual := func(a, b []int) bool {
		if len(a) != len(b) {
			return false
		}
		for i, v := range a {
			if b[i] != v {
				return false
			}
		}
		return true
	}

	if !slicesEqual(ap.TilesLeft, other.TilesLeft) {
		return false
	}
	if !slicesEqual(ap.TilesRight, other.TilesRight) {
		return false
	}
	if !slicesEqual(ap.TilesUp, other.TilesUp) {
		return false
	}
	if !slicesEqual(ap.TilesDown, other.TilesDown) {
		return false
	}
	if !slicesEqual(ap.AuxLeft, other.AuxLeft) {
		return false
	}
	if !slicesEqual(ap.AuxRight, other.AuxRight) {
		return false
	}
	if !slicesEqual(ap.AuxUp, other.AuxUp) {
		return false
	}
	if !slicesEqual(ap.AuxDown, other.AuxDown) {
		return false
	}
	if !slicesEqual(ap.StepsOffsetY, other.StepsOffsetY) {
		return false
	}
	if !slicesEqual(ap.StepsOffsetYLeft, other.StepsOffsetYLeft) {
		return false
	}
	if !slicesEqual(ap.StepsOffsetYRight, other.StepsOffsetYRight) {
		return false
	}
	if !slicesEqual(ap.StepsOffsetYUp, other.StepsOffsetYUp) {
		return false
	}
	if !slicesEqual(ap.StepsOffsetYDown, other.StepsOffsetYDown) {
		return false
	}
	return true
}

func (ap AnimationParams) DebugString() string {
	s := fmt.Sprintf("Name: %s TilesetSrc: %s Skip: %v\n", ap.Name, ap.TilesetSrc, ap.Skip)
	s += fmt.Sprintf("L: %v\nR: %v\nU: %v\nD: %v\n", ap.TilesLeft, ap.TilesRight, ap.TilesUp, ap.TilesDown)
	if len(ap.AuxLeft) != 0 {
		s += fmt.Sprintf("AuxL: %v\n", ap.AuxLeft)
	}
	if len(ap.AuxRight) != 0 {
		s += fmt.Sprintf("AuxR: %v\n", ap.AuxRight)
	}
	if len(ap.AuxUp) != 0 {
		s += fmt.Sprintf("AuxU: %v\n", ap.AuxUp)
	}
	if len(ap.AuxDown) != 0 {
		s += fmt.Sprintf("AuxD: %v\n", ap.AuxDown)
	}
	if len(ap.StepsOffsetY) != 0 {
		s += fmt.Sprintf("offY: %v\n", ap.StepsOffsetY)
	}
	if len(ap.StepsOffsetYLeft) != 0 {
		s += fmt.Sprintf("offYL: %v\n", ap.StepsOffsetYLeft)
	}
	if len(ap.StepsOffsetYRight) != 0 {
		s += fmt.Sprintf("offYR: %v\n", ap.StepsOffsetYRight)
	}
	if len(ap.StepsOffsetYUp) != 0 {
		s += fmt.Sprintf("offYU: %v\n", ap.StepsOffsetYUp)
	}
	if len(ap.StepsOffsetYDown) != 0 {
		s += fmt.Sprintf("offYD: %v\n", ap.StepsOffsetYDown)
	}
	return s
}
