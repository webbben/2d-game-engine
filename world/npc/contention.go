package npc

import (
	"time"

	"github.com/webbben/2d-game-engine/config"
	"github.com/webbben/2d-game-engine/logz"
	"github.com/webbben/2d-game-engine/data/model"
)

// Waypoint contention.
//
// When two or more NPCs pick the same next tile, neither can resolve the standoff alone. This file
// detects that and, since the standoff stalls both NPCs until one proceeds, decides who goes first.
// Acting on the verdict (stop moving, hold, resume) lives in waypointYielder below, so any path-walking
// task can use identical semantics by embedding one and calling maybeYield at the top of its update.

// contentionYieldCooldown is how long the losing claimant holds still after losing a standoff before
// the claim is re-evaluated. Roughly one movement tick of the winner covers the contested tile, so a
// second is enough for the winner to pass; keeping the cooldown near the path-request throttle means
// the loser's first re-request after yielding reads a cost map that already reflects the winner's new
// position.
const contentionYieldCooldown = time.Second

// contentionRepeatYieldsBeforeBlock is how many consecutive yield episodes against the same tile it
// takes before the standoff is treated as wedged rather than merely contested. One deterministic
// yield-and-wait should clear any pair standoff; losing the same tile a second time in a row means
// something is stuck in it (a body that won't leave, or a pile where each new winner stalls in turn).
const contentionRepeatYieldsBeforeBlock = 1

// contentionMaxHoldExtension caps the doubled hold used by tasks whose escalation extends the hold
// rather than re-routing (GotoTask, which has no blocked-tile path support). An eight second cap is
// still short enough that a crowd doesn't feel frozen, and long enough that anything genuinely
// blocking the tile has usually cleared by then.
const contentionMaxHoldExtension = time.Second * 8

// contendedWaypointYield reports whether this npc loses the standoff over its committed waypoint,
// along with the contested tile when it does (meaningless when it doesn't).
//
// Both sides of the standoff evaluate this over the same claimant set, so the verdict is symmetric.
// The rule: the claimant whose entity position is strictly closer (Euclidean, in pixels, to the
// contested tile's center) wins the tile. Pixel distance rather than tile distance matters because
// most standoffs end up with every claimant in an adjacent tile -- tile distances would tie them all
// and leave the ID fallback deciding the common case. Pixel distance also naturally favors the NPC
// that has already partially stepped into the tile, which is the behavior a crowd wants: whoever is
// already committing keeps going. Exact pixel ties (rare, e.g. mirrored approaches) fall back to the
// lowest entity string ID so the verdict is always deterministic.
//
// The losing side acts on the verdict synchronously (see waypointYielder.beginEpisode): it stops
// moving, which drops it out of the claimant set immediately (contestedClaim counts only moving
// NPCs), so the winner walks through and the standoff self-resolves. Because the loser stops the
// same tick it loses, there is no window for both sides to re-derive opposite verdicts as positions
// drift -- the comparison effectively happens once per standoff.
func (n *NPC) contendedWaypointYield() (model.Coords, bool) {
	tile, claimants := n.contestedClaim()
	if len(claimants) == 0 {
		return model.Coords{}, false
	}
	tileCenter := model.Vec2{
		X: float64(tile.X*config.TileSize) + float64(config.TileSize)/2,
		Y: float64(tile.Y*config.TileSize) + float64(config.TileSize)/2,
	}
	mine := model.Vec2{X: n.Entity.X, Y: n.Entity.Y}.Dist(tileCenter)
	for _, c := range claimants {
		theirs := model.Vec2{X: c.Entity.X, Y: c.Entity.Y}.Dist(tileCenter)
		if theirs < mine {
			return tile, true
		}
		if theirs == mine && string(n.Entity.ID()) > string(c.Entity.ID()) {
			return tile, true
		}
	}
	return tile, false
}

