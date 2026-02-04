package cmd

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/webbben/2d-game-engine/clock"
	"github.com/webbben/2d-game-engine/game"
	"github.com/webbben/2d-game-engine/internal/config"
	"github.com/webbben/2d-game-engine/internal/display"
	"github.com/webbben/2d-game-engine/internal/image"
	"github.com/webbben/2d-game-engine/internal/logz"
	"github.com/webbben/2d-game-engine/internal/rendering"
	"github.com/webbben/2d-game-engine/internal/text"
	"github.com/webbben/2d-game-engine/internal/tiled"
	"golang.org/x/image/font"
)

type WorldHUD struct {
	clockTilesetSrc    string
	clockOrigin        int
	clockImg           *ebiten.Image
	minute, hour, year int
	timeString         string
	season             string
	seasonDay          int
	dayOfWeek          string
	clockFont          font.Face
	bodyFont           font.Face

	location string
}

func (hud WorldHUD) ZzCompileCheck() {
	_ = append([]game.HUD{}, &hud)
}

type WorldHUDParams struct {
	ClockTilesetSrc string
	ClockOrigin     int
}

func NewWorldHUD(params WorldHUDParams) WorldHUD {
	hud := WorldHUD{
		clockTilesetSrc: params.ClockTilesetSrc,
		clockOrigin:     params.ClockOrigin,
		timeString:      "No Data!",
	}

	hud.clockFont = image.LoadFont("Romulus.ttf", 34, 0)
	hud.bodyFont = image.LoadFont("Romulus.ttf", 28, 0)

	hud.clockImg = tiled.GetTileImage(hud.clockTilesetSrc, hud.clockOrigin, true)

	hud.location = "Roma, Latium"

	return hud
}

func (hud *WorldHUD) Draw(screen *ebiten.Image) {
	if hud.clockImg == nil {
		logz.Panicln("HUD", "clock image is nil")
	}
	// draw clock in top right of screen
	clockDx := float64(hud.clockImg.Bounds().Dx()) * config.HUDScale
	clockX := display.SCREEN_WIDTH - int(clockDx) - 20
	clockY := 20

	rendering.DrawImage(screen, hud.clockImg, float64(clockX), float64(clockY), config.HUDScale)
	tx := clockX + 90
	ty := clockY + 90
	meridiem := "AM"
	if hud.hour > 11 {
		meridiem = "PM"
	}
	hourVal := hud.hour
	if hourVal == 0 {
		hourVal = 12 // 12 AM instead
	}
	m := fmt.Sprintf("%02d", hud.minute)
	h := fmt.Sprintf("%v", hourVal)
	// draw hour, :, minute, and AM separately so they don't push each other around when numbers of different width appear
	hx := tx
	hDx, _, _ := text.GetStringSize(h, hud.clockFont)
	hx -= hDx
	text.DrawShadowText(screen, h, hud.clockFont, hx, ty, nil, nil, 0, 0)
	cx := tx
	text.DrawShadowText(screen, ":", hud.clockFont, cx, ty, nil, nil, 0, 0)
	mx := cx + 8
	text.DrawShadowText(screen, m, hud.clockFont, mx, ty, nil, nil, 0, 0)

	text.DrawShadowText(screen, meridiem, hud.clockFont, mx+50, ty, nil, nil, 0, 0)

	dayString := fmt.Sprintf("%s, %s %v", hud.dayOfWeek, hud.season, hud.seasonDay+1)
	ty += 25
	tx = clockX + 45
	text.DrawShadowText(screen, dayString, hud.bodyFont, tx, ty, nil, nil, 0, 0)

	ty += 25
	text.DrawShadowText(screen, hud.location, hud.bodyFont, tx, ty, nil, nil, 0, 0)

	ty += 40

	text.DrawShadowText(screen, fmt.Sprintf("Anno %v", hud.year), hud.bodyFont, tx+70, ty, color.RGBA{0, 0, 0, 100}, nil, 0, 0)
}

func (hud *WorldHUD) Update(g *game.Game) {
	m, h, y, season, seasonDay, dow := g.Clock.GetCurrentDateAndTime()
	hud.minute = m
	hud.hour = h
	hud.year = y
	hud.season = string(clock.Seasons[season])
	hud.seasonDay = seasonDay
	hud.dayOfWeek = string(dow)
	hud.timeString = g.Clock.GetTimeString(true)
}
