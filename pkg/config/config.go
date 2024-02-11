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
	TLSEnabled     bool   `yaml:"tls_enabled"`
	TLSMinVersion  string `yaml:"tls_min_version"`
	TLSCaPath      string `yaml:"tls_ca_path"`
	TLSInsecure    bool   `yaml:"tls_insecure"`
}

type TopicsConfig struct {
	ValetudoPrefix     string `yaml:"valetudo_prefix"`
	ValetudoIdentifier string `yaml:"valetudo_identifier"`
	HaAutoconfPrefix   string `yaml:"ha_autoconf_prefix"`
}

type MapConfig struct {
	MinRefreshInt  time.Duration `yaml:"min_refresh_int"`
	PNGCompression int           `yaml:"png_compression"`
	Scale          float64       `yaml:"scale"`
	RotationTimes  int           `yaml:"rotate"`
	CustomLimits   struct {
		StartX int `yaml:"start_x"`
		StartY int `yaml:"start_y"`
		EndX   int `yaml:"end_x"`
		EndY   int `yaml:"end_y"`
	} `yaml:"custom_limits"`
	Colors struct {
		Floor       string   `yaml:"floor"`
		Obstacle    string   `yaml:"obstacle"`
		Path        string   `yaml:"path"`
		NoGoArea    string   `yaml:"no_go_area"`
		VirtualWall string   `yaml:"virtual_wall"`
		Segments    []string `yaml:"segments"`
	} `yaml:"colors"`
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

	c, err = validate(c)
	if err != nil {
		return nil, err
	}

	return setDefaultColors(c)
}

func setDefaultColors(c *Config) (*Config, error) {
	if c.Map.Colors.Floor == "" {
		c.Map.Colors.Floor = "#0076ffff"
	}

	if c.Map.Colors.Obstacle == "" {
		c.Map.Colors.Obstacle = "#5d5d5d"
	}

	if c.Map.Colors.Path == "" {
		c.Map.Colors.Path = "#ffffffff"
	}

	if c.Map.Colors.NoGoArea == "" {
		c.Map.Colors.NoGoArea = "#ff00004a"
	}

	if c.Map.Colors.VirtualWall == "" {
		c.Map.Colors.VirtualWall = "#ff0000bf"
	}

	if len(c.Map.Colors.Segments) < 4 {
		c.Map.Colors.Segments = []string{"#19a1a1ff", "#7ac037ff", "#ff9b57ff", "#f7c841ff"}
	}

	return c, nil
}

func validate(c *Config) (*Config, error) {
	// Check if any section is nil (missing)
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

	// Check MQTT topics section
	if c.Mqtt.Topics.ValetudoIdentifier == "" {
		return nil, errors.New("missing mqtt.topics.valetudo_identifier value")
	}
	if c.Mqtt.Topics.ValetudoPrefix == "" {
		return nil, errors.New("missing mqtt.topics.valetudo_prefix value")
	}
	if c.Mqtt.Topics.HaAutoconfPrefix == "" {
		return nil, errors.New("missing mqtt.topics.ha_autoconf_prefix value")
	}

	// Check map section
	if c.Map.Scale < 1 {
		return nil, errors.New("missing map.scale cannot be lower than 1")
	}
	if c.Map.PNGCompression < 0 || c.Map.PNGCompression > 3 {
		return nil, errors.New("invalid map.png_compression value")
	}

	// Everything else should fail when used (e.g. wrong IP/port will cause
	// fatal error when starting http server)

	return c, nil
}
