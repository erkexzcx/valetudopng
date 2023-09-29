package renderer

import (
	"image/color"
	"math"
	"sync"
)

func (vi *valetudoImage) drawLayer(l *Layer, col color.RGBA, wg *sync.WaitGroup) {
	wg.Add(1)
	go func() {
		for i := 0; i < len(l.CompressedPixels); i += 3 {
			drawX := l.CompressedPixels[i] - vi.robotCoords.minX
			drawY := l.CompressedPixels[i+1] - vi.robotCoords.minY
			count := l.CompressedPixels[i+2]

			for c := 0; c < count; c++ {
				x, y := vi.layoutImageCoordRotate(drawX+c, drawY)
				vi.img.Set(x, y, col)
			}
		}
		wg.Done()
	}()
}

func getSegmentColor(value int, maxLayers int) color.RGBA {
	hue := (float64(value)/float64(maxLayers))*120 + 60 // control hue between 60 and 180
	hue = hue * math.Pi / 180                           // converting hue degree to radians as math package needs radians
	r, g, b := 0.0, 0.0, 0.0

	// Modified HSV to RGB conversion to light colors and avoiding blues and reds
	if hue < 2*math.Pi/3 {
		r = math.Max(0.70*(1-math.Cos(hue)/2), 0.6)
		g = math.Max(0.70*(1+math.Cos(hue)/2), 0.6)
		b = math.Max(0.70*(1-math.Sin(hue)/2), 0.6)
	} else {
		hue -= 2 * math.Pi / 3
		r = math.Max(0.70*(1-math.Sin(hue)/2), 0.6)
		g = math.Max(0.70*(1-math.Cos(hue)/2), 0.6)
		b = math.Max(0.70*(1+math.Cos(hue)/2), 0.6)
	}

	return color.RGBA{uint8(r * 255), uint8(g * 255), uint8(b * 255), 255}
}

func (vi *valetudoImage) layoutImageCoordRotate(x, y int) (adjustedX, adjustedY int) {
	switch vi.renderer.settings.RotationTimes {
	case 0:
		// No rotation
		return x, y
	case 1:
		// 90 degrees clockwise
		return vi.unscaledImgWidth - 1 - y, x
	case 2:
		// 180 degrees clockwise
		return vi.unscaledImgWidth - 1 - x, vi.unscaledImgHeight - 1 - y
	case 3:
		// 270 degrees clockwise
		return y, vi.unscaledImgHeight - 1 - x
	}
	return
}
