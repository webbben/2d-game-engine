package body

import (
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/webbben/2d-game-engine/internal/logz"
)

// BodyPartSet represents an individual (visual) part of an entity, such as the body, arms, equiped items, etc.
// It is essentially a collection of animations.
type BodyPartSet struct {
	Name        string
	sourceSet   bool            // indicates if a source has been set yet (tilesetSrc, etc)
	PartSrc     SelectedPartDef // tileset and image source definitions
	IsRemovable bool            // if true, this body part set can be removed or hidden (i.e. have None set to true).

	// animation definitions

	animIndex          int  // index or "step" of the animation we are currently on
	reachedLastFrame   bool // used to detect when an animation has finished (if all sets are at last frame, entire animation is done)
	IdleAnimation      Animation
	WalkAnimation      Animation
	RunAnimation       Animation
	SlashAnimation     Animation
	BackslashAnimation Animation
	ShieldAnimation    Animation
	HasUp              bool // if true, this set has an "up" direction animation. some don't since they will be covered by the body (such as eyes)

	img *ebiten.Image `json:"-"`
}

func (bps BodyPartSet) HasLoaded() bool {
	return bps.sourceSet
}

type BodyPartSetParams struct {
	IsBody      bool // if true, this body part set will be treated as the main body set. this allows things like StepsOffsetY to be used.
	HasUp       bool // if true, this set has animation frames for "up". some may not, since they might be covered up (e.g. the eyes set)
	IsRemovable bool // if true, this set can be removed or hidden from rendering
	Name        string
}

// NewBodyPartSet creates a new body part (e.g. arms, body, equiped head, etc) to which a part def (i.e. the actual data for animations, etc) can be set.
// a BodyPartSet is essentially just the "slot" where a part definition can be placed. Contains some high level rules, like if the part has an "up" direction, is removable, etc.
func NewBodyPartSet(params BodyPartSetParams) BodyPartSet {
	if params.IsBody && params.IsRemovable {
		panic("body set cannot be removed")
	}
	if params.Name == "" {
		panic("must set name for bodyPartSet (for debugging purposes)")
	}

	bps := BodyPartSet{
		HasUp:       params.HasUp,
		IsRemovable: params.IsRemovable,
		Name:        params.Name,
		// all parts start off as being "none"/disabled. a partSrc can be added later.
		PartSrc: SelectedPartDef{None: true},
	}

	return bps
}

func (bps BodyPartSet) animationDebugString(dir byte) string {
	if bps.PartSrc.None {
		return fmt.Sprintf("[%s] NONE", bps.Name)
	}
	if !bps.HasUp && dir == 'U' {
		return fmt.Sprintf("[%s] No Up", bps.Name)
	}

	s := fmt.Sprintf("[%s] animIndex: %v lastframe: %v", bps.Name, bps.animIndex, bps.reachedLastFrame)

	return s
}

func (bps BodyPartSet) validate() {
	if bps.PartSrc.None {
		return
	}
	if !bps.sourceSet {
		panic("source not set!")
	}
	if bps.Name == "" {
		panic("no name set")
	}
	bps.WalkAnimation.validate()
	bps.RunAnimation.validate()
	bps.SlashAnimation.validate()
	bps.BackslashAnimation.validate()
	bps.IdleAnimation.validate()
	bps.ShieldAnimation.validate()
}

func (bps *BodyPartSet) unsetAllImages() {
	bps.WalkAnimation.reset()
	bps.RunAnimation.reset()
	bps.SlashAnimation.reset()
	bps.BackslashAnimation.reset()
	bps.IdleAnimation.reset()
	bps.ShieldAnimation.reset()
	bps.img = nil
	logz.Println(bps.Name, "unsetting all images")
}

func (bps *BodyPartSet) setImageSource(def SelectedPartDef, stretchX, stretchY int, aux bool) {
	bps.PartSrc = def
	bps.sourceSet = true
	bps.load(stretchX, stretchY, aux)
}

func (set *BodyPartSet) load(stretchX, stretchY int, aux bool) {
	set.unsetAllImages()
	logz.Println(set.Name, "loading bodyPartSet")

	set.animIndex = 0

	if set.PartSrc.None {
		return
	}
	// leaving this below the above None check, since it makes it easier to define a None set without having to actually do the load process.
	if !set.sourceSet {
		panic("source not set before attempting to load")
	}

	set.WalkAnimation.Name = fmt.Sprintf("%s/walk", set.Name)
	set.RunAnimation.Name = fmt.Sprintf("%s/run", set.Name)
	set.SlashAnimation.Name = fmt.Sprintf("%s/slash", set.Name)
	set.BackslashAnimation.Name = fmt.Sprintf("%s/backslash", set.Name)
	set.IdleAnimation.Name = fmt.Sprintf("%s/idle", set.Name)
	set.ShieldAnimation.Name = fmt.Sprintf("%s/shield", set.Name)
	set.PartSrc.IdleAnimation.Name = "idle"

	set.WalkAnimation.load(set.PartSrc.WalkAnimation, aux, set.HasUp, set.PartSrc.FlipRForL, stretchX, stretchY)
	set.RunAnimation.load(set.PartSrc.RunAnimation, aux, set.HasUp, set.PartSrc.FlipRForL, stretchX, stretchY)
	set.SlashAnimation.load(set.PartSrc.SlashAnimation, aux, set.HasUp, set.PartSrc.FlipRForL, stretchX, stretchY)
	set.BackslashAnimation.load(set.PartSrc.BackslashAnimation, aux, set.HasUp, set.PartSrc.FlipRForL, stretchX, stretchY)
	set.ShieldAnimation.load(set.PartSrc.ShieldAnimation, aux, set.HasUp, set.PartSrc.FlipRForL, stretchX, stretchY)
	set.IdleAnimation.load(set.PartSrc.IdleAnimation, aux, set.HasUp, set.PartSrc.FlipRForL, stretchX, stretchY)

	set.validate()
}

