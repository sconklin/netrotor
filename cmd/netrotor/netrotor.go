package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	logger "github.com/sconklin/go-logger"
	"github.com/sconklin/netrotor/internal/config"
)

// Rinfo contains Info about the rotator
type Rinfo struct {
	Azimuth float64
	Name    string
	Source  string
	Command string
}

// Create needed mutexes and associated data
var admutex = &sync.Mutex{}
var azvalue float64

// Global config structure accessible to all
var conf *config.Config

func parseBool(val string) bool {
	return (strings.ToUpper(val) == "TRUE")
}

func main() {
	var verbose = flag.Bool("v", false, "Enable verbose output")
	flag.Parse()

	logger.ChangePackageLogLevel("main", logger.WarnLevel)
	logger.ChangePackageLogLevel("i2c", logger.WarnLevel)
	logger.ChangePackageLogLevel("ads", logger.WarnLevel)
	logger.ChangePackageLogLevel("lcd-backpack", logger.WarnLevel)
	logger.ChangePackageLogLevel("relay", logger.WarnLevel)

	// Using this
	// https://stackoverflow.com/questions/15715605/multiple-goroutines-listening-on-one-channel
	// Create the channels we'll use
	errc := make(chan error)      // for passing back errors to main event loop
	lcdc := make(chan LcdMsg)     // Send messages to the LCD
	azimuthc := make(chan Rinfo)  // used to receive position updates
	setpointc := make(chan Rinfo) // used to pass desired position to motion control

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

	conf, err = config.ReadConfig(configpath)
	if err != nil {
		fmt.Println("Error: ", err)
		os.Exit(1)
	}

	if *verbose {
		config.DumpConfig(conf)
		logger.ChangePackageLogLevel("main", logger.InfoLevel)
		logger.ChangePackageLogLevel("i2c", logger.InfoLevel)
		logger.ChangePackageLogLevel("ads", logger.InfoLevel)
		logger.ChangePackageLogLevel("lcd-backpack", logger.InfoLevel)
		logger.ChangePackageLogLevel("relay", logger.InfoLevel)
	}

	// Start LCD handler to display messages
	go LcdHandler(errc, lcdc)

	// Start the UDP handler for N1MM protocol
	if parseBool(conf.Network.N1mmEnable) {
		go N1MMHandler(errc, azimuthc, setpointc, conf)
	}

	// Start A/D handler to read position
	go AdsHandler(errc, lcdc)

	// Start the UDP handler for N1MM protocol
	if parseBool(conf.MqttI.MqttEnable) {
		go MqttHandler(errc, setpointc, conf)
	}

	// Start motion control handler to move the rotator
	go MotionHandler(errc, setpointc, lcdc)

	// Test Code for LCD
	//lcdc <- LcdMsg{LcdMsgSp, "987.6"}
	lcdc <- LcdMsg{LcdMsgSrc, "   "}
	lcdc <- LcdMsg{LcdMsgInf, "    "}

	for {
		select {
		case myerr := <-errc:
			// We got an error from somewhere
			s := fmt.Sprintf("%v", err)
			lcdc <- LcdMsg{LcdMsgMsg, s}
			log.Fatal(myerr)
		case <-time.After(500 * time.Millisecond):
			// Update the LCD Display
			admutex.Lock()
			laz := azvalue
			admutex.Unlock()
			lcdc <- LcdMsg{LcdMsgAz, fmt.Sprintf("%5.1f", laz)}

			// Send Azimuth to the N1MM UDP handler
			azimuthc <- Rinfo{laz, "", "", "Move"}
		}
	}
}
