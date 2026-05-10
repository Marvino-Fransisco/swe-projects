package lib

import (
	"log"
	"os"

	"gopkg.in/yaml.v3"
)

type ServiceConfig struct {
	Host             string `yaml:"host"`
	Port             int    `yaml:"port"`
	Timeout          int    `yaml:"timeout"`
	Threshold        int    `yaml:"threshold"`
	MaxConnections   int    `yaml:"max_connections"`
	QueueTimeout     int    `yaml:"queue_timeout"`
	RequestDeadline  int    `yaml:"request_deadline"`
}

type ConfigStruct struct {
	Services map[string]ServiceConfig `yaml:"services"`
}

var Config ConfigStruct

func init() {
	data, err := os.ReadFile("config.yaml")
	if err != nil {
		Config = ConfigStruct{
			Services: make(map[string]ServiceConfig),
		}
		return
	}

	if err := yaml.Unmarshal(data, &Config); err != nil {
		log.Fatal("Failed to parse config.yaml:", err)
	}
}
