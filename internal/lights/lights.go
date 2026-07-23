// Package lights has logic for drawing lights in the map
package lights

import (
	_ "embed"
	"fmt"
	"math/rand"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/webbben/2d-game-engine/config"
	"github.com/webbben/2d-game-engine/data/defs"
	"github.com/webbben/2d-game-engine/display"
	"github.com/webbben/2d-game-engine/logz"
	"github.com/webbben/2d-game-engine/tiled"
)

//go:embed shaders/light.kage
var lightShaderSrc []byte

var (
	lightShader     *ebiten.Shader
	lightShaderInit bool // if this shader has been loaded
)

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

type LightFader struct {
	TargetColor           defs.LightColor
	currentColor          defs.LightColor
	TargetDarknessFactor  float32
	currentDarknessFactor float32
	changeFactor          float32
	changeInterval        time.Duration
	lastChange            time.Time

	overallFactor float32 // this factor influences all light colors; used for eliminating light or increasing its strength
}

func (l LightFader) GetCurrentColor() defs.LightColor {
	return l.currentColor.Scale(l.overallFactor)
}

func (l LightFader) GetDarknessFactor() float32 {
	return l.currentDarknessFactor * l.overallFactor
}

func (l *LightFader) SetOverallFactor(val float32) {
	l.overallFactor = val
}

