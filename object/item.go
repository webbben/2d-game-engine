package object

import (
	"fmt"
	"math"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/webbben/2d-game-engine/audio"
	"github.com/webbben/2d-game-engine/data/datamanager"
	"github.com/webbben/2d-game-engine/data/defs"
	"github.com/webbben/2d-game-engine/data/state"
	"github.com/webbben/2d-game-engine/logz"
	"github.com/webbben/2d-game-engine/pubsub"
	"github.com/webbben/2d-game-engine/tiled"
	"github.com/webbben/2d-game-engine/tiled/properties"
	"github.com/webbben/2d-game-engine/utils"
)

const (
	// how long one full up-and-down cycle of a dropped item takes, in real seconds.
	bobPeriod = 1.5

	// the peak vertical bob amplitude for a dropped item, in world pixels
	defaultBobAmplitude = 3.0
)

type Item struct {
	ItemID defs.ItemID

	// Dropped mirrors MapItemState for a runtime drop. The map state is the source of truth.
	Dropped bool

	drawYOffset    float64 // vertical offset in world pixels of the bobbing dropped item.
	drawYOffsetAmp float64 // peak amplitude of the bob of a the dropped item. 0 means "do not bob".

	bobPhase float64 // accumulated real-time seconds driving the sine.

	bobLastUpdate time.Time
}

func (o *Object) loadItemObject(allProps []properties.Property) {
	itemID, found := properties.GetStringProperty(properties.PropItemID, allProps)
	if !found {
		logz.Panic("item didn't have item_id")
	}
	itemDef := o.dataman.GetItemDef(defs.ItemID(itemID))
	o.DisplayName = itemDef.Name
	o.Item.ItemID = itemDef.ID

	// don't allow items to load in without an embedded tile image; otherwise, we could have invisible items hanging around.
	if len(o.imgFrames) == 0 {
		logz.PanicCtx("loadItemObject", "item object has no embedded tile.", o.Name, o.ID, itemID)
	}
}

func NewDroppedItem(
	st state.MapItemState,
	dataman *datamanager.DataManager,
	audioman *audio.AudioManager,
	eventBus *pubsub.EventBus,
	mapID defs.MapID,
	worldCtx WorldContext,
) *Object {
	utils.PanicAssert(st.Dropped, "item state was not dropped")

	itemDef := dataman.GetItemDef(st.ItemState.DefID)
	img := tiled.GetTileImage(itemDef.TileImgTilesetSrc, itemDef.TileImgIndex, true)

	o := newObjectScaffolding(
		st.ID,
		fmt.Sprintf("dropped_%s_%v", itemDef.ID, st.ID),
		eventBus,
		audioman,
		dataman,
		worldCtx,
		mapID,
		0,
		0,
	)

	o.SetImageFrames([]*ebiten.Image{img})
	// we don't pass the flag for embeddedTileOrigin since this object isn't originated from a Tiled file, the issue shouldn't exist here
	o.SetPosition(st.X, st.Y, false)

	// fill in item data
	o.Type = TypeItem
	o.DisplayName = itemDef.Name
	o.Item = Item{
		ItemID:         st.ItemState.DefID,
		Dropped:        true,
		drawYOffsetAmp: defaultBobAmplitude,
	}

	// TODO: should we be getting this data directly from a tileset or something?
	// looks a bit off to be filling it in manually, and would feel a bit more secure if we knew the data we are setting is
	// consistent with how it is done elsewhere. Although, since this is a dropped item, I suppose tileData may not be the same
	// as when it's part of an authored item object. Deserves a little investigation I think.
	o.tileData = tiled.TileData{CurrentFrame: img}

	// no zOffset since a dropped tile will land on the player's own tile, and it's okay for it to render behind the player at that point.

	// no event subscriptions either, so OnMapClose is a no-op for dropped items.

	o.Validate()
	return o
}

func bobOffset(phase, amplitude float64) float64 {
	return amplitude * math.Sin(phase*2*math.Pi/bobPeriod)
}

func (o *Object) activateItem() ObjectUpdateResult {
	utils.PanicAssert(o.Type == TypeItem, "object isn't an item")

	return ObjectUpdateResult{
		UpdateOccurred:    true,
		PickedUpItemObjID: o.ID,
	}
}
