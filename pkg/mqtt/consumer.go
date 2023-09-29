package mqtt

import (
	"crypto/tls"
	"crypto/x509"
	"log"
	"os"
	"time"

	mqttgo "github.com/eclipse/paho.mqtt.golang"
	"github.com/erkexzcx/valetudopng/pkg/config"
	"github.com/erkexzcx/valetudopng/pkg/mqtt/decoder"
)

func startConsumer(c *config.MQTTConfig, mapJSONChan chan []byte) {
	opts := mqttgo.NewClientOptions()
	opts.AddBroker("tcp://" + c.Connection.Host + ":" + c.Connection.Port)
	opts.SetClientID(c.Connection.ClientIDPrefix + "_consumer")
	opts.SetUsername(c.Connection.Username)
	opts.SetPassword(c.Connection.Password)
	opts.SetAutoReconnect(true)

	tlsConfig := &tls.Config{
		InsecureSkipVerify: c.Connection.TLSInsecure,
	}
	if c.Connection.TLSCaPath != "" {
		caCert, err := os.ReadFile(c.Connection.TLSCaPath)
		if err != nil {
			log.Fatalln(err)
		}
		caCertPool := x509.NewCertPool()
		caCertPool.AppendCertsFromPEM(caCert)
		tlsConfig.RootCAs = caCertPool
	}
	opts.SetTLSConfig(tlsConfig)

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
