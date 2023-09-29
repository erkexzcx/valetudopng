package renderer

import (
	"encoding/json"
)

type CalibrationPoint struct {
	Vacuum *CalibrationCoords `json:"vacuum"`
	Map    *CalibrationCoords `json:"map"`
}

type CalibrationCoords struct {
	X int `json:"x"`
	Y int `json:"y"`
}

func (vi *valetudoImage) getCalibrationPointsJSON() []byte {
	calImgx1, calImgy1 := vi.layoutImageCoordRotate(0, 0)
	calImgx2, calImgy2 := vi.layoutImageCoordRotate(vi.robotCoords.maxX-vi.robotCoords.minX, 0)
	calImgx3, calImgy3 := vi.layoutImageCoordRotate(vi.robotCoords.maxX-vi.robotCoords.minX, vi.robotCoords.maxY-vi.robotCoords.minY)
	scale := int(vi.renderer.settings.Scale)

	data := []*CalibrationPoint{
		{
			Vacuum: &CalibrationCoords{vi.robotCoords.minX * vi.valetudoJSON.PixelSize, vi.robotCoords.minY * vi.valetudoJSON.PixelSize},
			Map:    &CalibrationCoords{calImgx1 * scale, calImgy1 * scale},
		},
		{
			Vacuum: &CalibrationCoords{vi.robotCoords.maxX * vi.valetudoJSON.PixelSize, vi.robotCoords.minY * vi.valetudoJSON.PixelSize},
			Map:    &CalibrationCoords{calImgx2 * scale, calImgy2 * scale},
		},
		{
			Vacuum: &CalibrationCoords{vi.robotCoords.maxX * vi.valetudoJSON.PixelSize, vi.robotCoords.maxY * vi.valetudoJSON.PixelSize},
			Map:    &CalibrationCoords{calImgx3 * scale, calImgy3 * scale},
		},
	}

	jsonData, _ := json.Marshal(data)
	return jsonData
}
