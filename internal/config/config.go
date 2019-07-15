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
type Rotator struct {
	Name      string `json:"name"`
	Port      string `json:"port"`
	PortSpeed string `json:"port_speed"`
	Model     string `json:"model"`
}

// Net Represents the network-associated configuration items
type Net struct {
	RotorRx  string `json:"rotorrx"`
	RotorTx  string `json:"rotortx"`
	StatusRx string `json:"statusrx"`
}

// Config Represents the top-level config structure
type Config struct {
	Rotators []Rotator `json:"Rotators"`
	Network  Net       `json:"Network"`
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
	fmt.Printf("Rotators:\n")
	for _, rotator := range conf.Rotators {
		fmt.Printf("    Name:  %s\n", rotator.Name)
		fmt.Printf("    Model: %s\n", rotator.Model)
		fmt.Printf("    Port:  %s\n", rotator.Port)
		fmt.Printf("    Speed: %s\n", rotator.PortSpeed)
	}
	fmt.Printf("Network:\n")
	fmt.Printf("    Rotor Rx:  %s\n", conf.Network.RotorRx)
	fmt.Printf("    Rotor Tx:  %s\n", conf.Network.RotorTx)
	fmt.Printf("    Status Rx: %s\n", conf.Network.StatusRx)
}
