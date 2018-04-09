/*
 * N1MM broadcasts Rotator commands from port 12040
 * Rotator status we send are sent from port 13010
 *
 * https://stackoverflow.com/questions/16465705/how-to-handle-configuration-in-go
 */

package main

import (
	"encoding/json"
	"fmt"
	"net"
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

// A Simple function to verify error
func CheckError(err error) {
	if err != nil {
		fmt.Println("Error: ", err)
		os.Exit(0)
	}
}

func main() {
	config, err := readConfig("multirotorconf.json")

	if err != nil {
		fmt.Printf("readConfig() returned %v\n", err)
		os.Exit(0)
	}

	// if you really want your rotor data in a map with the rotor name as
	// the key, you can do something like the following:

	rotMap := make(map[string]Rotator)

	for _, rot := range config.Rotators {
		rotMap[rot.Name] = rot
	}

	fmt.Printf("rotator map: %#v\n", rotMap)

	// ignore the rest for now
	os.Exit(0)

	// Lets prepare a address to listen from any address sending at port 12040
	ServerAddr, err := net.ResolveUDPAddr("udp", ":12040")
	CheckError(err)

	// Now listen at selected port
	ServerConn, err := net.ListenUDP("udp", ServerAddr)
	CheckError(err)
	defer ServerConn.Close()

	buf := make([]byte, 1024)

	for {
		n, addr, err := ServerConn.ReadFromUDP(buf)
		fmt.Println("Received ", string(buf[0:n]), " from ", addr)

		if err != nil {
			fmt.Println("Error: ", err)
		}
	}
}
