package renderer

import (
	"image"
	"image/color"
	"math"
	"runtime"
	"sync"

	"github.com/fogleman/gg"
)

type valetudoImage struct {
	img       *image.RGBA // This is used until image is upscaled
	ggContext *gg.Context // This is used after upscale is done

	// Store img width and height
	unscaledImgWidth  int
	unscaledImgHeight int
	scaledImgWidth    int
	scaledImgHeight   int

	// JSON data
	valetudoJSON *ValetudoJSON

	// Renderer reference, for easy access
	renderer *Renderer

	// Store details about the image within the robots coordinates system
	robotCoords struct {
		minX int
		minY int
		maxX int
		maxY int
	}

	// For faster acess, store them here
	layers   map[string][]*Layer
	entities map[string][]*Entity

	// Segment ID to segment (room) color
	segmentColor map[string]color.RGBA

	// Rotation functions
	RotateLayer  rotationFunc
	RotateEntity rotationFunc
}

func newValetudoImage(valetudoJSON *ValetudoJSON, r *Renderer) *valetudoImage {
	// Create new object
	vi := &valetudoImage{
		valetudoJSON: valetudoJSON,
		renderer:     r,
	}

	// Prepare layers and entities (to speed up iterations)
	vi.layers = make(map[string][]*Layer)
	vi.entities = make(map[string][]*Entity)
	for _, layer := range vi.valetudoJSON.Layers {
		_, found := vi.layers[layer.Type]
		if !found {
			vi.layers[layer.Type] = []*Layer{layer}
		} else {
			vi.layers[layer.Type] = append(vi.layers[layer.Type], layer)
		}
	}
	for _, entity := range vi.valetudoJSON.Entities {
		_, found := vi.entities[entity.Type]
		if !found {
			vi.entities[entity.Type] = []*Entity{entity}
		} else {
			vi.entities[entity.Type] = append(vi.entities[entity.Type], entity)
		}
	}

	// Load colors for each segment
	vi.segmentColor = make(map[string]color.RGBA)
	vi.findFourColors()

	// Find map bounds within robot's coordinates system (from given layers)
	vi.robotCoords.minX = math.MaxInt32
	vi.robotCoords.minY = math.MaxInt32
	vi.robotCoords.maxX = 0
	vi.robotCoords.maxY = 0

	// Either use user's static robot's coordinates, or find them dynamically
	if vi.renderer.settings.StaticStartX == 0 && vi.renderer.settings.StaticStartY == 0 &&
		vi.renderer.settings.StaticEndX == 0 && vi.renderer.settings.StaticEndY == 0 {

		for _, layer := range valetudoJSON.Layers {
			if layer.Dimensions.X.Min < vi.robotCoords.minX {
				vi.robotCoords.minX = layer.Dimensions.X.Min
			}
			if layer.Dimensions.Y.Min < vi.robotCoords.minY {
				vi.robotCoords.minY = layer.Dimensions.Y.Min
			}
			if layer.Dimensions.X.Max > vi.robotCoords.maxX {
				vi.robotCoords.maxX = layer.Dimensions.X.Max
			}
			if layer.Dimensions.Y.Max > vi.robotCoords.maxY {
				vi.robotCoords.maxY = layer.Dimensions.Y.Max
			}
		}
	} else {

		vi.robotCoords.minX = vi.renderer.settings.StaticStartX / 5
		vi.robotCoords.minY = vi.renderer.settings.StaticStartY / 5
		vi.robotCoords.maxX = vi.renderer.settings.StaticEndX / 5
		vi.robotCoords.maxY = vi.renderer.settings.StaticEndY / 5
	}

	// +1 because width is count of pixels, not difference
	// "123456", so if you perform 5-3, you get 2, but actually it's 345, so+1 and it's 3
	vi.unscaledImgWidth = vi.robotCoords.maxX - vi.robotCoords.minX + 1
	vi.unscaledImgHeight = vi.robotCoords.maxY - vi.robotCoords.minY + 1

	// Switch width and height if needed according to rotation
	if vi.renderer.settings.RotationTimes%2 != 0 {
		vi.unscaledImgWidth, vi.unscaledImgHeight = vi.unscaledImgHeight, vi.unscaledImgWidth
	}

	// Create a new image
	vi.img = image.NewRGBA(image.Rect(0, 0, vi.unscaledImgWidth, vi.unscaledImgHeight))

	// Explanation about image.Rect (documentation is lying):
	//
	// img := image.NewRGBA(image.Rect(0, 0, 100, 100))
	// would result in an image that has X from 0 to 99, Y from 0 to 99
	// width 100 and height 100

	// Create rotation funcs
	vi.RotateLayer = vi.getRotationFunc(true)
	vi.RotateEntity = vi.getRotationFunc(false)

	return vi
}

