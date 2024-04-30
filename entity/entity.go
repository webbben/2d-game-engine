package entity

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/webbben/2d-game-engine/config"
	m "github.com/webbben/2d-game-engine/model"
	"github.com/webbben/2d-game-engine/rendering"
	"github.com/webbben/2d-game-engine/tileset"

	"github.com/hajimehoshi/ebiten/v2"
)

// builds an animation def using the image file base name given.
//
// for example, if there are images "down_0", "down_1", ... "down_5", you can build the animation def for these by
// entering a baseName of "down". This will find all files that start with this base name and detect the first number and the last number present.
//
// baseName: string that all images in an animation are prefixed with. it should not include the "_", but we expect each numbered image to be of the format baseName_n.
func BuildAnimationDef(baseName string, tileset map[string]*ebiten.Image) m.AnimationDef {
	start := -1
	end := -1
	for key := range tileset {
		if strings.HasPrefix(key, baseName) {
			// we found an image starting with the basename - try to parse out which number it is
			parts := strings.Split(key, "_")
			if len(parts) != 2 {
				fmt.Println("error parsing animation def: image file not of the correct naming format. base name:", baseName, "image name:", key)
				return m.AnimationDef{}
			}
			number, err := strconv.Atoi(parts[1])
			if err != nil {
				fmt.Println("error parsing animation def: failed to parse image number. base name:", baseName, "image name:", key)
				return m.AnimationDef{}
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
			return m.AnimationDef{}
		}
	}
	fmt.Printf("%s: from %s_%v to %s_%v\n", baseName, baseName, start, baseName, end)
	return m.AnimationDef{
		FrameBase: baseName,
		Start:     start,
		End:       end,
	}
}

// sets the next animation frame for the given animation
func (e *Entity) setNextAnimationFrame(animationName string) {
	animationDef, ok := e.Animations[animationName]
	if !ok {
		fmt.Println("failed to get animation def for animation name:", animationName)
		return
	}
	// starting a new animation?
	if e.CurrentAnimation != animationName {
		e.CurrentAnimation = animationName
		e.AnimationStep = animationDef.Start
	} else {
		e.AnimationStep++
	}
	// past end of animation?
	if e.AnimationStep > animationDef.End {
		e.AnimationStep = animationDef.Start
	} else if e.AnimationStep < animationDef.Start {
		e.AnimationStep = animationDef.Start
	}
	// set the next frame
	nextFrameKey := fmt.Sprintf("%s_%v", animationDef.FrameBase, e.AnimationStep)
	nextFrame, ok := e.Frames[nextFrameKey]
	if !ok {
		fmt.Println("failed to get image for animation frame:", nextFrameKey)
		return
	}
	e.CurrentFrame = nextFrame
}

const (
	Old_Man_01 string = "old_man_01"
)

type Entity struct {
	EntKey   string // key of the ent def this entity uses
	EntID    string // unique ID of entity
	EntName  string // name of this type of entity (usually a sort of generalization, like "old man", or "town guard")
	Category string // category this entity falls in; such as soldier, villager, academic, bandit, etc.
	m.CharacterInfo
	IsHuman        bool
	IsInteractable bool
	m.Position
	Frames       map[string]*ebiten.Image
	CurrentFrame *ebiten.Image // the frame that is rendered for this entity when the screen is drawn
	m.Personality
	Animations       map[string]m.AnimationDef // possible animations for this entity
	CurrentAnimation string                    // animation this entity is currently in
	AnimationStep    int                       // current step in the animation
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
	animationMap := make(map[string]m.AnimationDef)
	for _, anim := range entityDefJsonData.Animations {
		animationMap[anim] = BuildAnimationDef(anim, frames)
	}
	ent := Entity{
		EntKey: entKey,
		CharacterInfo: m.CharacterInfo{
			FirstName: firstName,
			LastName:  lastName,
		},
		IsHuman:      entityDefJsonData.IsHuman,
		Category:     entityDefJsonData.Category,
		Frames:       frames,
		CurrentFrame: frames["down_0"],
		Position: m.Position{
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