// waypointYielder carries the hold/resume/escalation state for one task's contention handling.
// Tasks embed it and call maybeYield at the top of their update; everything else here is bookkeeping
// for that call. Written only by the main game loop.
type waypointYielder struct {
	// holdUntil is when the current hold (if any) expires. holding reports whether a hold is active.
	holdUntil time.Time
	holding   bool

	// lastContestedTile and repeatEpisodes count consecutive yield episodes against the same tile,
	// which is what escalation keys off. Episodes, not frames: a hold keeps the task's path in hand,
	// so the claim stays live and a naive per-tick check would count every frame of the hold.
	lastContestedTile model.Coords
	repeatEpisodes    int

	// lastYieldTile is the entity's tile when the first episode began; if the entity changes tiles
	// before the streak continues, progress was made and the streak is re-armed. This is the
	// re-arm-on-progress mechanism, keyed off the entity itself so every task gets it without each
	// task needing its own progress hook.
	lastYieldTile model.Coords

	// blockTile is armed when a streak hits the repeat threshold and this yielder's escalation is
	// re-routing: tasks with blocked-tile path support (FollowTask) consume it via takeBlock and put
	// the tile on their next path request as a wall.
	blockTile      model.Coords
	blockTileArmed bool

	// extendHoldOnEscalate selects the escalation for tasks without blocked-tile path support
	// (GotoTask routes synchronously via GoToPos): instead of arming a wall, each escalation doubles
	// the hold, capped at contentionMaxHoldExtension.
	extendHoldOnEscalate bool

	// lastYieldLog rate-limits the yield report so a persistent standoff doesn't bury everything else.
	lastYieldLog time.Time
}

// yieldVerdict tells the calling task how to proceed this tick.
type yieldVerdict int

const (
	// yieldProceed means there is no standoff to handle; run the normal update.
	yieldProceed yieldVerdict = iota
	// yieldHold means this NPC lost the standoff (or is still holding from a previous loss): skip the
	// normal update and stand still until the verdict changes.
	yieldHold
	// yieldResume means a hold has just expired with the contested tile now clear: call
	// Entity.ResumePath (the task keeps its path through a hold for exactly this reason), then skip
	// the rest of the update -- normal walking resumes next tick.
	yieldResume
)

// maybeYield consults the contention verdict on the calling task's behalf. Call it at the top of the
// task's update and act on the returned verdict.
//
// While a hold is running and not yet expired the verdict is simply yieldHold; the standoff is only
// re-derived once the hold expires (or when no hold is active). The path is never cancelled here:
// resuming along the existing route is what the yieldResume verdict is for.
func (y *waypointYielder) maybeYield(n *NPC) yieldVerdict {
	if y.holding && time.Now().Before(y.holdUntil) {
		return yieldHold
	}
	tile, lose := n.contendedWaypointYield()
	if !lose {
		if y.holding {
			// the wait resolved the standoff, so this streak is over: clear it so a later
			// standoff over the same tile starts counting from scratch instead of escalating.
			y.holding = false
			y.resetStreak()
			return yieldResume
		}
		return yieldProceed
	}
	y.beginEpisode(n, tile)
	if y.holding {
		logz.Println("waypointYielder", "started yielding.", "contested tile:", tile.String(), n.WhoAmI())
		return yieldHold
	}
	// if we get here, then that means we've hit the max repeated episodes limit and
	// decided to arm the tile blocker
	return yieldProceed
}

// beginEpisode starts (or extends) a hold for a lost standoff. The route is kept through a normal
// hold so resuming along it is cheap and non-disruptive, and a re-check on hold expiry covers the
// case where the winner is slower than the cooldown. The one exception is escalation, which drops
// the route (see below) because the whole point is to not walk it again.
func (y *waypointYielder) beginEpisode(n *NPC, tile model.Coords) {
	n.Entity.StopMovement()
	y.holding = true
	y.holdUntil = time.Now().Add(contentionYieldCooldown)

	if tile.Equals(y.lastContestedTile) {
		y.repeatEpisodes++
	} else {
		// different contested tile since last time
		y.lastContestedTile = tile
		y.repeatEpisodes = 0
		y.lastYieldTile = n.Entity.TilePos()
	}

	// re-arm the streak if the entity has made real progress (changed tiles) since the streak
	// began; a resolved standoff followed by a fresh one elsewhere isn't the same wedged tile.
	if !y.lastYieldTile.Equals(n.Entity.TilePos()) {
		y.lastYieldTile = n.Entity.TilePos()
		y.repeatEpisodes = 0
	}

	if y.repeatEpisodes >= contentionRepeatYieldsBeforeBlock {
		if y.extendHoldOnEscalate {
			// each escalation doubles the hold, so 2s, 4s, then the cap.
			extension := time.Duration(1<<((y.repeatEpisodes-contentionRepeatYieldsBeforeBlock)+1)) * contentionYieldCooldown
			extension = min(extension, contentionMaxHoldExtension)
			y.holdUntil = time.Now().Add(extension)
			logz.Warnln("waypointYielder", "too many repeated yields; escalating to longer yield:", extension.String())
		} else {
			// armed wall for the caller's next path request. The route is dropped and the hold
			// released right here: resuming along the old route would walk straight back into the
			// wedged tile, and the caller's normal flow re-requests with takeBlock consumed.
			y.blockTile = tile
			y.blockTileArmed = true
			if n.Entity.HasPath() {
				n.Entity.CancelCurrentPath()
			}
			y.holding = false
			y.holdUntil = time.Time{}
			logz.Warnln("waypointYielder", "too many repeated yields; escalating to blocking contested tile and rerouting:", tile.String())
		}
	}

	if time.Since(y.lastYieldLog) > time.Second*1 {
		y.lastYieldLog = time.Now()
		logz.Println("yielding contended waypoint", tile, n.WhoAmI())
	}
}

