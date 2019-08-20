/*
 * N1MM broadcasts Rotator commands from port 12040
 * Rotator status we send are sent from port 13010
 *
 * https://stackoverflow.com/questions/16465705/how-to-handle-configuration-in-go
 */

package config

import (
	"encoding/json"
	"fmt"
	"os"
)

// Rotator Represents the config items required for a single rotator
type Rot struct {
	SerialEnable string `json:"serial_enable"`
	Name         string `json:"name"`
	Port         string `json:"port"`
	PortSpeed    string `json:"port_speed"`
	Model        string `json:"model"`
}

// Net Represents the network-associated configuration items
type Net struct {
	N1mmEnable string `json:"n1mm_enable"`
	RotorRx    string `json:"rotorrx"`  // N1MM RX port
	RotorTx    string `json:"rotortx"`  // N1MM TX Port
	StatusRx   string `json:"statusrx"` // N1MM Status Port
}

// Mqtt Represents the MQTT items
type Mqtt struct {
	MqttEnable string `json:"mqtt_enable"`
	TopicSet   string `json:"topic_set"`
	TopicRead  string `json:"topic_read"`
	Broker     string `json:"broker"`
	BrokerUser string `json:"broker_user"`
	BrokerPass string `json:"broker_pass"`
	BrokerPort string `json:"broker_port"`
}

// Config Represents the top-level config structure
type Config struct {
	Rotator Rot  `json:"Rotator"`
	Network Net  `json:"Network"`
	MqttI   Mqtt `json:"Mqtt"`
}

// ReadConfig reads the config json file
func ReadConfig(jsonFileName string) (*Config, error) {
	file, err := os.Open(jsonFileName)

	if err != nil {
		return nil, err
	}

	defer file.Close()

	decoder := json.NewDecoder(file)
	config := Config{}

	if err = decoder.Decode(&config); err != nil {
		return nil, err
	}

	return &config, nil
}

// DumpConfig prints the config information
func DumpConfig(conf *Config) {
	fmt.Printf("Rotator:\n")
	fmt.Printf("    Name:  %s\n", conf.Rotator.Name)
	fmt.Printf("    Model: %s\n", conf.Rotator.Model)
	fmt.Printf("    Port:  %s\n", conf.Rotator.Port)
	fmt.Printf("    Speed: %s\n", conf.Rotator.PortSpeed)
	fmt.Printf("Network:\n")
	fmt.Printf("    N1MM Enable:  %s\n", conf.Network.N1mmEnable)
	fmt.Printf("    Rotor Rx:     %s\n", conf.Network.RotorRx)
	fmt.Printf("    Rotor Tx:     %s\n", conf.Network.RotorTx)
	fmt.Printf("    Status Rx:    %s\n", conf.Network.StatusRx)
	fmt.Printf("Mqtt:\n")
	fmt.Printf("    MQTT Enable:  %s\n", conf.MqttI.MqttEnable)
	fmt.Printf("    Topic Set:    %s\n", conf.MqttI.TopicSet)
	fmt.Printf("    Topic Read:   %s\n", conf.MqttI.TopicRead)
	fmt.Printf("    Broker:       %s\n", conf.MqttI.Broker)
	fmt.Printf("    User:         %s\n", conf.MqttI.BrokerUser)
	fmt.Printf("    Pass:         %s\n", conf.MqttI.BrokerPass)
	fmt.Printf("    Port:         %s\n", conf.MqttI.BrokerPort)
}
