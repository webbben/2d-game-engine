package lights

import (
	_ "embed"
	"fmt"
	"image/color"
	"math/rand"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/webbben/2d-game-engine/internal/config"
	"github.com/webbben/2d-game-engine/internal/display"
	"github.com/webbben/2d-game-engine/internal/logz"
	"github.com/webbben/2d-game-engine/internal/tiled"
)

//go:embed shaders/light.kage
var lightShaderSrc []byte

var lightShader *ebiten.Shader
var lightShaderInit bool // if this shader has been loaded

func LoadShaders() error {
	fmt.Println("Loading shaders...")
	var err error
	lightShader, err = ebiten.NewShader(lightShaderSrc)
	if err != nil {
		return err
	}
	lightShaderInit = true
	fmt.Println("All shaders successfully loaded.")
	return nil
}

func LightShader() *ebiten.Shader {
	if !lightShaderInit {
		panic("tried to use light shader before it was successfully loaded! Did an error occur when loading shaders?")
	}
	return lightShader
}

// instead of using [0, 255] RGB values, we use [0, 1] values.
// mainly because that's what's used in the shader, but also easier to conceptualize as percentages.
type LightColor [3]float32

func (l LightColor) Equals(lc LightColor) bool {
	return l[0] == lc[0] && l[1] == lc[1] && l[2] == lc[2]
}

func (l LightColor) Scale(factor float32) LightColor {
	return LightColor{l[0] * factor, l[1] * factor, l[2] * factor}
}

type LightFader struct {
	TargetColor           LightColor
	currentColor          LightColor
	TargetDarknessFactor  float32
	currentDarknessFactor float32
	changeFactor          float32
	changeInterval        time.Duration
	lastChange            time.Time

	overallFactor float32 // this factor influences all light colors; used for eliminating light or increasing its strength
}

func (l LightFader) GetCurrentColor() LightColor {
	return l.currentColor.Scale(l.overallFactor)
}

func (l LightFader) GetDarknessFactor() float32 {
	return l.currentDarknessFactor * l.overallFactor
}

func (l *LightFader) SetOverallFactor(val float32) {
	l.overallFactor = val
}

func NewLightFader(initialColor LightColor, initialDarknessFactor float32, changeFactor float32, changeInterval time.Duration) LightFader {
	if changeFactor <= 0 {
		changeFactor = 0.1
	}
	if changeInterval == 0 {
		changeInterval = time.Second
	}

	lf := LightFader{
		currentColor:          initialColor,
		TargetColor:           initialColor,
		changeFactor:          changeFactor,
		lastChange:            time.Now(),
		changeInterval:        changeInterval,
		currentDarknessFactor: initialDarknessFactor,
		TargetDarknessFactor:  initialDarknessFactor,
		overallFactor:         1,
	}

	return lf
}

func (lf *LightFader) SetCurrentColor(light LightColor) {
	lf.currentColor = light
	lf.lastChange = time.Now()
}

func (lf *LightFader) SetCurrentDarknessFactor(factor float32) {
	lf.currentDarknessFactor = factor
	lf.lastChange = time.Now()
}

func (l *LightFader) Update() {
	if l.currentColor.Equals(l.TargetColor) {
		return
	}
	if time.Since(l.lastChange) < l.changeInterval {
		return
	}
	l.lastChange = time.Now()

	l.currentColor[0] += (l.TargetColor[0] - l.currentColor[0]) * l.changeFactor
	l.currentColor[1] += (l.TargetColor[1] - l.currentColor[1]) * l.changeFactor
	l.currentColor[2] += (l.TargetColor[2] - l.currentColor[2]) * l.changeFactor

	l.currentDarknessFactor += (l.TargetDarknessFactor - l.currentDarknessFactor) * l.changeFactor
}

var (
	// light colors (for cutting through darkness as a light source)

	LIGHT_TORCH  = LightColor{1.0, 0.8, 0.8}
	LIGHT_CANDLE = color.RGBA{255, 220, 180, 0}
	LIGHT_FIRE   = color.RGBA{255, 240, 200, 0}

	// darkness colors (for the general overlay)

	DARK_NIGHTSKY = LightColor{0.15, 0.25, 1.0}
	DARK_CAVE     = color.RGBA{30, 30, 30, 0}
	DARK_MAGICAL  = color.RGBA{20, 0, 40, 0}
	DARK_DUSK     = color.RGBA{40, 30, 60, 0}
)

