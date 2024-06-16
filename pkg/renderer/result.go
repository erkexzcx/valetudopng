package renderer

import (
	"bytes"
	"image"
	"image/png"
)

// Set compression value from 'New' function.
var pngEncoder = png.Encoder{}

type Result struct {
	Image       *image.Image
	ImageSize   *ImgSize
	RobotCoords *RbtCoords
	Settings    *Settings
	Calibration []byte
	PixelSize   int    // taken from JSON, for traslating image coords to robot's coords system coordinates
	CardCfg     []byte // generated configuration for lovelace card
}

type ImgSize struct {
	Width  int
	Height int
}

type RbtCoords struct {
	MinX int
	MinY int
	MaxX int
	MaxY int
}

func (r *Result) RenderPNG() ([]byte, error) {
	var b bytes.Buffer
	err := pngEncoder.Encode(&b, *r.Image)
	if err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}
