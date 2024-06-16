package mqtt

import (
	"context"
	"log/slog"
	"sync"

	"github.com/erkexzcx/valetudopng/pkg/config"
)

func Start(ctx context.Context, pwg *sync.WaitGroup, panic chan bool, c *config.MQTTConfig, mapJSONChan, renderedMapChan, calibrationDataChan chan []byte) {
	defer pwg.Done()

	var wg sync.WaitGroup
	wg.Add(2)
	go startConsumer(ctx, &wg, panic, c, mapJSONChan)
	go startProducer(ctx, &wg, panic, c, renderedMapChan, calibrationDataChan)
	wg.Wait()
	slog.Info("MQTT shutting down")
}