type Light struct {
	X, Y                 float32
	MaxRadius, MinRadius float32

	// the "inner radius" is a percentage of the light's radius that is at full brightness. Outside of this radius, brightness starts fading out.
	// this should be a decimal value in the range (0, 1), but typically somewhere around 0.5
	innerRadiusFactor   float32
	FlickerTickInterval int
	LightColor          LightColor

	// a value between 0 and 1 which is the percent brightness. lower this value for a dimmer light.
	// defaults to 0.8
	maxBrightness float32

	flickerProgress int
	glowing         bool
	currentRadius   float32
}

func NewLight(x, y int, lightProp tiled.LightProps, customLight *LightColor) Light {
	lightColor := LIGHT_TORCH
	if lightProp.ColorPreset == "torch" {
		// TODO setup some color presets
		lightColor = LIGHT_TORCH
	}
	// if a custom light is defined, use it
	if customLight != nil {
		lightColor = *customLight
	}

	if lightProp.InnerRadiusFactor < 0 || lightProp.InnerRadiusFactor >= 1 {
		logz.Panicf("inner radius factor must be positive and < 1. got: %v", lightProp.InnerRadiusFactor)
	}
	// if lightProp.InnerRadiusFactor == 0 {
	// 	lightProp.InnerRadiusFactor = 0.5
	// }

	if lightProp.FlickerInterval < 50 {
		lightProp.FlickerInterval = 50
	}

	if lightProp.MaxBrightness < 0 {
		panic("tried creating light with negative max brightness")
	}
	if lightProp.MaxBrightness == 0 {
		lightProp.MaxBrightness = 0.8
	}
	fmt.Println("max brightness:", lightProp.MaxBrightness)

	// randomize initial flicker progress so that all lights aren't too synchronized
	flickerProgress := rand.Intn(lightProp.FlickerInterval)

	return Light{
		X:                   float32(x),
		Y:                   float32(y + lightProp.OffsetY),
		MinRadius:           float32(lightProp.Radius),
		MaxRadius:           float32(lightProp.Radius) + (float32(lightProp.Radius) * float32(lightProp.GlowFactor)),
		LightColor:          lightColor,
		FlickerTickInterval: lightProp.FlickerInterval,
		flickerProgress:     flickerProgress,
		innerRadiusFactor:   float32(lightProp.InnerRadiusFactor),
		maxBrightness:       float32(lightProp.MaxBrightness),
	}
}

func (l *Light) calculateNextRadius() {
	if l.glowing {
		l.flickerProgress++
		if l.flickerProgress >= l.FlickerTickInterval {
			l.glowing = false
		}
	} else {
		l.flickerProgress--
		if l.flickerProgress <= 0 {
			l.glowing = true
		}
	}

	flickerPercent := float64(l.flickerProgress) / float64(l.FlickerTickInterval)

	maxRadius := l.MaxRadius * float32(config.GameScale)
	minRadius := l.MinRadius * float32(config.GameScale)
	l.currentRadius = ((maxRadius - minRadius) * float32(flickerPercent)) + minRadius
}

