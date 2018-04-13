package main
import (
	"./config"
    "os"
  	"os/exec"
    "fmt"
	"strings"
)

func getPosition(rotator config.Rotator) (position string, err error) {

	cmd := fmt.Sprintf("/usr/bin/rotctl -m %s -r %s -s %s \\get_pos", rotator.Model, rotator.Port, rotator.PortSpeed) 

	out, err := exec.Command(cmd).Output()
	if err != nil {
		return "", err
	}
	result := string(out)
	azimuth := strings.Split(result,"\n")[0]

	return result, err
}

/*
func setPosition(rotator, position) (error) {
}
*/

func main() {

	config, err := config.ReadConfig("rotorconf.json")

	if err != nil {
		fmt.Println("Error: ", err)
		os.Exit(1)
	}

	fmt.Printf("Config:\n %v\n ", config.Rotators)

	for _, rotator := range config.Rotators {
		position, err := getPosition(rotator)
			if err != nil {
			fmt.Println("Error: ", err)
			os.Exit(1)
		}
		
		fmt.Printf("Name: %s, Azimuth: %s\n ", rotator.Name, position)
	}
}
