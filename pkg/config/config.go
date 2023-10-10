package config

import (
	"errors"
	"os"
	"time"

	"gopkg.in/yaml.v2"
)

type MQTTConfig struct {
	Connection    *ConnectionConfig `yaml:"connection"`
	Topics        *TopicsConfig     `yaml:"topics"`
	ImageAsBase64 bool              `yaml:"image_as_base64"`
}

type HTTPConfig struct {
	Enabled bool   `yaml:"enabled"`
	Bind    string `yaml:"bind"`
}

type ConnectionConfig struct {
	Host           string `yaml:"host"`
	Port           string `yaml:"port"`
	Username       string `yaml:"username"`
	Password       string `yaml:"password"`
	ClientIDPrefix string `yaml:"client_id_prefix"`
	TLSCaPath      string `yaml:"tls_ca_path"`
	TLSInsecure    bool   `yaml:"tls_insecure"`
}

type TopicsConfig struct {
	ValetudoPrefix     string `yaml:"valetudo_prefix"`
	ValetudoIdentifier string `yaml:"valetudo_identifier"`
	HaAutoconfPrefix   string `yaml:"ha_autoconf_prefix"`
}

type MapConfig struct {
	MinRefreshInt time.Duration `yaml:"min_refresh_int"`
	Scale         float64       `yaml:"scale"`
	RotationTimes int           `yaml:"rotate"`
	CustomLimits  struct {
		StartX int `yaml:"start_x"`
		StartY int `yaml:"start_y"`
		EndX   int `yaml:"end_x"`
		EndY   int `yaml:"end_y"`
	} `yaml:"custom_limits"`
}

type Config struct {
	Mqtt *MQTTConfig `yaml:"mqtt"`
	HTTP *HTTPConfig `yaml:"http"`
	Map  *MapConfig  `yaml:"map"`
}

func NewConfig(configFile string) (*Config, error) {
	c := &Config{}

	yamlFile, err := os.ReadFile(configFile)
	if err != nil {
		return nil, err
	}

	err = yaml.Unmarshal(yamlFile, c)
	if err != nil {
		return nil, err
	}

	return validate(c)
}

func validate(c *Config) (*Config, error) {
	if c.Mqtt == nil {
		return nil, errors.New("missing mqtt section")
	}
	if c.HTTP == nil {
		return nil, errors.New("missing http section")
	}
	if c.Map == nil {
		return nil, errors.New("missing map section")
	}

	if c.Mqtt.Connection == nil {
		return nil, errors.New("missing mqtt.connection section")
	}
	if c.Mqtt.Topics == nil {
		return nil, errors.New("missing mqtt.topics section")
	}

	// Everything else should fail when used (e.g. wrong IP/port will cause
	// fatal error when starting http server)

	return c, nil
}
