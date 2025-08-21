package entity

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/webbben/2d-game-engine/internal/config"
	"github.com/webbben/2d-game-engine/internal/general_util"
	"github.com/webbben/2d-game-engine/internal/logz"
	"github.com/webbben/2d-game-engine/internal/model"
	"github.com/webbben/2d-game-engine/internal/tiled"
)

var (
	defaultWalkSpeed float64 = float64(config.TileSize) / 15
)

func GetDefaultWalkSpeed() float64 {
	if defaultWalkSpeed == 0 {
		panic("entity default walk speed is 0?")
	}
	return defaultWalkSpeed
}

type Entity struct {
	EntityInfo
	Loaded              bool                     `json:"-"` // if the entity has been loaded into memory fully yet
	Movement            Movement                 `json:"movement"`
	CurrentFrame        *ebiten.Image            `json:"-"`
	AnimationFrameMap   map[string]*ebiten.Image `json:"-"`
	AnimationFrameCount map[string]int           `json:"-"`
	Position

	FrameTilesetSources []string `json:"frame_tilesets"`

	World WorldContext `json:"-"`
}

// create a duplicate entity from this one.
// both entities will share the same references to tiles and animations and such, but will be able to have different
// positions, movement targets, etc.
// Useful for when you need a bunch of NPC entities of the same kind
func (e Entity) Duplicate() Entity {
	copyEnt := Entity{
		EntityInfo:          e.EntityInfo,
		AnimationFrameMap:   e.AnimationFrameMap,
		AnimationFrameCount: e.AnimationFrameCount,
		FrameTilesetSources: e.FrameTilesetSources,
		World:               e.World,
	}

	copyEnt.IsPlayer = false // cannot have duplicate players

	copyEnt.Movement = e.Movement
	copyEnt.Movement.TargetPath = []model.Coords{}
	copyEnt.Movement.TargetTile = model.Coords{}
	copyEnt.Movement.IsMoving = false

	return copyEnt
}

type WorldContext interface {
	Collides(c model.Coords) bool
	FindPath(start, goal model.Coords) []model.Coords
	MapDimensions() (width int, height int)
}

// Create an entity by opening an entity's definition JSON
func OpenEntity(source string) (Entity, error) {
	data, err := os.ReadFile(source)
	if err != nil {
		return Entity{}, fmt.Errorf("error reading entity source file: %w", err)
	}
	var ent Entity
	err = json.Unmarshal(data, &ent)
	if err != nil {
		return Entity{}, fmt.Errorf("error while unmarshalling entity json data: %w", err)
	}

	if ent.Movement.WalkSpeed == 0 {
		logz.Warnln(ent.DisplayName, "loaded entity does not have a walking speed; setting default value.")
		ent.Movement.WalkSpeed = GetDefaultWalkSpeed()
	}

	return ent, nil
}

// load fully entity data into memory for rendering in a map
func (e *Entity) Load() error {
	err := e.loadAnimationFrames()
	if err != nil {
		return err
	}

	// ensure first frame is set
	e.Movement.Direction = 'D'
	e.Movement.IsMoving = false
	e.Movement.IsRunning = false
	e.updateCurrentFrame()

	// set world context functions

	e.Loaded = true
	return nil
}

type EntityInfo struct {
	DisplayName string `json:"display_name"`
	ID          string `json:"id"`
	Source      string `json:"-"` // JSON source file for this entity
	IsPlayer    bool   `json:"-"` // flag indicating if this entity is the player
}

type Position struct {
	X, Y    float64      `json:"-"` // the exact position the entity is at on the map
	TilePos model.Coords `json:"-"` // the tile the entity is technically inside of
}

