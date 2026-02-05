package cmd

import (
	"fmt"
	"image/color"
	"time"

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
	clockImgDay, clockImgEvening     *ebiten.Image
	clockImgNight, clockImgLateNight *ebiten.Image
	minute, hour, year               int
	timeString                       string
	season                           string
	seasonDay                        int
	dayOfWeek                        string
	clockFont                        font.Face
	bodyFont                         font.Face

	location        string
	playerWasActive bool // if, last time's update, the player was active
	playerActive    bool // if, this update, the player was found to be active

	positionFader    rendering.FadeToPosition
	hiddenX, hiddenY float64 // position to go to when not showing (off screen)
	showX, showY     float64 // position to go to when showing
}

func (hud WorldHUD) getClockImage() *ebiten.Image {
	if hud.hour >= 0 && hud.hour < 5 {
		return hud.clockImgLateNight
	}
	if hud.hour >= 5 && hud.hour < 9 {
		return hud.clockImgEvening
	}
	if hud.hour >= 9 && hud.hour < 17 {
		return hud.clockImgDay
	}
	if hud.hour >= 17 && hud.hour < 20 {
		return hud.clockImgEvening
	}
	return hud.clockImgNight
}

func (hud WorldHUD) ZzCompileCheck() {
	_ = append([]game.HUD{}, &hud)
}

type WorldHUDParams struct {
	ClockTilesetSrc                                                        string
	ClockDayIndex, ClockEveningIndex, ClockNightIndex, ClockLateNightIndex int
}

func NewWorldHUD(params WorldHUDParams) WorldHUD {
	hud := WorldHUD{
		timeString: "No Data!",
	}

	hud.clockFont = image.LoadFont("Romulus.ttf", 34, 0)
	hud.bodyFont = image.LoadFont("Romulus.ttf", 28, 0)

	hud.clockImgDay = tiled.GetTileImage(params.ClockTilesetSrc, params.ClockDayIndex, true)
	hud.clockImgEvening = tiled.GetTileImage(params.ClockTilesetSrc, params.ClockEveningIndex, true)
	hud.clockImgNight = tiled.GetTileImage(params.ClockTilesetSrc, params.ClockNightIndex, true)
	hud.clockImgLateNight = tiled.GetTileImage(params.ClockTilesetSrc, params.ClockLateNightIndex, true)

	hud.location = "Roma, Latium"

	clockDx := float64(hud.clockImgDay.Bounds().Dx()) * config.HUDScale
	clockDy := float64(hud.clockImgDay.Bounds().Dy()) * config.HUDScale
	clockX := display.SCREEN_WIDTH - int(clockDx) - 20
	clockY := 20
	hud.hiddenX = float64(clockX)
	hud.hiddenY = -clockDy - 20
	hud.showX = float64(clockX)
	hud.showY = float64(clockY)

	return hud
}

func (hud *WorldHUD) startMovementToPos(fromX, fromY, toX, toY float64) {
	speed := 0.12
	hud.positionFader = rendering.NewFadeToPosition(toX, toY, fromX, fromY, float64(speed))
}

func (hud *WorldHUD) Draw(screen *ebiten.Image) {
	clockImg := hud.getClockImage()
	if clockImg == nil {
		logz.Panicln("HUD", "clock image is nil")
	}
	// draw clock in top right of screen
	clockX := int(hud.positionFader.Current.X)
	clockY := int(hud.positionFader.Current.Y)

	rendering.DrawImage(screen, clockImg, float64(clockX), float64(clockY), config.HUDScale)
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
	hud.positionFader.Update()

	m, h, y, season, seasonDay, dow := g.Clock.GetCurrentDateAndTime()
	hud.minute = m
	hud.hour = h
	hud.year = y
	hud.season = string(clock.Seasons[season])
	hud.seasonDay = seasonDay
	hud.dayOfWeek = string(dow)
	hud.timeString = g.Clock.GetTimeString(true)

	hud.playerWasActive = hud.playerActive
	hud.playerActive = time.Since(g.LastPlayerUpdate()) > (time.Second * 2)

	// check if clock should move
	if hud.playerWasActive != hud.playerActive {
		if hud.playerActive {
			// move into view
			hud.startMovementToPos(hud.hiddenX, hud.hiddenY, hud.showX, hud.showY)
		} else {
			hud.startMovementToPos(hud.showX, hud.showY, hud.hiddenX, hud.hiddenY)
		}
	}
}