// takeBlock returns the tile a wedged standoff has escalated to blocking, consuming the escalation.
// Tasks without blocked-tile path support never arm it (they extend the hold instead).
func (y *waypointYielder) takeBlock() (model.Coords, bool) {
	if !y.blockTileArmed {
		return model.Coords{}, false
	}
	y.blockTileArmed = false
	return y.blockTile, true
}

// resetStreak clears the repeat-episode tracking. Called when a hold resolves the standoff, so a
// standoff that did not survive a wait is never treated as a repeat of the one before it.
func (y *waypointYielder) resetStreak() {
	y.lastContestedTile = model.Coords{}
	y.repeatEpisodes = 0
	y.lastYieldTile = model.Coords{}
}

// contestedClaim returns the tile this npc is committed to entering next, along with every other live
// NPC on the map that is walking into that same tile. An empty claimant list means there is no
// contention, and the returned tile is then meaningless.
//
// The reason this happens: A* charges only EntityCollision (2) for an occupied tile, which is no more
// than a two-tile detour costs. So when two NPCs want the same waypoint, both paths are individually
// optimal and identically shaped, and the search has no reason to prefer either. They converge, push
// each other apart, and that displacement can knock one off its own route -- its next waypoint then
// sits further away than the dist<=1 salvage range in trySetNextTargetPath, the path is cleared, and
// the NPC re-routes to the same waypoint. That cycle repeats indefinitely.
//
// This is deliberately a property of the NPCs rather than of any one task. "These two are walking
// into the same tile" is true regardless of what either is doing, and an NPC can be committed to a
// waypoint without a task being able to see it. Tasks ask the questions; this answers them for anyone
// who asks, and stays available to whichever task wants to act on it.
//
// Only NPCs are claimants. The player is deliberately excluded, so this answers "is another NPC
// competing for this tile" and not the separate question of player occupancy, which the cost map
// already handles.
func (n *NPC) contestedClaim() (model.Coords, []*NPC) {
	var tile model.Coords
	if next, ok := n.Entity.PathAhead(0); ok {
		tile = next
	} else if n.Entity.IsMoving() {
		tile = n.Entity.TargetTilePos()
	} else {
		// Standing still with no path: not committed to entering anything.
		return model.Coords{}, nil
	}

	var claimants []*NPC
	for _, other := range n.ActiveMapCtx.GetAllNPCs() {
		if other.ID() == n.ID() {
			continue
		}
		// A body lying on the map is still a map occupant but is going nowhere, and an NPC that
		// isn't moving has already arrived wherever it was headed. Note the moving check matters more
		// than it looks: an NPC in a standoff oscillates between moving and stopping as it gets
		// displaced and re-paths, and without it the check would blink out during exactly the frames
		// worth looking at.
		if other.Entity.IsDead() || !other.Entity.IsMoving() {
			continue
		}
		if other.Entity.TargetTilePos().Equals(tile) {
			claimants = append(claimants, other)
		}
	}
	if len(claimants) == 0 {
		return model.Coords{}, nil
	}
	return tile, claimants
}

// logContestedClaim reports detected contention, rate-limited so a standoff that persists doesn't bury
// every other message. The total number of NPCs is reported rather than just "yes/no", because three
// or more competing for one tile needs different handling than a pair, and it's cheaper to find out
// whether that actually happens than to assume it.
func (n *NPC) logContestedClaim() {
	tile, claimants := n.contestedClaim()
	if len(claimants) == 0 {
		return
	}
	if time.Since(n.lastContestedLog) < time.Second*5 {
		return
	}
	n.lastContestedLog = time.Now()

	who := make([]string, 0, len(claimants)+1)
	who = append(who, n.WhoAmI())
	for _, c := range claimants {
		who = append(who, c.WhoAmI())
	}
	logz.Println("contestedClaim:", len(claimants)+1, "NPCs are all walking into tile", tile, who)
}
