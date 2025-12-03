# Entity Body

The entity body contains all logic and animation mechanics for an entity being rendered in the game world.
It is meant to contain all the logic for these things so it can be easily reusable as needed.

## Body Structure

The `Body` consists of several `BodyPart` pieces, each of which represent a different dynamic part of the body (such as hair, eyes, arms, equipped headwear, etc).
A `BodyPart` essentially just defines the animation sequences for each animation (walking, running, etc). They contain all the logic, flags, etc used in making the animation work.
The `Body` basically orchestrates all of this so that the full animation progresses, and it also exposes functions to calling code for things like initiating a new animation.

If a bodypart doesn't implement a certain animation, it can use the `skip` flag. For example, since there is no animation occurring for the eyes while running, the eyes set would use the Skip flag for the run animation.

## SelectedPartDef

This struct represents, basically, a "currently selected part" for the body. The `BodyPart` itself defines the expected structure of all animations, but the `SelectedPartDef`
actually tells it which tileset to use, where each physical direction's start point (index in tileset) is, etc. Gives the actual "image definition" to a body part.

## Animations

There are a handful of animations we can define. Pretty much all of them are supposed to behave only in one specific sequence of possible frames, and so those sequences
are defined in the `BodyPart`. For example, since the walk animation has a singularly defined sequence of frames for a human entity's body, that means all other body parts must follow
this same sequence/pattern when defining the walk animation; if the body has 4 frames in a walk animation, but the shirt bodyPart only has 2, then the animation will get out of sync
and look bad. So, since shirts will *always* need to have the same overall pattern as the underlying body, we define these "rigid animation" sequences in the body part.

### Idle Animation

However, some animations do **not** have rigidly defined sequences that are always used. This is especially true for `auxiliary items` (items used in the left hand, such as torches).
Some items, like the torch, have an `Idle` animation that runs when the player is standing there without moving anywhere. The torch flickers and sputters, like a flame would.
But, the shield auxiliary item does not have an Idle animation, so this causes a problem to our previously described system of always having the same sequence for specific animations.

So, to handle this, we treat the `Idle` animation differently. Since the Idle animation doesn't involve any body movement (body stays still while, for example, a torch sputters)
we don't need to be concerned about everything "syncing up" with the body (as we do with a run animation or something).
The SelectedPartDef will define the sequence of the idle animation, which will give different parts more flexibility in defining animations.
This could be a bit confusing, since the other animation sequences are all defined via the BodyPart, but I think it's the best option. If, in the end, it seems like it makes more sense
to just define all animation sequences in the SelectedPartDef, then we can migrate all the others to that same system. But I don't see why other animations would need to have
different sequences for each body part.
