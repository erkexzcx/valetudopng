package mqtt

import (
	"log"
	"time"

	mqttgo "github.com/eclipse/paho.mqtt.golang"
	"github.com/erkexzcx/valetudopng/pkg/config"
	"github.com/erkexzcx/valetudopng/pkg/mqtt/decoder"
)

func startConsumer(c *config.MQTTConfig, mapJSONChan chan []byte) {
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
		log.Println("[MQTT consumer] Connected")
		token := client.Subscribe(c.Topics.ValetudoPrefix+"/"+c.Topics.ValetudoIdentifier+"/MapData/map-data", 1, nil)
		token.Wait()
		log.Println("[MQTT consumer] Subscribed to map data topic")
	}

	// On disconnection
	opts.OnConnectionLost = func(client mqttgo.Client, err error) {
		log.Printf("[MQTT consumer] Connection lost: %v", err)
	}

	// Initial connection
	client := mqttgo.NewClient(opts)
	for {
		if token := client.Connect(); token.Wait() && token.Error() != nil {
			log.Printf("[MQTT consumer] Failed to connect: %v. Retrying in 5 seconds...\n", token.Error())
			time.Sleep(5 * time.Second)
		} else {
			break
		}
	}
}

func consumerMapDataReceiveHandler(client mqttgo.Client, msg mqttgo.Message, mapJSONChan chan []byte) {
	payload, err := decoder.Decode(msg.Payload())
	if err != nil {
		log.Println("[MQTT consumer] Failed to process raw data:", err)
		return
	}
	mapJSONChan <- payload
}
