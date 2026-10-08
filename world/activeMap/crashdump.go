package activemap

import (
	"fmt"

	"github.com/webbben/2d-game-engine/data/defs"
	"github.com/webbben/2d-game-engine/data/state"
	"github.com/webbben/2d-game-engine/entity"
	"github.com/webbben/2d-game-engine/model"
	"github.com/webbben/2d-game-engine/object"
)

// This file builds a JSON snapshot of the active map's state, intended to be written to disk when the game
// crashes, so the world state at the moment of the crash can be inspected afterwards.
//
// Two rules shape everything here:
//
//  1. It must never panic, and must never make the original crash worse. The snapshot is gathered while the
//     game is already unwinding from a panic, on whatever goroutine happened to fail, so a failure here could
//     either mask the real stack trace or deadlock the process outright.
//  2. It must not block. If the panic happened while a lock was held (e.g. inside a critical section), a
//     blocking read would hang forever instead of producing a report. Locks are always taken with TryLock,
//     and a section that can't be acquired is recorded in Skipped rather than waited on.
//
// CostMap is dumped in full rather than summarized, so an investigator can compute whatever they need from it
// (per-tile cost, reachability, why a tile is blocked) without the engine having to anticipate the question.
//
// Registering a provider for this is the job of the caller -- see world.NewWorld, which owns the only
// reference to the active map. Callers get the snapshot by calling BuildCrashStateSnapshot.

// CrashEntitySnapshot is the state shared by anything with a position and health: the player and every NPC.
type CrashEntitySnapshot struct {
	ID          string       `json:"id"`
	DisplayName string       `json:"display_name,omitempty"`
	Tile        model.Coords `json:"tile"`
	X           float64      `json:"x"`
	Y           float64      `json:"y"`
	Health      int          `json:"health"`
	MaxHealth   int          `json:"max_health"`
	IsDead      bool         `json:"is_dead"`
}

// CrashNPCSnapshot is an NPC's shared state plus its task, movement, and pose, which is usually what you need
// to explain where the NPC was and what it was doing.
type CrashNPCSnapshot struct {
	CrashEntitySnapshot
	TaskID              string `json:"task_id,omitempty"`
	TaskName            string `json:"task_name,omitempty"`
	TaskStatus          string `json:"task_status,omitempty"`
	IsSitting           bool   `json:"is_sitting"`
	IsSleeping          bool   `json:"is_sleeping"`
	DisableCollisions   bool   `json:"disable_collisions"`
	IsMoving            bool   `json:"is_moving"`
	MovementInterrupted bool   `json:"movement_interrupted"`
	PathLen             int    `json:"path_len"`
}

// CrashObjectSnapshot is an object's identity, geometry, and occupancy. Type-specific fields are omitted when
// they don't apply to the object's type, keeping the JSON flat rather than nested per type.
type CrashObjectSnapshot struct {
	ID          int             `json:"id"`
	Type        defs.ObjectType `json:"type"`
	DisplayName string          `json:"display_name,omitempty"`
	Rect        model.Rect      `json:"rect"`
	Collidable  bool            `json:"collidable"`
	LockID      string          `json:"lock_id,omitempty"`
	OwnerID     string          `json:"owner_id,omitempty"`

	// occupancy, for the object types a character can hold
	InUse  bool   `json:"in_use,omitempty"`
	UserID string `json:"user_id,omitempty"`

	// gate and door specifics
	GateIsOpen    bool       `json:"gate_is_open,omitempty"`
	DoorTargetMap defs.MapID `json:"door_target_map,omitempty"`
	DoorSpawnIdx  int        `json:"door_spawn_index,omitempty"`
}

// CrashStateSnapshot is the whole dump: which map we were on, what was on it, and what each character was doing.
type CrashStateSnapshot struct {
	MapID      defs.MapID `json:"map_id"`
	TileWidth  int        `json:"tile_width"`
	TileHeight int        `json:"tile_height"`
	CostMap    [][]int    `json:"cost_map,omitempty"`

	Player  *CrashEntitySnapshot  `json:"player,omitempty"`
	NPCs    []CrashNPCSnapshot    `json:"npcs"`
	Objects []CrashObjectSnapshot `json:"objects"`

	// Skipped records which sections couldn't be gathered, and why. A missing section is always better than a
	// dump that hung the process trying to produce it.
	Skipped []string `json:"skipped,omitempty"`
}

