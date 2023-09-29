package renderer

import "testing"

func TestImageCoordRotate(t *testing.T) {
	vi := &valetudoImage{
		unscaledImgWidth:  100,
		unscaledImgHeight: 160,
		renderer: &Renderer{
			settings: &Settings{
				RotationTimes: 0,
			},
		},
	}

	testCases := []struct {
		name                 string
		x, y                 int
		rotationTimes        int
		expectedX, expectedY int
	}{
		{"No rotation", 20, 10, 0, 20, 10},
		{"90 degrees rotation", 20, 10, 1, 90, 20},
		{"180 degrees rotation", 20, 10, 2, 80, 150},
		{"270 degrees rotation", 20, 10, 3, 10, 140},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			vi.renderer.settings.RotationTimes = tc.rotationTimes
			adjustedX, adjustedY := vi.layoutImageCoordRotate(tc.x, tc.y)
			if adjustedX != tc.expectedX || adjustedY != tc.expectedY {
				t.Errorf("Expected (%d, %d), but got (%d, %d)", tc.expectedX, tc.expectedY, adjustedX, adjustedY)
			}
		})
	}
}
