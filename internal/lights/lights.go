package lights

import (
	_ "embed"
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/webbben/2d-game-engine/internal/display"
)

//go:embed shaders/light.kage
var lightShaderSrc []byte

var lightShader *ebiten.Shader
var lightShaderInit bool // if this shader has been loaded

func LoadShaders() error {
	var err error
	lightShader, err = ebiten.NewShader(lightShaderSrc)
	if err != nil {
		return err
	}
	lightShaderInit = true
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
	FlickerTickInterval  int
	LightColor           LightColor

	flickerProgress int
	glowing         bool
	currentRadius   float32
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

	l.currentRadius = ((l.MaxRadius - l.MinRadius) * float32(flickerPercent)) + l.MinRadius
}

func RadialGradient(radius int, lightColor color.Color) *ebiten.Image {
	// light color defaults to white
	var r, g, b uint8 = 0, 0, 0
	if lightColor != nil {
		r1, g1, b1, _ := lightColor.RGBA()
		r = uint8(r1)
		g = uint8(g1)
		b = uint8(b1)
	}
	size := radius * 2
	img := ebiten.NewImage(size, size)

	center := float64(radius)
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			dx := float64(x) - center
			dy := float64(y) - center
			dist := math.Sqrt((dx * dx) + (dy * dy))
			if dist <= float64(radius) {
				// normalize 0 (center) to 1 (edge)
				t := dist / float64(radius)
				alpha := uint8((1 - t) * 255) // fade out
				img.Set(x, y, color.RGBA{r, g, b, alpha})
			}
		}
	}

	return img
}

func CreateLightTexture(radius int) *ebiten.Image {
	maxAlpha := 80
	img := ebiten.NewImage(radius*2, radius*2)
	for y := 0; y < radius*2; y++ {
		for x := 0; x < radius*2; x++ {
			dx := float64(x - radius)
			dy := float64(y - radius)
			dist := dx*dx + dy*dy
			if dist < float64(radius*radius) {
				alpha := 1 - (dist / float64(radius*radius))
				a := uint8(alpha * 255)
				if a > uint8(maxAlpha) {
					a = uint8(maxAlpha)
				}
				col := color.NRGBA{
					R: 255,
					G: 255,
					B: 255,
					A: a,
				}
				img.Set(x, y, col)
			}
		}
	}
	return img
}

func InvertedRadialGradient(radius int, lightColor color.Color) *ebiten.Image {
	// light color defaults to white
	var r, g, b uint8 = 0, 0, 0
	if lightColor != nil {
		r1, g1, b1, _ := lightColor.RGBA()
		r = uint8(r1)
		g = uint8(g1)
		b = uint8(b1)
	}
	size := radius * 2
	img := ebiten.NewImage(size, size)

	center := float64(radius)
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			dx := float64(x) - center
			dy := float64(y) - center
			dist := math.Sqrt((dx * dx) + (dy * dy))
			if dist <= float64(radius) {
				// normalize 0 (center) to 1 (edge)
				t := dist / float64(radius)
				alpha := uint8(t * 180) // fade out
				img.Set(x, y, color.RGBA{r, g, b, alpha})
			}
		}
	}

	return img
}

func DrawMapLighting(screen, scene *ebiten.Image, lights []*Light) {
	drawLights(screen, scene, lights, DARK_NIGHTSKY)
}

func drawLights(screen, scene *ebiten.Image, lights []*Light, darknessTint LightColor) {
	maxLights := 16
	lightPositions := make([]float32, maxLights*2) // X, Y
	lightRadii := make([]float32, maxLights)       // radius
	lightColors := make([]float32, maxLights*3)    // R, G, B

	for i := range lights {
		lights[i].calculateNextRadius()
		l := lights[i]

		// light position
		lightPositions[i*2] = l.X
		lightPositions[i*2+1] = l.Y

		// light radius
		lightRadii[i] = l.currentRadius

		// light color
		lightColors[i*3] = l.LightColor[0]
		lightColors[i*3+1] = l.LightColor[1]
		lightColors[i*3+2] = l.LightColor[2]
	}

	op := &ebiten.DrawRectShaderOptions{}
	op.Images[0] = scene
	op.Uniforms = map[string]interface{}{
		"LightPositions": lightPositions,
		"LightRadii":     lightRadii,
		"LightColors":    lightColors,
		"NightTint":      darknessTint,
	}
	screen.DrawRectShader(display.SCREEN_WIDTH, display.SCREEN_HEIGHT, lightShader, op)
}

func shaderColorScale(c color.Color) LightColor {
	r, g, b, _ := c.RGBA()
	return LightColor{float32(r) / 0xffff, float32(g) / 0xffff, float32(b) / 0xffff}
}