// BuildCrashStateSnapshot gathers the current state of the active map.
//
// Best-effort and never panics: every section is independently guarded, so one failure still leaves a partial
// dump of everything else. Safe to call from whichever goroutine panicked, without assuming which locks that
// goroutine may already be holding.
func (mi *ActiveMap) BuildCrashStateSnapshot() CrashStateSnapshot {
	snap := CrashStateSnapshot{
		NPCs:    []CrashNPCSnapshot{},
		Objects: []CrashObjectSnapshot{},
	}

	// Map header. mi.Map is nil until the map finishes loading, which a very early crash could beat us to.
	crashDumpSection(&snap, "map header", func() {
		snap.MapID = mi.MapID
		if mi.Map != nil {
			snap.TileWidth = mi.Map.Width
			snap.TileHeight = mi.Map.Height
		}
	})

	// Cost map. Read the atomic directly rather than via GetPathfindingSnapshot(), which panics when no
	// snapshot has been taken yet -- possible if the crash happened early in map load.
	crashDumpSection(&snap, "cost map", func() {
		if snapshot := mi.pathfindingSnapshot.Load(); snapshot != nil {
			snap.CostMap = snapshot.CostMap
		}
	})

	// NPCs. npcMu guards the NPC slice, so try for it -- never block on it.
	crashDumpSection(&snap, "npcs", func() {
		if !mi.npcMu.TryLock() {
			snap.skip("npcs: npcMu held by another goroutine")
			return
		}
		defer mi.npcMu.Unlock()
		snap.NPCs = mi.crashDumpNPCs(&snap)
	})

	// The player isn't in the NPC list.
	crashDumpSection(&snap, "player", func() {
		if mi.PlayerRef == nil || mi.PlayerRef.CharacterStateRef == nil {
			return
		}
		p := crashDumpEntity(mi.PlayerRef.Entity, mi.PlayerRef.CharacterStateRef)
		snap.Player = &p
	})

	// Objects. NOTE: mi.Objects has no mutex, and is appended to at runtime (AddObject) and removed from at
	// runtime (door handling), so reading it while the game is running is technically a data race. We accept
	// that here: a torn read panics, and this section is guarded, so the worst case is a dump missing some
	// objects rather than a hang or a masked crash. Doing this properly means locking every write to the
	// slice, which is a larger change than this dump warrants.
	crashDumpSection(&snap, "objects", func() {
		snap.Objects = make([]CrashObjectSnapshot, 0, len(mi.Objects))
		for i := range mi.Objects {
			snap.Objects = append(snap.Objects, crashDumpObject(mi.Objects[i]))
		}
	})

	return snap
}

// skip records a section that couldn't be gathered. Not concurrency-safe on its own, but every caller is on
// the goroutine that built the snapshot, sequentially.
func (snap *CrashStateSnapshot) skip(reason string) {
	snap.Skipped = append(snap.Skipped, reason)
}

// crashDumpSection runs fn, converting any panic into a Skipped entry so one bad section can't take down the
// rest of the dump -- or mask the crash we're trying to report.
func crashDumpSection(snap *CrashStateSnapshot, section string, fn func()) {
	defer func() {
		if r := recover(); r != nil {
			snap.skip(section + ": " + crashDumpErrString(r))
		}
	}()
	fn()
}

// crashDumpNPCs snapshots every NPC. Callers must hold npcMu.
func (mi *ActiveMap) crashDumpNPCs(snap *CrashStateSnapshot) []CrashNPCSnapshot {
	out := make([]CrashNPCSnapshot, 0, len(mi.NPCs))
	for _, n := range mi.NPCs {
		if n == nil || n.Entity == nil || n.CharacterStateRef == nil {
			continue
		}
		s := CrashNPCSnapshot{
			CrashEntitySnapshot: crashDumpEntity(n.Entity, n.CharacterStateRef),
			IsSitting:           n.Entity.IsSitting,
			IsSleeping:          n.Entity.IsSleeping,
			DisableCollisions:   n.Entity.DisableCollisions,
			IsMoving:            n.Entity.Movement.IsMoving,
			MovementInterrupted: n.Entity.Movement.Interrupted,
			PathLen:             len(n.Entity.Movement.TargetPath),
		}
		// TryGetCurrentTaskForDump rather than GetCurrentTaskForBgAssist: that one takes a blocking read
		// lock, and we may be running on a goroutine that already holds (or leaked) it, in which case a
		// blocking read would hang the process instead of reporting the crash.
		if task, readOK := n.TryGetCurrentTaskForDump(); !readOK {
			snap.skip(string(n.CharacterStateRef.ID) + ": task state unavailable (taskStateMu held)")
		} else if task != nil {
			s.TaskID = string(task.GetID())
			s.TaskName = task.GetName()
			s.TaskStatus = task.GetStatus().String()
		}
		out = append(out, s)
	}
	return out
}

// crashDumpEntity snapshots the position and health fields shared by the player and NPCs. cs is required,
// because Entity.ID() and Entity.DisplayName() panic without a backing character state.
func crashDumpEntity(e *entity.Entity, cs *state.CharacterState) CrashEntitySnapshot {
	s := CrashEntitySnapshot{}
	if e == nil || cs == nil {
		return s
	}
	s.ID = string(e.ID())
	s.DisplayName = e.DisplayName()
	s.X = e.X
	s.Y = e.Y
	s.Health = cs.Health
	s.MaxHealth = cs.MaxHealth
	// TilePos depends on map dimensions via the entity's World, which isn't valid before load.
	if e.Loaded {
		s.Tile = e.TilePos()
		s.IsDead = e.IsDead()
	}
	return s
}

// crashDumpObject snapshots one object, including whichever occupancy fields apply to its type.
func crashDumpObject(obj *object.Object) CrashObjectSnapshot {
	if obj == nil {
		return CrashObjectSnapshot{}
	}
	s := CrashObjectSnapshot{
		ID:          obj.ID,
		Type:        obj.Type,
		DisplayName: obj.DisplayName,
		Rect:        obj.GetRect(),
		Collidable:  obj.IsCollidable(false),
		LockID:      obj.GetLockID(),
		OwnerID:     string(obj.OwnerID),
	}
	switch obj.Type {
	case object.TypeChair:
		s.InUse = obj.Chair.InUse
		s.UserID = string(obj.Chair.SitterID)
	case object.TypeBed:
		s.InUse = obj.Bed.InUse
		s.UserID = string(obj.Bed.SleeperID)
	case object.TypeGate:
		s.GateIsOpen = obj.Gate.IsOpen()
	case object.TypeDoor:
		s.DoorTargetMap = obj.Door.TargetMapID
		s.DoorSpawnIdx = obj.Door.TargetSpawnIndex
	}
	return s
}

func crashDumpErrString(r any) string {
	if err, ok := r.(error); ok {
		return err.Error()
	}
	return fmt.Sprint(r)
}
