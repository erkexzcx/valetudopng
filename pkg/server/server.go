package server

import (
	"encoding/base64"
	"fmt"
	"log"
	"os"
	"os/signal"
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
		Scale:         c.Map.Scale,
		RotationTimes: c.Map.RotationTimes,

		StaticStartX: c.Map.CustomLimits.StartX,
		StaticStartY: c.Map.CustomLimits.StartY,
		StaticEndX:   c.Map.CustomLimits.EndX,
		StaticEndY:   c.Map.CustomLimits.EndY,
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

		timestamp := time.Now()
		res, err := r.Render(payload, c.Map)
		if err != nil {
			log.Fatalln("Error occurred while rendering map:", err)
		}
		img, err := res.RenderPNG()
		if err != nil {
			log.Fatalln("Error occurred while rendering PNG image:", err)
		}
		elapsedDuration := time.Since(timestamp)
		log.Printf("Image rendered in %v milliseconds\n", elapsedDuration.Milliseconds())

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
