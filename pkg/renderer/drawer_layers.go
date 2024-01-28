package renderer

import (
	"image/color"
	"runtime"
	"sync"
)

type layerColor struct {
	layer *Layer
	color color.RGBA
}

func (vi *valetudoImage) drawLayers() {
	numWorkers := runtime.NumCPU()
	layerCh := make(chan layerColor, numWorkers)
	wg := &sync.WaitGroup{}

	// Start the workers
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for lc := range layerCh {
				vi.drawLayer(lc.layer, lc.color)
			}
		}()
	}

	// Send layers to the channel
	col := vi.renderer.settings.FloorColor
	for _, l := range vi.layers["floor"] {
		layerCh <- layerColor{l, col}
	}

	col = vi.renderer.settings.ObstacleColor
	for _, l := range vi.layers["wall"] {
		layerCh <- layerColor{l, col}
	}

	for _, l := range vi.layers["segment"] {
		col = vi.segmentColor[l.MetaData.SegmentId]
		layerCh <- layerColor{l, col}
	}

	// Close the channel to signal the workers to stop
	close(layerCh)

	// Wait for all workers to finish
	wg.Wait()
}

func (vi *valetudoImage) drawLayer(l *Layer, col color.RGBA) {
	for i := 0; i < len(l.CompressedPixels); i += 3 {
		drawX := l.CompressedPixels[i] - vi.robotCoords.minX
		drawY := l.CompressedPixels[i+1] - vi.robotCoords.minY
		count := l.CompressedPixels[i+2]

		for c := 0; c < count; c++ {
			x, y := vi.RotateLayer(drawX+c, drawY)
			vi.img.SetRGBA(x, y, col)
		}
	}
}