func DrawMapLighting(screen, scene *ebiten.Image, lights []*Light, objLights []*Light, daylight LightColor, nightFx float32, offsetX, offsetY float64) {
	maxLights := 16

	lightPositions := make([]float32, maxLights*2)        // X, Y
	lightRadii := make([]float32, maxLights)              // radius
	lightInnerRadiusFactors := make([]float32, maxLights) // inner radius factors
	lightMaxBrightness := make([]float32, maxLights)      // max brightness at center of light
	lightColors := make([]float32, maxLights*3)           // R, G, B

	for i := range lights {
		lights[i].calculateNextRadius()
		l := lights[i]

		// light position
		lightPositions[i*2] = (l.X - float32(offsetX)) * float32(config.GameScale)
		lightPositions[i*2+1] = (l.Y - float32(offsetY)) * float32(config.GameScale)

		// light radius
		lightRadii[i] = l.currentRadius
		lightInnerRadiusFactors[i] = l.innerRadiusFactor

		// light color
		lightColors[i*3] = l.LightColor[0]
		lightColors[i*3+1] = l.LightColor[1]
		lightColors[i*3+2] = l.LightColor[2]
	}
	for i := range objLights {
		objLights[i].calculateNextRadius()
		l := objLights[i]

		// light position
		lightPositions[i*2] = (l.X - float32(offsetX)) * float32(config.GameScale)
		lightPositions[i*2+1] = (l.Y - float32(offsetY)) * float32(config.GameScale)

		// light radius
		lightRadii[i] = l.currentRadius
		lightInnerRadiusFactors[i] = l.innerRadiusFactor

		// brightness
		lightMaxBrightness[i] = l.maxBrightness

		// light color
		lightColors[i*3] = l.LightColor[0]
		lightColors[i*3+1] = l.LightColor[1]
		lightColors[i*3+2] = l.LightColor[2]
	}

	op := &ebiten.DrawRectShaderOptions{}
	op.Images[0] = scene
	op.Uniforms = map[string]interface{}{
		"LightPositions":          lightPositions,
		"LightRadii":              lightRadii,
		"LightInnerRadiusFactors": lightInnerRadiusFactors,
		"LightMaxBrightness":      lightMaxBrightness,
		"LightColors":             lightColors,
		"NightTint":               daylight,
		"ExtraDarken":             nightFx,
	}
	screen.DrawRectShader(display.SCREEN_WIDTH, display.SCREEN_HEIGHT, lightShader, op)
}

func CalculateDaylight(hour int) (LightColor, float32) {
	if hour < 0 || hour > 23 {
		panic("invalid hour!")
	}

	switch hour {
	// midnight: 0 - 4
	// dark blue to black
	case 0:
		return LightColor{0.15, 0.2, 0.8}, 1.2
	case 1:
		return LightColor{0.15, 0.2, 0.8}, 1.2
	case 2:
		return LightColor{0.15, 0.15, 0.6}, 1
	case 3:
		return LightColor{0.15, 0.15, 0.4}, 0.9
	case 4:
		return LightColor{0.25, 0.15, 0.4}, 0.8
	// dawn: 5 - 7
	// black to red
	case 5:
		return LightColor{0.35, 0.2, 0.45}, 0.7
	case 6:
		return LightColor{0.55, 0.35, 0.5}, 0.5
	case 7:
		return LightColor{0.7, 0.55, 0.55}, 0.3
	// morning: 8 - 11
	// red to light blue
	case 8:
		return LightColor{0.8, 0.7, 0.7}, 0.1
	case 9:
		return LightColor{0.8, 0.8, 0.9}, 0
	case 10:
		return LightColor{0.8, 0.85, 1}, 0
	case 11:
		return LightColor{0.85, 0.85, 1}, 0
	// midday: 12 - 15
	// light blue to yellow
	case 12:
		return LightColor{0.9, 0.9, 1}, 0
	case 13:
		return LightColor{1.0, 0.9, 0.9}, 0
	case 14:
		return LightColor{1.0, 0.9, 0.8}, 0
	case 15:
		return LightColor{1.0, 0.9, 0.7}, 0
	// evening: 16 - 19
	// yellow to red
	case 16:
		return LightColor{1.0, 0.8, 0.6}, 0
	case 17:
		return LightColor{0.9, 0.75, 0.5}, 0.1
	case 18:
		return LightColor{0.8, 0.5, 0.5}, 0.2
	case 19:
		return LightColor{0.7, 0.4, 0.5}, 0.4
	// night: 20 - 23
	// red to dark blue
	case 20:
		return LightColor{0.4, 0.4, 0.6}, 0.6
	case 21:
		return LightColor{0.3, 0.3, 0.7}, 0.7
	case 22:
		return LightColor{0.2, 0.2, 0.8}, 0.8
	case 23:
		return LightColor{0.15, 0.2, 0.9}, 0.9
	default:
		panic("unknown hour")
	}
}

func shaderColorScale(c color.Color) LightColor {
	r, g, b, _ := c.RGBA()
	return LightColor{float32(r) / 0xffff, float32(g) / 0xffff, float32(b) / 0xffff}
}
