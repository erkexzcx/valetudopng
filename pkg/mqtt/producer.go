package mqtt

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/bitly/go-simplejson"
	mqttgo "github.com/eclipse/paho.mqtt.golang"
	"github.com/erkexzcx/valetudopng/pkg/config"
)

type Device struct {
	Name        string   `json:"name"`
	Identifiers []string `json:"identifiers"`
}

type Map struct {
	Name     string `json:"name"`
	UniqueID string `json:"unique_id"`
	Device   Device `json:"device"`
	Topic    string `json:"topic"`
}

func startProducer(ctx context.Context, pwg *sync.WaitGroup, panic chan bool, c *config.MQTTConfig, renderedMapChan, calibrationDataChan chan []byte) {
	defer pwg.Done()
	opts := mqttgo.NewClientOptions()

	if c.Connection.TLSEnabled {
		opts.AddBroker("ssl://" + c.Connection.Host + ":" + c.Connection.Port)
		tlsConfig, err := newTLSConfig(c.Connection.TLSCaPath, c.Connection.TLSInsecure, c.Connection.TLSMinVersion)
		if err != nil {
			slog.Error("Failed to setup TLS config", slog.String("error", err.Error()))
			panic <- true
		}
		opts.SetTLSConfig(tlsConfig)
	} else {
		opts.AddBroker("tcp://" + c.Connection.Host + ":" + c.Connection.Port)
	}

	opts.SetClientID(c.Connection.ClientIDPrefix + "_producer")
	opts.SetUsername(c.Connection.Username)
	opts.SetPassword(c.Connection.Password)
	opts.SetAutoReconnect(true)

	// On connection
	opts.OnConnect = func(client mqttgo.Client) {
		slog.Info("[MQTT producer] Connected")
	}

	// On disconnection
	opts.OnConnectionLost = func(client mqttgo.Client, err error) {
		slog.Error("[MQTT producer] Connection lost: %v", slog.String("error", err.Error()))
	}

	// Initial connection
	client := mqttgo.NewClient(opts)
	const maxTries = 24
	success := false
	for i := 1; i <= maxTries; i++ { // try to connect for 2 minutes, then give up
		if token := client.Connect(); token.Wait() && token.Error() != nil {
			slog.Error("[MQTT producer] Failed to connect to MQTT, trying again in 5s", slog.Int("tries left", maxTries-i), slog.String("type", "consumer"), slog.String("error", token.Error().Error()))
			time.Sleep(5 * time.Second)
		} else {
			success = true
			break
		}
	}
	if !success {
		slog.Error("Publisher failed to connect to MQTT")
		panic <- true
		return
	}
	var wg sync.WaitGroup
	wg.Add(2)
	renderedMapTopic := c.Topics.ValetudoPrefix + "/" + c.Topics.ValetudoIdentifier + "/MapData/map"
	go produceAnnounceMapTopic(&wg, client, renderedMapTopic, c)
	go producerMapUpdatesHandler(ctx, &wg, panic, client, renderedMapChan, renderedMapTopic, c)

	wg.Add(2)
	calibrationTopic := c.Topics.ValetudoPrefix + "/" + c.Topics.ValetudoIdentifier + "/MapData/calibration"
	go producerAnnounceCalibrationTopic(&wg, client, calibrationTopic, c)
	go producerCalibrationDataHandler(ctx, &wg, panic, client, calibrationDataChan, calibrationTopic, c)

	done := make(chan bool)
	go func() {
		wg.Wait()
		done <- true
	}()
DONE:
	for {
		select {
		case <-done:
			break DONE
		case <-panic:
			break DONE
		case <-ctx.Done():
			break DONE
		}
	}
	slog.Info("[MQTT producer] shutdown")
}

