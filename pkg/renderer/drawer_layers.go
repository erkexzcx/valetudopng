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
				x, y := vi.RotateLayer(drawX+c, drawY)
				vi.img.Set(x, y, col)
			}
		}
		wg.Done()
	}()
}
