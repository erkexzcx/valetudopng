package renderer

// Entities coordinates are basically same as layers coordinates, just multiplied by
// vi.valetudoJSON.PixelSize value, so simply divide by it and we get their coords at
// 1x scale. Then we can upscale to our scale integer.
func (vi *valetudoImage) entityToImageCoords(vacuumX, vacuumY int) (float64, float64) {
	imgX := (vacuumX/vi.valetudoJSON.PixelSize - vi.robotCoords.minX)
	imgY := (vacuumY/vi.valetudoJSON.PixelSize - vi.robotCoords.minY)
	rotatedX, rotatedY := vi.entityImageCoordRotate(imgX, imgY)
	return float64(rotatedX) * vi.renderer.settings.Scale, float64(rotatedY) * vi.renderer.settings.Scale
}

func (vi *valetudoImage) entityImageCoordRotate(x, y int) (adjustedX, adjustedY int) {
	switch vi.renderer.settings.RotationTimes {
	case 0:
		// No rotation
		return x, y
	case 1:
		// 90 degrees clockwise
		return vi.unscaledImgWidth - y, x
	case 2:
		// 180 degrees clockwise
		return vi.unscaledImgWidth - x, vi.unscaledImgHeight - y
	case 3:
		// 270 degrees clockwise
		return y, vi.unscaledImgHeight - x
	}
	return
}

func (vi *valetudoImage) drawEntityVirtualWall(e *Entity) {
	sx, sy := vi.entityToImageCoords(e.Points[0], e.Points[1])
	ex, ey := vi.entityToImageCoords(e.Points[2], e.Points[3])
	vi.ggContext.DrawLine(sx, sy, ex, ey)
}

func (vi *valetudoImage) drawEntityNoGoArea(e *Entity) {
	sx, sy := vi.entityToImageCoords(e.Points[0], e.Points[1])
	ex, ey := vi.entityToImageCoords(e.Points[4], e.Points[5])

	width := ex - sx
	height := ey - sy
	vi.ggContext.DrawRectangle(sx, sy, width, height)
}

func (vi *valetudoImage) drawEntityPath(e *Entity) {
	sx, sy := vi.entityToImageCoords(e.Points[0], e.Points[1])
	vi.ggContext.MoveTo(sx, sy)
	for i := 2; i < len(e.Points); i += 2 {
		currX, currY := vi.entityToImageCoords(e.Points[i], e.Points[i+1])
		vi.ggContext.LineTo(currX, currY)
	}
}

func (vi *valetudoImage) drawEntityRobot(e *Entity, xOffset, yOffset int) {
	coordX, coordY := vi.entityToImageCoords(e.Points[0], e.Points[1])
	angle := (e.MetaData.Angle + (vi.renderer.settings.RotationTimes * 90)) % 360
	vi.ggContext.DrawImageAnchored(vi.renderer.assetRobot[angle], int(coordX)+xOffset, int(coordY)+yOffset, 0.5, 0.5)
}

func (vi *valetudoImage) drawEntityCharger(e *Entity, xOffset, yOffset int) {
	coordX, coordY := vi.entityToImageCoords(e.Points[0], e.Points[1])
	vi.ggContext.DrawImageAnchored(vi.renderer.assetCharger, int(coordX)+xOffset, int(coordY)+yOffset, 0.5, 0.5)
}
