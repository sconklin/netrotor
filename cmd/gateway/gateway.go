package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	//	"strconv"
	"strings"
	"sync"
	"time"

	logger "github.com/sconklin/go-logger"
	"github.com/sconklin/rotor-network/internal/config"
)

// Rinfo contains Info about the rotator
type Rinfo struct {
	Azimuth float64
	Name    string
}

// Create needed mutexes and associated data
var admutex = &sync.Mutex{}
var advalue int16

func extractTag(inp, tag string) string {
	bar := strings.Split(strings.Split(inp, "</"+tag+">")[0], "<"+tag+">")
	return bar[len(bar)-1]
}

func main() {
	var verbose = flag.Bool("v", false, "Enable verbose output")
	flag.Parse()

	logger.ChangePackageLogLevel("i2c", logger.InfoLevel)
	logger.ChangePackageLogLevel("ads", logger.InfoLevel)
	logger.ChangePackageLogLevel("lcdbackpack", logger.InfoLevel)

	// Using this
	// https://stackoverflow.com/questions/15715605/multiple-goroutines-listening-on-one-channel
	// Create the channels we'll use
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
	go LcdHandler(errc, lcdc)

	// Start the UDP handler for N1MM protocol
	go N1MMHandler(errc, azimuthc, setpointc, conf)

	// Start A/D handler to read position
	go AdsHandler(errc)

	// Start motion control handler to move the rotator

	// Test Code for LCD
	lcdc <- LcdMsg{LcdMsgAz, "123.4"}
	lcdc <- LcdMsg{LcdMsgSp, "987.6"}
	lcdc <- LcdMsg{LcdMsgSrc, "Net"}
	lcdc <- LcdMsg{LcdMsgInf, "BXR1"}
	for {
		time.Sleep(1 * time.Second)
		admutex.Lock()
		lval := advalue
		admutex.Unlock()
		s := fmt.Sprintf("Val: %d     ", lval)
		lcdc <- LcdMsg{LcdMsgMsg, s}
	}
}