type Movement struct {
	IdleLeft       *ebiten.Image   `json:"-"`
	Left           []*ebiten.Image `json:"-"`
	LeftRun        []*ebiten.Image `json:"-"`
	IdleRight      *ebiten.Image   `json:"-"`
	Right          []*ebiten.Image `json:"-"`
	RightRun       []*ebiten.Image `json:"-"`
	IdleUp         *ebiten.Image   `json:"-"`
	Up             []*ebiten.Image `json:"-"`
	UpRun          []*ebiten.Image `json:"-"`
	IdleDown       *ebiten.Image   `json:"-"`
	Down           []*ebiten.Image `json:"-"`
	DownRun        []*ebiten.Image `json:"-"`
	Direction      byte            `json:"-"` // L R U D
	AnimationTimer int             `json:"-"` // counts the ticks until next animation frame
	AnimationFrame int             `json:"-"` // the current movement animation frame index

	CanRun bool `json:"can_run"`

	IsMoving  bool    `json:"-"`
	IsRunning bool    `json:"-"`
	WalkSpeed float64 `json:"walk_speed"` // value should be a TileSize / NumFrames calculation
	RunSpeed  float64 `json:"run_speed"`  // value should be a TileSize / NumFrames calculation
	Speed     float64 `json:"-"`          // actual speed the entity is moving at

	TargetTile model.Coords   `json:"-"` // next tile the entity is currently moving
	TargetPath []model.Coords `json:"-"` // path the entity is currently trying to travel on
}

func (e Entity) SaveJSON() error {
	if e.ID == "" {
		e.ID = general_util.GenerateUUID()
	}

	data, err := json.MarshalIndent(e, "", "  ")
	if err != nil {
		return fmt.Errorf("error while marshalling entity JSON data: %w", err)
	}
	filename := fmt.Sprintf("ent_%s.json", e.ID)
	path := filepath.Join(config.GameDefsPath(), "ent")
	os.MkdirAll(path, os.ModePerm)
	return os.WriteFile(filepath.Join(path, filename), data, os.ModePerm)
}

func (e *Entity) loadAnimationFrames() error {
	e.AnimationFrameMap = make(map[string]*ebiten.Image)
	e.AnimationFrameCount = make(map[string]int)

	for _, tilesetSource := range e.FrameTilesetSources {
		t, err := tiled.LoadTileset(tilesetSource)
		if err != nil {
			return fmt.Errorf("error while loading tileset data: %w", err)
		}

		err = t.GenerateTiles()
		if err != nil {
			return fmt.Errorf("error while generating tiles: %w", err)
		}

		// find all the animation frames
		for _, tile := range t.Tiles {
			frames := ""
			for _, p := range tile.Properties {
				if p.Name == "frames" {
					frames = p.GetStringValue()
				}
			}
			if frames != "" {
				// found animation frame; load image
				imagePath := filepath.Join(t.GeneratedImagesPath, fmt.Sprintf("%v.png", tile.ID))
				img, _, err := ebitenutil.NewImageFromFile(imagePath)
				if err != nil {
					return err
				}
				// the "frame" property expects the following format:
				// "animationName1:0,animationName2:x"
				// where "animationName" is the name of a specific animation, and the number is the frame number in that animation
				frameDefs := strings.Split(frames, ",")
				for _, frame := range frameDefs {
					vals := strings.Split(frame, ":")
					if len(vals) != 2 {
						return errors.New("frames property in tileset is malformed")
					}
					e.AnimationFrameMap[fmt.Sprintf("%s_%s", vals[0], vals[1])] = img

					// update frame count
					_, exists := e.AnimationFrameCount[vals[0]]
					if !exists {
						e.AnimationFrameCount[vals[0]] = 0
					}
					e.AnimationFrameCount[vals[0]]++
				}
			}
		}
	}

	// TODO add validation of loaded animation frames (e.g. verify there are no missing frames, verify frame counts, etc)

	return nil
}

func (e Entity) getAnimationFrame(animationName string, frameNumber int) *ebiten.Image {
	if e.AnimationFrameMap == nil {
		panic("entity animation frame map accessed before it was loaded")
	}

	key := fmt.Sprintf("%s_%v", animationName, frameNumber)
	img, exists := e.AnimationFrameMap[key]
	if !exists {
		fmt.Println("tried:", key)
		keys := make([]string, 0, len(e.AnimationFrameMap))
		for k := range e.AnimationFrameMap {
			keys = append(keys, k)
		}
		fmt.Println("existing keys:", strings.Join(keys, ", "))
		panic("accessed animation frame key for entity does not exist")
	}
	return img
}

func (e *Entity) SetPosition(c model.Coords) {
	e.TilePos = c
	e.X = float64(c.X) * float64(config.TileSize)
	e.Y = float64(c.Y) * float64(config.TileSize)
}