func (set *BodyPartSet) setCurrentFrame(dir byte, animationName string) {
	if animationName == "" {
		panic("animation is unset")
	}
	if set.PartSrc.None {
		set.img = nil
		return
	}
	if dir == 'U' && !set.HasUp {
		set.img = nil
		return
	}

	switch animationName {
	case AnimWalk:
		set.img = set.WalkAnimation.getFrame(dir, set.animIndex)
	case AnimRun:
		set.img = set.RunAnimation.getFrame(dir, set.animIndex)
	case AnimSlash:
		set.img = set.SlashAnimation.getFrame(dir, set.animIndex)
	case AnimBackslash:
		set.img = set.BackslashAnimation.getFrame(dir, set.animIndex)
	case AnimIdle:
		set.img = set.IdleAnimation.getFrame(dir, set.animIndex)
	case AnimShield:
		set.img = set.ShieldAnimation.getFrame(dir, set.animIndex)
	default:
		panic("unrecognized animation name: " + animationName)
	}
}

func (set BodyPartSet) getCurrentYOffset(animationName string) int {
	switch animationName {
	case AnimWalk:
		if len(set.WalkAnimation.StepsOffsetY) > 0 {
			return set.WalkAnimation.StepsOffsetY[set.animIndex]
		}
	case AnimRun:
		if len(set.RunAnimation.StepsOffsetY) > 0 {
			return set.RunAnimation.StepsOffsetY[set.animIndex]
		}
	case AnimSlash:
		if len(set.SlashAnimation.StepsOffsetY) > 0 {
			return set.SlashAnimation.StepsOffsetY[set.animIndex]
		}
	case AnimBackslash:
		if len(set.BackslashAnimation.StepsOffsetY) > 0 {
			return set.BackslashAnimation.StepsOffsetY[set.animIndex]
		}
	case AnimIdle:
		if len(set.IdleAnimation.StepsOffsetY) > 0 {
			return set.IdleAnimation.StepsOffsetY[set.animIndex]
		}
	case AnimShield:
		if len(set.ShieldAnimation.StepsOffsetY) > 0 {
			return set.ShieldAnimation.StepsOffsetY[set.animIndex]
		}
	}

	return 0
}

func (set *BodyPartSet) nextFrame(animationName string) {
	if set.PartSrc.None {
		return
	}
	if !set.sourceSet {
		panic("source not set!")
	}
	if animationName == "" {
		logz.Panic("called nextFrame on empty animation. should this be the idle animation?")
	}

	set.animIndex++
	set.reachedLastFrame = false
	numSteps := 0
	// For now, we assume all directions of an animation have the same length. if this changes, we need to redo this logic
	switch animationName {
	case AnimWalk:
		// this body part skips this animation
		if set.WalkAnimation.Skip {
			return
		}
		numSteps = len(set.WalkAnimation.L)
	case AnimRun:
		if set.RunAnimation.Skip {
			return
		}
		numSteps = len(set.RunAnimation.L)
	case AnimSlash:
		if set.SlashAnimation.Skip {
			return
		}
		numSteps = len(set.SlashAnimation.L)
	case AnimBackslash:
		if set.BackslashAnimation.Skip {
			return
		}
		numSteps = len(set.BackslashAnimation.L)
	case AnimIdle:
		if set.IdleAnimation.Skip {
			return
		}
		numSteps = len(set.IdleAnimation.L)
	case AnimShield:
		if set.ShieldAnimation.Skip {
			return
		}
		numSteps = len(set.ShieldAnimation.L)
	default:
		logz.Panicln(set.Name, "nextFrame: animation name has no registered animation sequence:", animationName)
	}

	if numSteps == 0 {
		logz.Panicln(set.Name, "anim: ", animationName, "num steps is somehow 0")
		// set.animIndex = 0
		// set.reachedLastFrame = true
	}
	// ensure we don't go past the last frame - and mark this body part as done with the animation, if it has.
	if set.animIndex >= numSteps {
		set.reachedLastFrame = true
		set.animIndex = numSteps - 1
	}

	if set.animIndex < 0 {
		logz.Panicf("nextFrame: somehow animIndex became negative")
	}
}

// Remove completely removes the definition and images of the body part (clears PartSrc and all animation frames).
// should be used when actually removing an item from the entity's body.
func (set *BodyPartSet) Remove() {
	if !set.IsRemovable {
		logz.Panic("set is not removable!")
	}
	set.setImageSource(SelectedPartDef{None: true}, 0, 0, false)
}

// Hide hides the body part (without actually clearing PartSrc).
// basically meant for toggling on and off a part from rendering, such as when a weapon is sheathed or unsheathed.
func (set *BodyPartSet) Hide() {
	if !set.IsRemovable {
		logz.Panic("set is not removable!")
	}
	set.PartSrc.None = true
	set.setImageSource(set.PartSrc, 0, 0, false)
}
