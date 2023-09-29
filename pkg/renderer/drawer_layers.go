package renderer

import (
	"image/color"
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
