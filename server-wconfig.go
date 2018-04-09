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

type Config struct {
	Rotators []string `json:"Users"`
	Groups   []string `json:"Groups"`
}

func readConfig(jsonFileName string) (*Config, error) {
	file, err := os.Open(jsonFileName)

	if err != nil {
		return nil, err
	}

	defer file.Close()

	decoder := json.NewDecoder(file)
	config := Config{}
	err = decoder.Decode(&config)

	if err != nil {
		return nil, err
	}

	return &config, nil
}

/* A Simple function to verify error */
func CheckError(err error) {
	if err != nil {
		fmt.Println("Error: ", err)
		os.Exit(0)
	}
}

func main() {
	config, err := readConfig("rotorconf.json")

	if err != nil {
		fmt.Printf("readConfig() returned %v\n", err)
		os.Exit(0)
	}

	for _, re := range config.Rotators {
		// TODO if we have restirctions like no spaces in names, enforce it here
		fmt.Printf("Rotor name: %s\n", re)
		// Here I'll read the rest of the config parts, and put them in the rotorData map for use in main()
	}

	fmt.Printf("Groups: %#v\n", config.Groups)

	// ignore the rest for now
	os.Exit(0)

	/* Lets prepare a address to listen from any address sending at port 12040*/
	ServerAddr, err := net.ResolveUDPAddr("udp", ":12040")
	CheckError(err)

	/* Now listen at selected port */
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