func producerMapUpdatesHandler(ctx context.Context, wg *sync.WaitGroup, panic chan bool, client mqttgo.Client, renderedMapChan chan []byte, topic string, c *config.MQTTConfig) {
	defer wg.Done()
	for {
		select {
		case img := <-renderedMapChan:
			token := client.Publish(topic, 1, true, img)
			if ok := token.WaitTimeout(c.SendTimeout); !ok || token.Error() != nil {
				slog.Error("[MQTT producer] Failed to publish", slog.String("error", token.Error().Error()))
			} else {
				slog.Debug("[MQTT producer] published a message")
			}
		case <-ctx.Done():
			return
		case <-panic:
			return
		}
	}
}

func produceAnnounceMapTopic(wg *sync.WaitGroup, client mqttgo.Client, rmt string, c *config.MQTTConfig) {
	defer wg.Done()
	announceTopic := c.Topics.HaAutoconfPrefix + "/camera/" + c.Topics.ValetudoIdentifier + "/" + c.Topics.ValetudoPrefix + "_" + c.Topics.ValetudoIdentifier + "_map/config"

	js := simplejson.New()
	js.Set("name", "Map")
	js.Set("unique_id", c.Topics.ValetudoIdentifier+"_rendered_map")
	js.Set("topic", rmt)

	device := simplejson.New()
	device.Set("name", c.Topics.ValetudoIdentifier)
	device.Set("identifiers", []string{c.Topics.ValetudoIdentifier})

	js.Set("device", device)

	announcementData, err := js.MarshalJSON()
	if err != nil { // this isn't really possible
		slog.Error("[MQTT producer] failed to parse annoucement")
		return
	}

	token := client.Publish(announceTopic, 1, true, announcementData)
	if ok := token.WaitTimeout(c.SendTimeout); !ok || token.Error() != nil {
		slog.Error("[MQTT producer] Failed to publish", slog.String("error", token.Error().Error()))
	} else {
		slog.Debug("[MQTT producer] published AnnounceMapTopic")
	}
}

func producerCalibrationDataHandler(ctx context.Context, wg *sync.WaitGroup, panic chan bool, client mqttgo.Client, calibrationDataChan chan []byte, topic string, c *config.MQTTConfig) {
	defer wg.Done()
	for {
		select {
		case img := <-calibrationDataChan:
			token := client.Publish(topic, 1, true, img)
			if ok := token.WaitTimeout(c.SendTimeout); !ok || token.Error() != nil {
				slog.Error("[MQTT producer] Failed to publish", slog.String("error", token.Error().Error()))
			} else {
				slog.Debug("[MQTT producer] published a message")
			}
		case <-ctx.Done():
			return
		case <-panic:
			return
		}
	}
}

func producerAnnounceCalibrationTopic(wg *sync.WaitGroup, client mqttgo.Client, cdt string, c *config.MQTTConfig) {
	defer wg.Done()
	announceTopic := c.Topics.HaAutoconfPrefix + "/sensor/" + c.Topics.ValetudoIdentifier + "/" + c.Topics.ValetudoPrefix + "_" + c.Topics.ValetudoIdentifier + "_calibration/config"

	js := simplejson.New()
	js.Set("name", "Calibration")
	js.Set("unique_id", c.Topics.ValetudoIdentifier+"_calibration")
	js.Set("state_topic", cdt)

	device := simplejson.New()
	device.Set("name", c.Topics.ValetudoIdentifier)
	device.Set("identifiers", []string{c.Topics.ValetudoIdentifier})

	js.Set("device", device)

	announcementData, err := js.MarshalJSON()
	if err != nil { // this isn't really possible
		slog.Error("[MQTT producer] failed to parse CalibrationTopic annoucement")
		return
	}

	token := client.Publish(announceTopic, 1, true, announcementData)
	if ok := token.WaitTimeout(c.SendTimeout); !ok || token.Error() != nil {
		slog.Error("[MQTT producer] Failed to publish", slog.String("error", token.Error().Error()))
	} else {
		slog.Debug("[MQTT producer] published AnnounceCalibrationTopic")
	}
}
