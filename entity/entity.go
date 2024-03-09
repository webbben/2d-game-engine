package entity

import (
	"ancient-rome/config"
	"ancient-rome/rendering"
	"ancient-rome/tileset"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
)

type AnimationDef struct {
	FrameBase string // base name of the image files for the animation. e.g. "down"
	Start     int    // frame number the animation starts at. usually 0 or 1.
	End       int    // frame number the animation ends at.
}

// builds an animation def using the image file base name given.
//
// for example, if there are images "down_0", "down_1", ... "down_5", you can build the animation def for these by
// entering a baseName of "down". This will find all files that start with this base name and detect the first number and the last number present.
//
// baseName: string that all images in an animation are prefixed with. it should not include the "_", but we expect each numbered image to be of the format baseName_n.
func BuildAnimationDef(baseName string, tileset map[string]*ebiten.Image) AnimationDef {
	start := -1
	end := -1
	for key := range tileset {
		if strings.HasPrefix(key, baseName) {
			// we found an image starting with the basename - try to parse out which number it is
			parts := strings.Split(key, "_")
			if len(parts) != 2 {
				fmt.Println("error parsing animation def: image file not of the correct naming format. base name:", baseName, "image name:", key)
				return AnimationDef{}
			}
			number, err := strconv.Atoi(parts[1])
			if err != nil {
				fmt.Println("error parsing animation def: failed to parse image number. base name:", baseName, "image name:", key)
				return AnimationDef{}
			}
			if start == -1 || number < start {
				start = number
			}
			if end == -1 || number > end {
				end = number
			}
		}
	}
	// confirm that there are no gaps in the numbers by attempting to iterate through each
	for i := start; i <= end; i++ {
		key := fmt.Sprintf("%s_%v", baseName, i)
		if _, ok := tileset[key]; !ok {
			fmt.Println("error parsing animation def: animation image frame seems to be missing. base name:", baseName, "missing frame:", key)
			return AnimationDef{}
		}
	}
	return AnimationDef{
		FrameBase: baseName,
		Start:     start,
		End:       end,
	}
}

const (
	Old_Man_01 string = "old_man_01"
)

type Position struct {
	X             float64
	Y             float64
	Facing        string
	IsMoving      bool
	MovementSpeed float64
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

type Entity struct {
	EntKey   string // key of the ent def this entity uses
	EntID    string // unique ID of entity
	EntName  string // name of this type of entity (usually a sort of generalization, like "old man", or "town guard")
	Category string // category this entity falls in; such as soldier, villager, academic, bandit, etc.
	CharacterInfo
	IsHuman        bool
	IsInteractable bool
	Position
	Frames       map[string]*ebiten.Image
	CurrentFrame *ebiten.Image // the frame that is rendered for this entity when the screen is drawn
	Personality
	Animations       map[string]AnimationDef // possible animations for this entity
	CurrentAnimation string                  // animation this entity is currently in
	AnimationStep    int                     // current step in the animation
}

// instantiates an entity of the given definition
func CreateEntity(entKey string, globalEntID string, firstName string, lastName string) *Entity {
	entityDefJsonData, err := loadEntityDefJson(entKey)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	tilesetKey := fmt.Sprintf("ent_%s", entKey)
	frames, err := tileset.LoadTileset(tilesetKey)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	animationMap := make(map[string]AnimationDef)
	for _, anim := range entityDefJsonData.Animations {
		animationMap[anim] = BuildAnimationDef(anim, frames)
	}
	ent := Entity{
		EntKey: entKey,
		CharacterInfo: CharacterInfo{
			FirstName: firstName,
			LastName:  lastName,
		},
		IsHuman:      entityDefJsonData.IsHuman,
		Category:     entityDefJsonData.Category,
		Frames:       frames,
		CurrentFrame: frames["down_0"],
		Position: Position{
			X: 0, Y: 50,
			MovementSpeed: entityDefJsonData.WalkSpeed,
		},
		Animations: animationMap,
	}
	return &ent
}

type EntityDefJsonData struct {
	ID         string   `json:"id"`
	Name       string   `json:"name"`
	Category   string   `json:"category"`
	IsHuman    bool     `json:"is_human"`
	WalkSpeed  float64  `json:"walk_speed"`
	Animations []string `json:"animations"`
}

func loadEntityDefJson(entKey string) (*EntityDefJsonData, error) {
	path := fmt.Sprintf("entity/entity_defs/%s.json", entKey)

	jsonData, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to load entity def at %s", path)
	}
	var entityData EntityDefJsonData
	if err = json.Unmarshal(jsonData, &entityData); err != nil {
		return nil, errors.New("json unmarshalling failed for entity def")
	}
	return &entityData, nil
}

func (e *Entity) Draw(screen *ebiten.Image, offsetX float64, offsetY float64) {
	drawX, drawY := rendering.GetImageDrawPos(e.CurrentFrame, e.Position.X, e.Position.Y, offsetX, offsetY)
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(drawX, drawY)
	op.GeoM.Scale(config.GameScale, config.GameScale)
	screen.DrawImage(e.CurrentFrame, op)
}
