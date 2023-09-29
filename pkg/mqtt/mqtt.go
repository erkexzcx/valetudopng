package mqtt

import (
	"github.com/erkexzcx/valetudopng/pkg/config"
)

func Start(c *config.MQTTConfig, mapJSONChan, renderedMapChan, calibrationDataChan chan []byte) {
	go startConsumer(c, mapJSONChan)
	go startProducer(c, renderedMapChan, calibrationDataChan)
}
