package renderer

import (
	"image"
	"image/png"
	"math"

	"github.com/erkexzcx/valetudopng"
	"github.com/erkexzcx/valetudopng/pkg/config"
	"golang.org/x/image/draw"
	"golang.org/x/image/math/f64"
)

type Renderer struct {
	assetRobot   map[int]image.Image
	assetCharger image.Image
	settings     *Settings
}

type Settings struct {
	Scale          float64
	PNGCompression int
	RotationTimes  int

	// Hardcoded limits for a map within robot's coordinates system
	StaticStartX, StaticStartY int
	StaticEndX, StaticEndY     int
}

func New(s *Settings) *Renderer {
	switch s.PNGCompression {
	case 0:
		pngEncoder.CompressionLevel = png.BestSpeed
	case 1:
		pngEncoder.CompressionLevel = png.BestCompression
	case 2:
		pngEncoder.CompressionLevel = png.DefaultCompression
	case 3:
		pngEncoder.CompressionLevel = png.NoCompression
	}

	r := &Renderer{
		settings: s,
	}
	loadAssetRobot(r)
	loadAssetCharger(r)
	return r
}

func (r *Renderer) Render(data []byte, mc *config.MapConfig) (*Result, error) {
	// Parse data to JSON object
	JSON, err := toJSON(data)
	if err != nil {
		return nil, err
	}

	// Render image
	vi := newValetudoImage(JSON, r)
	vi.DrawAll()

	img := vi.ggContext.Image()
	return &Result{
		Image: &img,
		ImageSize: &ImgSize{
			Width:  vi.scaledImgWidth,
			Height: vi.scaledImgHeight,
		},
		RobotCoords: &RbtCoords{
			MinX: vi.robotCoords.minX,
			MinY: vi.robotCoords.minY,
			MaxX: vi.robotCoords.maxX,
			MaxY: vi.robotCoords.maxY,
		},
		Settings:    vi.renderer.settings,
		Calibration: vi.getCalibrationPointsJSON(),
		PixelSize:   vi.valetudoJSON.PixelSize,
	}, nil
}

func loadAssetRobot(r *Renderer) {
	file, err := valetudopng.ResFS.Open("res/robot.png")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	img, err := png.Decode(file)
	if err != nil {
		panic(err)
	}

	r.assetRobot = make(map[int]image.Image, 360)
	for degree := 0; degree < 360; degree++ {
		// Create a new image with the same dimensions as the original image
		rotatedImg := image.NewRGBA(img.Bounds())

		// Create a rotation matrix
		rotationMatrix := f64.Aff3{}
		sin, cos := math.Sincos(math.Pi * float64(degree) / 180.0)
		rotationMatrix[0], rotationMatrix[1] = cos, -sin
		rotationMatrix[3], rotationMatrix[4] = sin, cos

		// Adjust the rotation matrix to rotate around the center of the image
		rotationMatrix[2], rotationMatrix[5] = float64(img.Bounds().Dx())/2*(1-cos)+float64(img.Bounds().Dy())/2*sin, float64(img.Bounds().Dy())/2*(1-cos)-float64(img.Bounds().Dx())/2*sin

		// Use the rotation matrix to rotate the image
		draw.BiLinear.Transform(rotatedImg, rotationMatrix, img, img.Bounds(), draw.Over, nil)

		// Create a new image with the scaled dimensions
		newWidth := rotatedImg.Bounds().Dx() * int(r.settings.Scale) / 4
		newHeight := rotatedImg.Bounds().Dy() * int(r.settings.Scale) / 4
		scaledImg := image.NewRGBA(image.Rect(0, 0, newWidth, newHeight))

		// Draw the rotated image onto the scaled image
		draw.BiLinear.Scale(scaledImg, scaledImg.Bounds(), rotatedImg, rotatedImg.Bounds(), draw.Over, nil)

		r.assetRobot[degree] = scaledImg
	}
}

func loadAssetCharger(r *Renderer) {
	file, err := valetudopng.ResFS.Open("res/charger.png")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	img, err := png.Decode(file)
	if err != nil {
		panic(err)
	}

	newWidth := img.Bounds().Dx() * int(r.settings.Scale) / 4
	newHeight := img.Bounds().Dy() * int(r.settings.Scale) / 4
	scaledImg := image.NewRGBA(image.Rect(0, 0, newWidth, newHeight))

	draw.BiLinear.Scale(scaledImg, scaledImg.Bounds(), img, img.Bounds(), draw.Over, nil)
	r.assetCharger = scaledImg
}
