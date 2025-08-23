package player

import (
	"github.com/hajimehoshi/ebiten/v2"
)

func (p *Player) Update() {
	// capture movement input
	c := p.Entity.TilePos.Copy()
	if ebiten.IsKeyPressed(ebiten.KeyW) {
		c.Y -= 1
		p.Entity.TryMove(c)
	} else if ebiten.IsKeyPressed(ebiten.KeyA) {
		c.X -= 1
		p.Entity.TryMove(c)
	} else if ebiten.IsKeyPressed(ebiten.KeyD) {
		c.X += 1
		p.Entity.TryMove(c)
	} else if ebiten.IsKeyPressed(ebiten.KeyS) {
		c.Y += 1
		p.Entity.TryMove(c)
	}

	p.Entity.Update()
}
