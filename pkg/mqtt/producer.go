package mqtt

import (
	"log"
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

func startProducer(c *config.MQTTConfig, renderedMapChan, calibrationDataChan chan []byte) {
	opts := mqttgo.NewClientOptions()

	if c.Connection.TLSEnabled {
		opts.AddBroker("ssl://" + c.Connection.Host + ":" + c.Connection.Port)
		tlsConfig, err := newTLSConfig(c.Connection.TLSCaPath, c.Connection.TLSInsecure, c.Connection.TLSMinVersion)
		if err != nil {
			log.Fatalln(err)
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
		log.Println("[MQTT producer] Connected")
	}

	// On disconnection
	opts.OnConnectionLost = func(client mqttgo.Client, err error) {
		log.Printf("[MQTT producer] Connection lost: %v", err)
	}

	// Initial connection
	client := mqttgo.NewClient(opts)
	for {
		if token := client.Connect(); token.Wait() && token.Error() != nil {
			log.Printf("[MQTT producer] Failed to connect: %v. Retrying in 5 seconds...\n", token.Error())
			time.Sleep(5 * time.Second)
		} else {
			break
		}
	}

	renderedMapTopic := c.Topics.ValetudoPrefix + "/" + c.Topics.ValetudoIdentifier + "/MapData/map"
	go produceAnnounceMapTopic(client, renderedMapTopic, c)
	go producerMapUpdatesHandler(client, renderedMapChan, renderedMapTopic)

	calibrationTopic := c.Topics.ValetudoPrefix + "/" + c.Topics.ValetudoIdentifier + "/MapData/calibration"
	go producerAnnounceCalibrationTopic(client, calibrationTopic, c)
	go producerCalibrationDataHandler(client, calibrationDataChan, calibrationTopic)
}

func producerMapUpdatesHandler(client mqttgo.Client, renderedMapChan chan []byte, topic string) {
	for img := range renderedMapChan {
		token := client.Publish(topic, 1, true, img)
		token.Wait()
		if token.Error() != nil {
			log.Printf("[MQTT producer] Failed to publish: %v\n", token.Error())
		}
	}
}

func produceAnnounceMapTopic(client mqttgo.Client, rmt string, c *config.MQTTConfig) {
	announceTopic := c.Topics.HaAutoconfPrefix + "/camera/" + c.Topics.ValetudoIdentifier + "/" + c.Topics.ValetudoPrefix + "_" + c.Topics.ValetudoIdentifier + "_map/config"

	js := simplejson.New()
	js.Set("name", c.Topics.ValetudoIdentifier+"_map")
	js.Set("unique_id", c.Topics.ValetudoIdentifier+"_rendered_map")
	js.Set("topic", rmt)

	device := simplejson.New()
	device.Set("name", c.Topics.ValetudoIdentifier)
	device.Set("identifiers", []string{c.Topics.ValetudoIdentifier})

	js.Set("device", device)

	announcementData, err := js.MarshalJSON()
	if err != nil {
		panic(err)
	}

	token := client.Publish(announceTopic, 1, false, announcementData)
	token.Wait()
	if token.Error() != nil {
		log.Printf("[MQTT producer] Failed to publish: %v\n", token.Error())
	}
}

func producerCalibrationDataHandler(client mqttgo.Client, renderedMapChan chan []byte, topic string) {
	for img := range renderedMapChan {
		token := client.Publish(topic, 1, true, img)
		token.Wait()
		if token.Error() != nil {
			log.Printf("[MQTT producer] Failed to publish: %v\n", token.Error())
		}
	}
}

func producerAnnounceCalibrationTopic(client mqttgo.Client, cdt string, c *config.MQTTConfig) {
	announceTopic := c.Topics.HaAutoconfPrefix + "/sensor/" + c.Topics.ValetudoIdentifier + "/" + c.Topics.ValetudoPrefix + "_" + c.Topics.ValetudoIdentifier + "_calibration/config"

	js := simplejson.New()
	js.Set("name", c.Topics.ValetudoIdentifier+"_calibration")
	js.Set("unique_id", c.Topics.ValetudoIdentifier+"_calibration")
	js.Set("state_topic", cdt)

	device := simplejson.New()
	device.Set("name", c.Topics.ValetudoIdentifier)
	device.Set("identifiers", []string{c.Topics.ValetudoIdentifier})

	js.Set("device", device)

	announcementData, err := js.MarshalJSON()
	if err != nil {
		panic(err)
	}

	token := client.Publish(announceTopic, 1, false, announcementData)
	token.Wait()
	if token.Error() != nil {
		log.Printf("[MQTT producer] Failed to publish: %v\n", token.Error())
	}
}
