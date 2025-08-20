package player

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/webbben/2d-game-engine/internal/model"
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
	} else if ebiten.IsKeyPressed(ebiten.Key0) {
		p.Entity.GoToPos(model.Coords{X: 10, Y: 20})
	}

	p.Entity.Update()
}
