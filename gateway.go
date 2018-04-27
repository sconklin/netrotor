package main

import (
	"./config"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

type Rinfo struct {
	Azimuth float64
	Name    string
}

func main() {
	quit := make(chan bool)
	errc := make(chan error)
	pos := make(chan Rinfo)

	conf, err := config.ReadConfig("rotorconf.json")

	if err != nil {
		fmt.Println("Error: ", err)
		os.Exit(1)
	}

	/*config.DumpConfig(conf)*/

	for _, rotator := range conf.Rotators {
		fmt.Print("For each rotator\n")
		go func(rotator config.Rotator) {
			tLast := time.Now()
			var posLast float64 = 0.0
			var deltap float64 = 0.0
			var azI float64 = 0.0

			for {
				cmdargs := fmt.Sprintf("/usr/bin/rotctl -m %s -r %s -s %s get_pos", rotator.Model, rotator.Port, rotator.PortSpeed)
				out, err := exec.Command("bash", "-c", cmdargs).Output()
				if err != nil {
					fmt.Println(err)
					errc <- err
				} else {
					result := string(out)
					azimuth := strings.Split(result, "\n")[0]
					azI, err = strconv.ParseFloat(azimuth, 64)
					if err != nil {
						errc <- err
					}
					deltap = azI - posLast
					if deltap < 0 {
						deltap = deltap * -1
					}

					if (deltap > 1) || (time.Now().Sub(tLast) > (15 * time.Second)) {
						pos <- Rinfo{azI, rotator.Name}
						posLast = azI
						tLast = time.Now()
					}
				}
				select {
				case <-errc:
					return
				case <-quit:
					return
				default:
				}
				time.Sleep(1000 * time.Millisecond)
			}
		}(rotator)
	}

	for {
		select {
		case <-errc:
			fmt.Printf("Quitting . . . \n")
			close(quit)
			return
		case p := <-pos:
			fmt.Printf("MAIN LOOP Rotor <%s> Position: %f\n", p.Name, p.Azimuth)
			/* Send the UDP packet with rotator position */
			/* first see if we have a conflicting name with a received value? */
			/*default: */
		}
	}
}
