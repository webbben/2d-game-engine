package proc_gen

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math/rand"
	"os"
	"time"

	"github.com/aquilax/go-perlin"
)

const (
	/*
		"Lacunarity"

		Controls the frequency of the octaves.

		Higher lacunarity leads to more fine details and smaller features in the generated noise.

		"The weight when the sum is formed. as it approaches 1 the function gets noisier"
	*/
	alpha = 3
	/*
		"Persistence"

		Controls the amplitude of the octaves in the perlin noise.

		Higher persistence values lead to more contrast and roughness in the generated noise, while lower values
		result in smoother patterns.

		"the harmonic scaling/spacing.  typically it is 2."
	*/
	beta = 2
	/*
		Octaves

		Determine the number of layers of noise that are combined to create the final result.

		Each layer adds more complexity to the generated noise, but also increases computational cost.
	*/
	oct = 3
)

var (
	townElevationParams = []float64{2, 2, 1, 2}
	mountainParams      = []float64{1.7, 1, 4, 4}
	forestParams        = []float64{1.1, 4, 4, 3}
)

func GenerateForest(width int, height int) {
	noiseMap := generateNoiseWithParams(width, height, forestParams[0], forestParams[1], int32(forestParams[2]), int(forestParams[3]))
	//noiseMap = downSample(noiseMap, 1)
	NoiseMapToPNG(noiseMap, int(forestParams[3]))
}
func GenerateTownElevation(width int, height int) [][]int {
	noiseMap := generateNoiseWithParams(width, height, townElevationParams[0], townElevationParams[1], int32(townElevationParams[2]), int(townElevationParams[3]))
	noiseMap = downSample(noiseMap, 2)
	//NoiseMapToPNG(noiseMap, int(townElevationParams[3]))
	return noiseMap
}
func GenerateMountain(width int, height int) {
	noiseMap := generateNoiseWithParams(width, height, mountainParams[0], mountainParams[1], int32(mountainParams[2]), int(mountainParams[3]))
	noiseMap = downSample(noiseMap, 3)
	NoiseMapToPNG(noiseMap, int(mountainParams[3]))
}

func generateNoiseWithParams(width int, height int, a float64, b float64, n int32, levels int) [][]int {
	rand.Seed(time.Now().UnixNano())
	seed := rand.Int63()
	//seed := int64(3)
	p := perlin.NewPerlin(a, b, n, seed)

	noiseMap := make([][]float64, height)
	minVal := float64(1000)
	maxVal := float64(-1000)

	// get the noise map
	for y := 0; y < height; y++ {
		row := make([]float64, width)
		for x := 0; x < width; x++ {
			value := p.Noise2D(float64(x)/float64(width), float64(y)/float64(height))
			if value > maxVal {
				maxVal = value
			}
			if value < minVal {
				minVal = value
			}
			row[x] = value
		}
		noiseMap[y] = row
	}

	// scale the noise map's values to the range [0, levels - 1]
	scaledMap := make([][]int, height)
	for y := 0; y < height; y++ {
		row := make([]int, width)
		for x := 0; x < width; x++ {
			value := noiseMap[y][x]
			row[x] = scaleValue(value, minVal, maxVal, 0, levels)
		}
		scaledMap[y] = row
	}

	return scaledMap
}

func scaleValue(value float64, minOriginal float64, maxOriginal float64, scaleMin int, scaleMax int) int {
	return int(((value-minOriginal)/(maxOriginal-minOriginal))*float64(scaleMax-scaleMin) + float64(scaleMin))
}

func downSample(noiseMap [][]int, downSampleBy int) [][]int {
	width := len(noiseMap[0])
	height := len(noiseMap)
	downsampled := make([][]int, height)
	for y := 0; y < height; y++ {
		downsampled[y] = make([]int, width)
	}

	// downsample the original data
	for y := 0; y < height; y += downSampleBy {
		if y >= height {
			break
		}
		for x := 0; x < width; x += downSampleBy {
			if x >= width {
				break
			}
			average := averageRegionValue(noiseMap, x, y, downSampleBy)

			for y1 := 0; y1 < downSampleBy; y1++ {
				for x1 := 0; x1 < downSampleBy; x1++ {
					updateX := x + x1
					updateY := y + y1

					if updateX < width && updateY < height {
						downsampled[updateX][updateY] = average
					}
				}
			}
		}
	}
	return downsampled
}

// averages the values in the noise data for every n by n sample
func averageRegionValue(data [][]int, startX, startY int, sampleSize int) int {
	var sum int
	maxY := len(data)
	maxX := len(data[0])
	sampleCount := 0
	for y := startY; y < startY+sampleSize; y++ {
		if y >= maxY {
			break
		}
		for x := startX; x < startX+sampleSize; x++ {
			if x >= maxX {
				break
			}
			sum += data[y][x]
			sampleCount++
		}
	}
	average := sum / sampleCount
	return average
}

func NoiseMapToPNG(noiseMap [][]int, levels int) {
	width := len(noiseMap[0])
	height := len(noiseMap)
	img := image.NewGray(image.Rect(0, 0, width, height))

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			value := noiseMap[y][x]
			//gray := uint8(value * 255 / (levels - 1)) // map values from [-1,1] to [0,255]
			gray := 255 - scaleValue(float64(value), 0, float64(levels), 0, (255/5)*levels)
			img.SetGray(x, y, color.Gray{Y: uint8(gray)})
		}
	}

	file, err := os.Create("noise_map.png")
	if err != nil {
		fmt.Println("error generating noise map png")
		return
	}
	defer file.Close()

	err = png.Encode(file, img)
	if err != nil {
		fmt.Println("failed to encode image for noise map")
	}
}
