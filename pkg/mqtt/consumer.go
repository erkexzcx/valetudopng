package mqtt

import (
	"context"
	"log/slog"
	"sync"
	"time"

	mqttgo "github.com/eclipse/paho.mqtt.golang"
	"github.com/erkexzcx/valetudopng/pkg/config"
	"github.com/erkexzcx/valetudopng/pkg/mqtt/decoder"
)

func startConsumer(ctx context.Context, wg *sync.WaitGroup, panic chan bool, c *config.MQTTConfig, mapJSONChan chan []byte) {
	defer wg.Done()
	opts := mqttgo.NewClientOptions()

	if c.Connection.TLSEnabled {
		opts.AddBroker("ssl://" + c.Connection.Host + ":" + c.Connection.Port)
		tlsConfig, err := newTLSConfig(c.Connection.TLSCaPath, c.Connection.TLSInsecure, c.Connection.TLSMinVersion)
		if err != nil {
			panic <- true
		}
		opts.SetTLSConfig(tlsConfig)
	} else {
		opts.AddBroker("tcp://" + c.Connection.Host + ":" + c.Connection.Port)
	}

	opts.SetClientID(c.Connection.ClientIDPrefix + "_consumer")
	opts.SetUsername(c.Connection.Username)
	opts.SetPassword(c.Connection.Password)
	opts.SetAutoReconnect(true)

	// On received message
	var handler mqttgo.MessageHandler = func(client mqttgo.Client, msg mqttgo.Message) {
		consumerMapDataReceiveHandler(client, msg, mapJSONChan)
	}
	opts.SetDefaultPublishHandler(handler)

	// On connection
	opts.OnConnect = func(client mqttgo.Client) {
		slog.Info("[MQTT consumer] Connected")
	}

	// On disconnection
	opts.OnConnectionLost = func(client mqttgo.Client, err error) {
		slog.Error("[MQTT consumer] Connection lost", slog.String("error", err.Error()))
	}

	// Initial connection
	client := mqttgo.NewClient(opts)

	const maxTries = 24
	success := false
	for i := 1; i <= maxTries; i++ { // try to connect for 2 minutes, then give up
		if token := client.Connect(); token.Wait() && token.Error() != nil {
			slog.Error("[MQTT consumer] Failed to connect to MQTT, trying again in 5s", slog.Int("tries left", maxTries-i), slog.String("type", "consumer"), slog.String("error", token.Error().Error()))
			time.Sleep(5 * time.Second)
		} else {
			success = true
			break
		}
	}
	if !success {
		slog.Error("[MQTT consumer] failed to connect to MQTT")
		panic <- true
		return
	}

	topic := c.Topics.ValetudoPrefix + "/" + c.Topics.ValetudoIdentifier + "/MapData/map-data"
	slog.Info("[MQTT consumer] Subscribing to topic", slog.String("topic", topic))
	token := client.Subscribe(topic, 1, nil)
	token.Wait()
	slog.Info("[MQTT consumer] Subscribed to map data topic")

DONE:
	for {
		select {
		case <-ctx.Done():
			break DONE
		case <-panic:
			break DONE
		}
	}
	client.Disconnect(0)
	slog.Info("[MQTT consumer] shutdown")
}

func consumerMapDataReceiveHandler(_ mqttgo.Client, msg mqttgo.Message, mapJSONChan chan []byte) {
	payload, err := decoder.Decode(msg.Payload())
	if err != nil {
		slog.Error("[MQTT consumer] Failed to parse MQTT message", slog.String("error", err.Error()))
		return
	}
	slog.Debug("[MQTT consumer] Received message")
	mapJSONChan <- payload
}
