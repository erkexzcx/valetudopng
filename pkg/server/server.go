package server

import (
	"encoding/base64"
	"fmt"
	"image/color"
	"log"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	"time"

	"github.com/erkexzcx/valetudopng/pkg/config"
	"github.com/erkexzcx/valetudopng/pkg/mqtt"
	"github.com/erkexzcx/valetudopng/pkg/renderer"
)

var (
	renderedPNG    = make([]byte, 0)
	renderedPNGMux = &sync.RWMutex{}
	result         *renderer.Result
)

func Start(c *config.Config) {
	r := renderer.New(&renderer.Settings{
		Scale:          c.Map.Scale,
		PNGCompression: c.Map.PNGCompression,
		RotationTimes:  c.Map.RotationTimes,

		StaticStartX: c.Map.CustomLimits.StartX,
		StaticStartY: c.Map.CustomLimits.StartY,
		StaticEndX:   c.Map.CustomLimits.EndX,
		StaticEndY:   c.Map.CustomLimits.EndY,

		FloorColor:       HexColor(c.Map.Colors.Floor),
		ObstacleColor:    HexColor(c.Map.Colors.Obstacle),
		PathColor:        HexColor(c.Map.Colors.Path),
		NoGoAreaColor:    HexColor(c.Map.Colors.NoGoArea),
		VirtualWallColor: HexColor(c.Map.Colors.VirtualWall),
		SegmentColors: []color.RGBA{
			HexColor(c.Map.Colors.Segments[0]),
			HexColor(c.Map.Colors.Segments[1]),
			HexColor(c.Map.Colors.Segments[2]),
			HexColor(c.Map.Colors.Segments[3]),
		},
	})

	if c.HTTP.Enabled {
		go runWebServer(c.HTTP.Bind)
	}

	mapJSONChan := make(chan []byte)
	renderedMapChan := make(chan []byte)
	calibrationDataChan := make(chan []byte)
	go mqtt.Start(c.Mqtt, mapJSONChan, renderedMapChan, calibrationDataChan)

	renderedAt := time.Now().Add(-c.Map.MinRefreshInt)
	for payload := range mapJSONChan {
		if time.Now().Before(renderedAt) {
			log.Println("Skipping image render due to min_refresh_int")
			continue
		}
		renderedAt = time.Now().Add(c.Map.MinRefreshInt)

		tsStart := time.Now()
		res, err := r.Render(payload, c.Map)
		if err != nil {
			log.Fatalln("Error occurred while rendering map:", err)
		}
		drawnInMS := time.Since(tsStart).Milliseconds()

		img, err := res.RenderPNG()
		if err != nil {
			log.Fatalln("Error occurred while rendering PNG image:", err)
		}
		renderedIn := time.Since(tsStart).Milliseconds() - drawnInMS

		log.Printf("Image rendered! drawing:%dms, encoding:%dms, size:%s\n", drawnInMS, renderedIn, ByteCountSI(int64(len(img))))

		if !(c.Mqtt.ImageAsBase64 && !c.HTTP.Enabled) {
			renderedPNGMux.Lock()
			renderedPNG = img
			result = res
			renderedPNGMux.Unlock()
		}

		if c.Mqtt.ImageAsBase64 {
			img = []byte(base64.StdEncoding.EncodeToString(img))
		}

		// Send data to MQTT
		renderedMapChan <- img
		calibrationDataChan <- res.Calibration
	}

	// Create a channel to wait for OS interrupt signal
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	// Block main function here until an interrupt is received
	<-interrupt
	fmt.Println("Program interrupted")
}

func ByteCountSI(b int64) string {
	const unit = 1000
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f%cB", float64(b)/float64(div), "kMGTPE"[exp])
}

func HexColor(hex string) color.RGBA {
	red, _ := strconv.ParseUint(hex[1:3], 16, 8)
	green, _ := strconv.ParseUint(hex[3:5], 16, 8)
	blue, _ := strconv.ParseUint(hex[5:7], 16, 8)
	alpha := uint64(255)

	if len(hex) > 8 {
		alpha, _ = strconv.ParseUint(hex[7:9], 16, 8)
	}

	return color.RGBA{R: uint8(red), G: uint8(green), B: uint8(blue), A: uint8(alpha)}
}
