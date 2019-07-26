package main

import (
	"flag"
	"fmt"
	"os"
	//	"os/exec"
	"path/filepath"
	//	"strconv"
	"strings"
	//	"time"

	"github.com/sconklin/rotor-network/internal/config"
)

// Rinfo contains Info about the rotator
type Rinfo struct {
	Azimuth float64
	Name    string
}

/*
func check(err error, errc <-chan error) {
	if err != nil {
		errc <- err
	}
}
*/

func extractTag(inp, tag string) string {
	bar := strings.Split(strings.Split(inp, "</"+tag+">")[0], "<"+tag+">")
	return bar[len(bar)-1]
}

func main() {
	var verbose = flag.Bool("v", false, "Enable verbose output")
	flag.Parse()

	// Using this
	// https://stackoverflow.com/questions/15715605/multiple-goroutines-listening-on-one-channel
	quitc := make(chan bool)      // signals all threads to return
	errc := make(chan error)      // for passing back errors to main event loop
	lcdc := make(chan LcdMsg)     // Send messages to the LCD
	azimuthc := make(chan Rinfo)  // used to receive position updates
	setpointc := make(chan Rinfo) // used to pass desired position

	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		fmt.Println("Error: ", err)
		os.Exit(1)
	}

	configpath := filepath.Join(dir, "rotorconf.json")
	if err != nil {
		fmt.Println("Error: ", err)
		os.Exit(1)
	}

	conf, err := config.ReadConfig(configpath)
	if err != nil {
		fmt.Println("Error: ", err)
		os.Exit(1)
	}

	if *verbose {
		config.DumpConfig(conf)
	}

	// Start LCD handler to display messages
	go LcdHandler(quitc, errc, lcdc)

	// Start the UDP handler for N1MM protocol
	go N1MMHandler(quitc, errc, azimuthc, setpointc, conf)

	// Start A/D handler to read position

	// Start motion control handler to move the rotator

}