func NewLightFader(initialColor defs.LightColor, initialDarknessFactor float32, changeFactor float32, changeInterval time.Duration) LightFader {
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

func (lf *LightFader) SetCurrentColor(light defs.LightColor) {
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

type Light struct {
	X, Y                 float32
	MaxRadius, MinRadius float32

	// the "inner radius" is a percentage of the light's radius that is at full brightness. Outside of this radius, brightness starts fading out.
	// this should be a decimal value in the range (0, 1), but typically somewhere around 0.5
	innerRadiusFactor float32

	// the "core radius" is a percentage of the light's radius that is at extra high brightness, and represents the area close to the flame.
	// should not be too big, and must be less than the inner radius (if defined).
	coreRadiusFactor float32

	FlickerTickInterval int
	LightColor          defs.LightColor

	// a value between 0 and 1 which is the percent brightness. lower this value for a dimmer light.
	// defaults to 0.8
	maxBrightness float32

	flickerProgress int
	glowing         bool
	currentRadius   float32
}

func (l Light) String() string {
	return fmt.Sprintf("pos=(%.2f, %.2f) radius=(%v, %v)", l.X, l.Y, l.MinRadius, l.MaxRadius)
}

func NewLight(x, y int, lightDef defs.LightDef) Light {
	if lightDef.InnerRadiusFactor < 0 || lightDef.InnerRadiusFactor >= 1 {
		logz.Panicf("inner radius factor must be positive and < 1. got: %v", lightDef.InnerRadiusFactor)
	}

	if lightDef.FlickerInterval < 50 {
		lightDef.FlickerInterval = 50
	}

	if lightDef.MaxBrightness < 0 {
		panic("tried creating light with negative max brightness")
	}
	if lightDef.MaxBrightness == 0 {
		lightDef.MaxBrightness = 0.8
	}

	if lightDef.CoreRadiusFactor < 0 || lightDef.CoreRadiusFactor >= 1 {
		logz.Panicf("core radius factor must be positive and < 1. got: %v", lightDef.CoreRadiusFactor)
	}
	if lightDef.InnerRadiusFactor > 0 && lightDef.CoreRadiusFactor >= lightDef.InnerRadiusFactor {
		panic("if an inner radius is defined, the core radius must be smaller than it")
	}

	// randomize initial flicker progress so that all lights aren't too synchronized
	flickerProgress := rand.Intn(lightDef.FlickerInterval)

	return Light{
		X:                   float32(x + lightDef.OffsetX),
		Y:                   float32(y + lightDef.OffsetY),
		MinRadius:           float32(lightDef.Radius),
		MaxRadius:           float32(lightDef.Radius) + (float32(lightDef.Radius) * float32(lightDef.GlowFactor)),
		LightColor:          lightDef.Color,
		FlickerTickInterval: lightDef.FlickerInterval,
		flickerProgress:     flickerProgress,
		innerRadiusFactor:   float32(lightDef.InnerRadiusFactor),
		coreRadiusFactor:    float32(lightDef.CoreRadiusFactor),
		maxBrightness:       float32(lightDef.MaxBrightness),
	}
}

func NewLightFromTiledProps(x, y int, lightProp tiled.LightProps) Light {
	var lightColor defs.LightColor

	// otherwise, there must be a light defined in the props
	lightColor[0] = float32(lightProp.R)
	lightColor[1] = float32(lightProp.G)
	lightColor[2] = float32(lightProp.B)

	lightDef := defs.LightDef{
		Radius:            lightProp.Radius,
		GlowFactor:        lightProp.GlowFactor,
		InnerRadiusFactor: lightProp.InnerRadiusFactor,
		CoreRadiusFactor:  lightProp.CoreRadiusFactor,
		FlickerInterval:   lightProp.FlickerInterval,
		MaxBrightness:     lightProp.MaxBrightness,
		Color:             lightColor,
		OffsetY:           lightProp.OffsetY, // TODO: should we add an OffsetX? not sure if we really need it for tiled map objects
	}

	return NewLight(x, y, lightDef)
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

const MaxLights = 16

func DrawMapLighting(screen, scene *ebiten.Image, lights []*Light, daylight defs.LightColor, nightFx float32, offsetX, offsetY float64) {
	if len(lights) > MaxLights {
		logz.Panicln("DrawMapLighting", "number of lights exceeded max light count! max lights:", MaxLights, "num lights:", len(lights))
	}

	lightPositions := make([]float32, MaxLights*2)        // X, Y
	lightRadii := make([]float32, MaxLights)              // radius
	lightInnerRadiusFactors := make([]float32, MaxLights) // inner radius factors
	lightCoreRadiusFactors := make([]float32, MaxLights)  // core radius factors
	lightMaxBrightness := make([]float32, MaxLights)      // max brightness at center of light
	lightColors := make([]float32, MaxLights*3)           // R, G, B

	i := 0
	for range lights {
		lights[i].calculateNextRadius()
		l := lights[i]

		// light position
		lightPositions[i*2] = (l.X - float32(offsetX)) * float32(config.GameScale)
		lightPositions[i*2+1] = (l.Y - float32(offsetY)) * float32(config.GameScale)

		// light radius
		lightRadii[i] = l.currentRadius
		lightInnerRadiusFactors[i] = l.innerRadiusFactor
		lightCoreRadiusFactors[i] = l.coreRadiusFactor

		// brightness
		lightMaxBrightness[i] = l.maxBrightness

		// light color
		lightColors[i*3] = l.LightColor[0]
		lightColors[i*3+1] = l.LightColor[1]
		lightColors[i*3+2] = l.LightColor[2]

		i++
	}

	maxBrightness := 0.5 + (0.5 * (1 - min(1, nightFx)))

	if maxBrightness == 0 {
		panic("max brightness is somehow zero!")
	}

	op := &ebiten.DrawRectShaderOptions{}
	op.Images[0] = scene
	op.Uniforms = map[string]any{
		"LightPositions":          lightPositions,
		"LightRadii":              lightRadii,
		"LightInnerRadiusFactors": lightInnerRadiusFactors,
		"LightCoreRadiusFactors":  lightCoreRadiusFactors,
		"LightMaxBrightness":      lightMaxBrightness,
		"LightColors":             lightColors,
		"NightTint":               daylight,
		"ExtraDarken":             nightFx,
		"MaxBrightness":           maxBrightness,
	}
	screen.DrawRectShader(display.SCREEN_WIDTH, display.SCREEN_HEIGHT, lightShader, op)
}

func CalculateDaylight(hour int) (defs.LightColor, float32) {
	if hour < 0 || hour > 23 {
		panic("invalid hour!")
	}

	switch hour {
	// midnight: 0 - 4
	// dark blue to black
	case 0:
		return defs.LightColor{0.15, 0.2, 0.8}, 1.2
	case 1:
		return defs.LightColor{0.15, 0.2, 0.8}, 1.2
	case 2:
		return defs.LightColor{0.15, 0.15, 0.6}, 1
	case 3:
		return defs.LightColor{0.15, 0.15, 0.4}, 0.9
	case 4:
		return defs.LightColor{0.25, 0.15, 0.4}, 0.8
	// dawn: 5 - 7
	// black to red
	case 5:
		return defs.LightColor{0.35, 0.2, 0.45}, 0.7
	case 6:
		return defs.LightColor{0.55, 0.35, 0.5}, 0.5
	case 7:
		return defs.LightColor{0.7, 0.55, 0.55}, 0.3
	// morning: 8 - 11
	// red to light blue
	case 8:
		return defs.LightColor{0.8, 0.7, 0.7}, 0.1
	case 9:
		return defs.LightColor{0.8, 0.8, 0.9}, 0
	case 10:
		return defs.LightColor{0.8, 0.85, 1}, 0
	case 11:
		return defs.LightColor{0.85, 0.85, 1}, 0
	// midday: 12 - 15
	// light blue to yellow
	case 12:
		return defs.LightColor{0.9, 0.9, 1}, 0
	case 13:
		return defs.LightColor{1.0, 0.9, 0.9}, 0
	case 14:
		return defs.LightColor{1.0, 0.9, 0.8}, 0
	case 15:
		return defs.LightColor{1.0, 0.9, 0.7}, 0
	// evening: 16 - 19
	// yellow to red
	case 16:
		return defs.LightColor{1.0, 0.8, 0.6}, 0
	case 17:
		return defs.LightColor{0.9, 0.75, 0.5}, 0.1
	case 18:
		return defs.LightColor{0.8, 0.5, 0.5}, 0.2
	case 19:
		return defs.LightColor{0.7, 0.4, 0.5}, 0.4
	// night: 20 - 23
	// red to dark blue
	case 20:
		return defs.LightColor{0.4, 0.4, 0.6}, 0.6
	case 21:
		return defs.LightColor{0.3, 0.3, 0.7}, 0.7
	case 22:
		return defs.LightColor{0.2, 0.2, 0.8}, 0.8
	case 23:
		return defs.LightColor{0.15, 0.2, 0.9}, 0.9
	default:
		panic("unknown hour")
	}
}
