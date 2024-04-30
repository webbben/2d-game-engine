package model

import "time"

type Position struct {
	X               float64
	Y               float64
	IsMoving        bool
	MovementSpeed   float64
	Direction_Horiz string    // "L"/"R" - the direction the player is moving on the horizontal axis
	Direction_Vert  string    // "U"/"D" - the direction the player is moving on the vertical axis
	Facing          string    // "U"/"D"/"L"/"R" - the direction the player is facing (visually)
	AnimStep        int       // the step of the animation we are on
	LastAnimStep    time.Time // the last time the character changed frames
}

type AnimationDef struct {
	FrameBase string // base name of the image files for the animation. e.g. "down"
	Start     int    // frame number the animation starts at. usually 0 or 1.
	End       int    // frame number the animation ends at.
}

// This character's personality traits, which may influence their actions or combat behavior
type Personality struct {
	// aggression level of this entity, mainly pertaining to eagerness to attack
	//
	// 0: will never attack others, no matter the circumstance
	// (ex: Philosopher, elderly man)
	//
	// 1: will never attack others, unless they are personally provoked enough
	// (ex: )
	//
	// 2: also may attack if witnessing a highly offensive action (e.g. their friend is harmed)
	//
	// 3: also may attack if encounters an enemy faction's entity
	// (ex: soldier)
	//
	// 4: also may attack if encounters an entity of which it has a very low disposition (hates)
	// (ex: )
	//
	// 5: will attack any entity it encounters (maximum aggression)
	// (ex: murderous bandit, lion)
	Aggression int
	// the bravery level an entity has in hostile situations, mainly pertaining to its willingness to defend itself
	//
	// 0: will always flee from all entities, except for explicitly friendly (same faction, etc) entities. never defends itself.
	// (ex: wild deer, bird)
	//
	// 1: will always flee from hostile entities, and never defend itself.
	// (ex: elderly woman)
	//
	// 2: will defend itself from hostile entities as long as the enemy is similar strength as it (otherwise will flee).
	// (ex: peasant man, merchant)
	//
	// 3: will defend itself from hostile entities, as long as the enemy isn't *overwhelmingly stronger (*a centurion fighting a peasant is overwhelmingly stronger).
	// (ex: town guard, sword for hire)
	//
	// 4: will defend itself from hostile entities, regardless of how strong.
	// (ex: legionary, most soldiers)
	Bravery int
	// the discipline an entity has in combat situations, mainly pertaining to its discipline to avoid fleeing from battle
	//
	// 0: will flee from combat if attacked whatsoever (including attacks that deal no damage, or general hostile effects)
	//
	// 1: will flee from combat if health decreases by 15%
	// (ex: commoner in a city)
	//
	// 2: will flee from combat if health decreases by 50%
	// (ex: common bandit)
	//
	// 3: will flee from combat if health decreases by 75%
	// (ex: general legionary, town guard)
	//
	// 4: will flee from combat if health decreases by 90%
	// (ex: legion officers, triarii, principes)
	//
	// 5: will never flee from combat, no matter how close to death
	// (ex: centurion, professional gladiator)
	Discipline int
	// the morality of an entity, mainly pertaining to how this entity will behave towards criminal acts, violence, etc, and if it will "do the right thing"
	//
	// 0: this entity has no morality; it will rob or attack any other entities at any opportunity, and will never help a victim.
	// (ex: murderous thug)
	//
	// 1: this entity has low morality; it may rob or attack other entities as at most opportunities, but may not target exceptionally weak entities. it will never help a victim.
	// (ex: common bandit)
	//
	// 2: this entity is an outlaw, but has some morality; it may rob entities at some opportunities, but will not target entities that would be too dishonorable (elderly, women, etc).
	// it may help victims that are especially vulnerable (elderly, women, etc). This entity will never report crimes to authorities.
	// (ex: rogue)
	//
	// 3: this entity is neutral; they usually won't commit any crimes, but have the potential to commit some petty crimes on rare occasions. They won't intervene to help victims unless they are especially vulnerable.
	// this entity will not report crimes to authorities unless they are egregious (e.g. unprovoked murder).
	// (ex: beggars, drunks)
	//
	// 4: this entity is a good person; they won't commit any crimes, and will intervene in most cases to help victims, unless the crime is petty. this entity will report serious crimes to authorities (e.g. unprovoked assault, robbery)
	// (ex: commoner, )
	//
	// 5: this entity is highly moral; they won't commit crimes, and will always intervene to help any victim. this entity will report any crime it sees to authorities
	// (ex: town guard)
	Morality int
}

// Information about the individual character
type CharacterInfo struct {
	FirstName string // the individual character's first name
	LastName  string // the individual character's last name
}
