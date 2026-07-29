package body

import (
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/webbben/2d-game-engine/data/defs"
	"github.com/webbben/2d-game-engine/logz"
)

// BodyPartSet represents an individual (visual) part of an entity, such as the body, arms, equiped items, etc.
// It is essentially a collection of animations.
type BodyPartSet struct {
	Name        string
	sourceSet   bool                 // indicates if a source has been set yet (tilesetSrc, etc)
	PartSrc     defs.SelectedPartDef // tileset and image source definitions
	IsRemovable bool                 // if true, this body part set can be removed or hidden (i.e. have None set to true).

	// animation definitions

	animIndex        int                  // index or "step" of the animation we are currently on
	reachedLastFrame bool                 // used to detect when an animation has finished (if all sets are at last frame, entire animation is done)
	Animations       map[string]Animation `json:"-"`
	HasUp            bool                 // if true, this set has an "up" direction animation. some don't since they will be covered by the body (such as eyes)

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
		PartSrc:    defs.SelectedPartDef{None: true},
		Animations: make(map[string]Animation),
	}

	return bps
}

func (bps BodyPartSet) animationDebugString() string {
	if bps.PartSrc.None {
		return fmt.Sprintf("[%s] NONE", bps.Name)
	}
	if !bps.HasUp {
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
	for _, a := range bps.Animations {
		a.validate()
	}

	// animation validation
	if bps.animIndex > 20 {
		// animation index is oddly high; is there a bug in detecting the end of an animation?
		// Note: if we want to support really long animations that have 20+ frames, just increase this upper threshold number
		logz.Println(bps.Name, bps.animationDebugString())
		logz.Panicln(bps.Name, "anim index is oddly high (>20). either we have an animation with a lot of frames, or something is going wrong with anim index.")
	}
}

func (bps *BodyPartSet) unsetAllImages() {
	for name := range bps.Animations {
		a := bps.Animations[name]
		a.reset()
		bps.Animations[name] = a
	}
	bps.img = nil
}

func (bps *BodyPartSet) setImageSource(def defs.SelectedPartDef, stretchX, stretchY int, aux bool) {
	bps.PartSrc = def
	bps.sourceSet = true
	bps.load(stretchX, stretchY, aux)
}

func (set *BodyPartSet) load(stretchX, stretchY int, aux bool) {
	set.unsetAllImages()

	set.animIndex = 0

	if set.PartSrc.None {
		return
	}
	// leaving this below the above None check, since it makes it easier to define a None set without having to actually do the load process.
	if !set.sourceSet {
		panic("source not set before attempting to load")
	}

	for _, name := range AllAnimations() {
		a := set.Animations[name]
		animParams := set.PartSrc.Animations[name]
		a.Name = fmt.Sprintf("%s/%s", set.Name, name)
		a.load(animParams, aux, set.HasUp, set.PartSrc.FlipRForL, stretchX, stretchY)
		set.Animations[name] = a
	}

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

	anim, ok := set.Animations[animationName]
	if !ok {
		panic("unrecognized animation name: " + animationName)
	}
	set.img = anim.getFrame(dir, set.animIndex)
}

func (set BodyPartSet) getCurrentYOffset(animationName string, direction byte) int {
	anim, ok := set.Animations[animationName]
	if !ok {
		return 0
	}
	return anim.GetOffsetY(direction, set.animIndex)
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

	anim, ok := set.Animations[animationName]
	if !ok {
		logz.Panicln(set.Name, "nextFrame: animation name has no registered animation sequence:", animationName)
	}

	set.reachedLastFrame = false

	if anim.Skip {
		set.reachedLastFrame = true
		return
	}

	numSteps := len(anim.L)

	// do below the skip check, so that if the animation is skipped we don't keep incrementing animIndex
	set.animIndex++

	if numSteps == 0 {
		logz.Panicln(set.Name, "anim: ", animationName, "num steps is somehow 0")
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
	set.setImageSource(defs.SelectedPartDef{None: true}, 0, 0, false)

	// sanity checks
	if !set.PartSrc.None {
		panic("removed body part set, but None is false")
	}
	if set.img != nil {
		panic("removed body part set, but img is not nil")
	}
}

// Hide hides the body part (without actually clearing PartSrc).
// basically meant for toggling on and off a part from rendering, such as when a weapon is sheathed or unsheathed.
func (set *BodyPartSet) Hide() {
	if !set.IsRemovable {
		logz.Panic("set is not removable!")
	}
	set.PartSrc.None = true
	set.setImageSource(set.PartSrc, 0, 0, false)

	// sanity checks
	if !set.PartSrc.None {
		panic("hid body part set, but None is false")
	}
	if set.img != nil {
		panic("hid removed body part set, but img is not nil")
	}
}
