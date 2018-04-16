package main

import (
	"./config"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func getPosition(rotator config.Rotator) (position string, err error) {

	cmd := fmt.Sprintf("/usr/bin/rotctl -m %s -r %s -s %s \\get_pos", rotator.Model, rotator.Port, rotator.PortSpeed)

	out, err := exec.Command(cmd).Output()
	if err != nil {
		return "", err
	}
	result := string(out)
	azimuth := strings.Split(result, "\n")[0]

	return azimuth, err
}

/*
func setPosition(rotator, position) (error) {
}
*/

func dumpConfig(conf *config.Config) {
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

func main() {

	conf, err := config.ReadConfig("rotorconf.json")

	if err != nil {
		fmt.Println("Error: ", err)
		os.Exit(1)
	}

	dumpConfig(conf)

	for _, rotator := range conf.Rotators {
		position, err := getPosition(rotator)
		if err != nil {
			fmt.Println("Error: ", err)
			os.Exit(1)
		}

		fmt.Printf("Name: %s, Azimuth: %s\n ", rotator.Name, position)
	}
}
