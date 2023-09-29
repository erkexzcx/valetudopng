package renderer

import (
	"encoding/json"
)

type ValetudoJSON struct {
	Class     string    `json:"__class"`
	MetaData  *MetaData `json:"metaData"`
	Size      *Size     `json:"size"`
	PixelSize int       `json:"pixelSize"`
	Layers    []*Layer  `json:"layers"`
	Entities  []*Entity `json:"entities"`
}

type MetaData struct {
	VendorMapId    int    `json:"vendorMapId"`
	Version        int    `json:"version"`
	Nonce          string `json:"nonce"`
	TotalLayerArea int    `json:"totalLayerArea"`
	Area           int    `json:"area,omitempty"`
	SegmentId      string `json:"segmentId,omitempty"`
	Active         bool   `json:"active,omitempty"`
	Name           string `json:"name,omitempty"`
}

type Size struct {
	X int `json:"x"`
	Y int `json:"y"`
}

type Layer struct {
	Class            string     `json:"__class"`
	MetaData         MetaData   `json:"metaData"`
	Type             string     `json:"type"`
	Pixels           []int      `json:"pixels"`
	Dimensions       Dimensions `json:"dimensions"`
	CompressedPixels []int      `json:"compressedPixels"`
}

type Dimensions struct {
	X          Dimension `json:"x"`
	Y          Dimension `json:"y"`
	PixelCount int       `json:"pixelCount"`
}

type Dimension struct {
	Min int `json:"min"`
	Max int `json:"max"`
	Mid int `json:"mid"`
	Avg int `json:"avg"`
}

type Entity struct {
	Class    string         `json:"__class"`
	MetaData MetaDataEntity `json:"metaData"`
	Points   []int          `json:"points"`
	Type     string         `json:"type"`
}

type MetaDataEntity struct {
	Angle int `json:"angle,omitempty"`
}

func toJSON(payload []byte) (*ValetudoJSON, error) {
	var JSON *ValetudoJSON
	err := json.Unmarshal(payload, &JSON)
	if err != nil {
		return nil, err
	}
	return JSON, nil
}
