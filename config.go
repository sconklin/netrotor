/*
 * N1MM broadcasts Rotator commands from port 12040
 * Rotator status we send are sent from port 13010
 *
 * https://stackoverflow.com/questions/16465705/how-to-handle-configuration-in-go
 */

package rotor

import (
	"encoding/json"
	"os"
)

type Rotator struct {
	Name      string `json:"name"`
	Port      string `json:"port"`
	PortSpeed string `json:"port_speed"`
	Model     string `json:"model"`
}

type Config struct {
	Rotators     []Rotator `json:"Rotators"`
	Groups       []string  `json:"Groups"`
	AnotherThing string
}

func readConfig(jsonFileName string) (*Config, error) {
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