func (vi *valetudoImage) DrawAll() {
	vi.drawLayers()
	vi.upscaleToGGContext()

	// Draw path entity
	vi.ggContext.SetRGB255(255, 255, 255)
	vi.ggContext.SetLineWidth(float64(vi.renderer.settings.Scale) * 0.75)
	for _, e := range vi.entities["path"] {
		vi.drawEntityPath(e)
	}
	vi.ggContext.Stroke()

	// Draw virtual_wall entities
	vi.ggContext.SetRGBA255(255, 0, 0, 192)
	vi.ggContext.SetLineWidth(float64(vi.renderer.settings.Scale) * 1.5)
	vi.ggContext.SetLineCapButt()
	for _, e := range vi.entities["virtual_wall"] {
		vi.drawEntityVirtualWall(e)
	}
	vi.ggContext.Stroke()
	// Draw no_go_area entities
	lineWidth := float64(vi.renderer.settings.Scale * 0.5)
	noGoAreas := vi.entities["no_go_area"]
	vi.ggContext.SetRGBA255(255, 0, 0, 75)
	vi.ggContext.SetLineWidth(0)
	for _, e := range noGoAreas {
		vi.drawEntityNoGoArea(e)
	}
	vi.ggContext.Fill()
	vi.ggContext.SetRGB255(255, 0, 0)
	vi.ggContext.SetLineWidth(lineWidth)
	for _, e := range noGoAreas {
		vi.drawEntityNoGoArea(e)
	}
	vi.ggContext.Stroke()

	// Draw charger_location entity
	for _, e := range vi.entities["charger_location"] {
		vi.drawEntityCharger(e, 0, 0)
	}

	// Draw robot_position entity
	for _, e := range vi.entities["robot_position"] {
		vi.drawEntityRobot(e, int(vi.renderer.settings.Scale)/2, -1)
	}
}

func (vi *valetudoImage) upscaleToGGContext() {
	scale := int(vi.renderer.settings.Scale)
	scaledImgWidth := vi.unscaledImgWidth * scale
	scaledImgHeight := vi.unscaledImgHeight * scale
	scaledImg := image.NewRGBA(image.Rect(0, 0, scaledImgWidth, scaledImgHeight))

	numCPUs := runtime.NumCPU()
	var wg sync.WaitGroup
	jobs := make(chan int, vi.unscaledImgHeight)

	// Start workers
	for w := 0; w < numCPUs; w++ {
		go func() {
			for y := range jobs {
				yScale := y * scale
				yUnscaledImgWidth := y * vi.unscaledImgWidth
				for x := 0; x < vi.unscaledImgWidth; x++ {
					xScale := x * scale
					for scaleIndex := 0; scaleIndex < scale; scaleIndex++ {
						copy(scaledImg.Pix[(yScale*scaledImgWidth+xScale+scaleIndex)*4:(yScale*scaledImgWidth+xScale+scaleIndex+1)*4], vi.img.Pix[(yUnscaledImgWidth+x)*4:(yUnscaledImgWidth+x+1)*4])
					}
				}
				for scaleIndex := 1; scaleIndex < scale; scaleIndex++ {
					copy(scaledImg.Pix[((yScale+scaleIndex)*scaledImgWidth)*4:((yScale+scaleIndex+1)*scaledImgWidth)*4], scaledImg.Pix[(yScale*scaledImgWidth)*4:(yScale+1)*scaledImgWidth*4])
				}
				wg.Done()
			}
		}()
	}

	// Distribute work
	for y := 0; y < vi.unscaledImgHeight; y++ {
		wg.Add(1)
		jobs <- y
	}
	close(jobs)

	// Wait for all workers to finish
	wg.Wait()

	vi.ggContext = gg.NewContextForRGBA(scaledImg)
	vi.scaledImgWidth = scaledImgWidth
	vi.scaledImgHeight = scaledImgHeight
}

type rotationFunc func(x, y int) (int, int)

// For layers, "subtractOne" should be true
// For entities, "subtractOne" should be false
func (vi *valetudoImage) getRotationFunc(subtractOne bool) rotationFunc {
	switch vi.renderer.settings.RotationTimes {
	case 0:
		// No rotation
		return func(x, y int) (int, int) { return x, y }
	case 1:
		// 90 degrees clockwise
		if subtractOne {
			return func(x, y int) (int, int) { return vi.unscaledImgWidth - 1 - y, x }
		}
		return func(x, y int) (int, int) { return vi.unscaledImgWidth - y, x }
	case 2:
		// 180 degrees clockwise
		if subtractOne {
			return func(x, y int) (int, int) { return vi.unscaledImgWidth - 1 - x, vi.unscaledImgHeight - 1 - y }
		}
		return func(x, y int) (int, int) { return vi.unscaledImgWidth - x, vi.unscaledImgHeight - y }
	case 3:
		// 270 degrees clockwise
		if subtractOne {
			return func(x, y int) (int, int) { return y, vi.unscaledImgHeight - 1 - x }
		}
		return func(x, y int) (int, int) { return y, vi.unscaledImgHeight - x }
	}
	return func(x, y int) (int, int) { return x, y }
}
