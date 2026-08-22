// Package imgcache is a cache for all image data loaded in the game engine
package imgcache

import (
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/hajimehoshi/ebiten/v2"
)

type Key string

func RawKey(tilesetSrc string, tileID int) Key {
	return Key(fmt.Sprintf("tileset=%s,tileID=%v", tilesetSrc, tileID))
}

func FrameKey(tilesetSrc string, tileID int, flip bool, stretchX, stretchY int, trimRows []int) Key {
	// TODO: should we be using flip in the key? or should consuming code just apply the flip themselves?
	return Key(fmt.Sprintf("tileset=%s,tileID=%v,flip=%v,stretchX=%v,stretchY=%v,trimRows=%v", tilesetSrc, tileID, flip, stretchX, stretchY, trimRows))
}

var (
	cache        = make(map[Key]*ebiten.Image)
	mu           sync.RWMutex
	Hits, Misses atomic.Int64
)

func Get(key Key) (*ebiten.Image, bool) {
	mu.RLock()
	img, ok := cache[key]
	mu.RUnlock()
	if ok {
		Hits.Add(1)
	} else {
		Misses.Add(1)
	}
	return img, ok
}

func Put(key Key, img *ebiten.Image) {
	mu.Lock()
	cache[key] = img
	mu.Unlock()
}

func Clear() {
	mu.Lock()
	cache = make(map[Key]*ebiten.Image)
	mu.Unlock()
}
