package main

import (
	"./config"
	"flag"
	"fmt"
	"net"
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

func extractTag(inp, tag string) string {
	bar := strings.Split(strings.Split(inp, "</"+tag+">")[0], "<"+tag+">")
	return bar[len(bar)-1]
}

func main() {
	var verbose = flag.Bool("v", false, "Enable verbose output")
	flag.Parse()

	quit := make(chan bool)
	errc := make(chan error)
	readpos := make(chan Rinfo)
	cmdpos := make(chan Rinfo)
	writepos := make(chan Rinfo)

	conf, err := config.ReadConfig("rotorconf.json")
	if err != nil {
		fmt.Println("Error: ", err)
		os.Exit(1)
	}

	if *verbose {
		config.DumpConfig(conf)
	}

	for _, rotator := range conf.Rotators {
		if *verbose {
			fmt.Printf("Starting reads for rotator %s\n", rotator.Name)
		}
		go func(rotator config.Rotator) {
			/* Read rotor positions and send commands using serial port */
			tLast := time.Now()
			var posLast float64 = 0.0
			var deltap float64 = 0.0
			var azI float64 = 0.0

			for {
				cmdargs := fmt.Sprintf("/usr/bin/rotctl -m %s -r %s -s %s get_pos", rotator.Model, rotator.Port, rotator.PortSpeed)
				out, err := exec.Command("bash", "-c", cmdargs).Output()
				if err != nil {
					if *verbose {
						fmt.Println(err)
					}
					errc <- err
				} else {
					result := string(out)
					azimuth := strings.Split(result, "\n")[0]
					azI, err = strconv.ParseFloat(azimuth, 64)
					if err != nil {
						if *verbose {
							fmt.Println(err)
						}
						errc <- err
					}
					deltap = azI - posLast
					if deltap < 0 {
						deltap = deltap * -1
					}
					if (deltap > 1) || (time.Now().Sub(tLast) > (15 * time.Second)) {
						readpos <- Rinfo{azI, rotator.Name}
						posLast = azI
						tLast = time.Now()
					}
				}
				select {
				case <-errc:
					return
				case <-quit:
					return
				case <-time.After(1 * time.Second):
				case newpos := <-writepos:
					if strings.Compare(newpos.Name, rotator.Name) == 0 {
						/* We received a request to write a position to this rotator */
						cmdargs := fmt.Sprintf("/usr/bin/rotctl -m %s -r %s -s %s set_pos %4.1f 0", rotator.Model, rotator.Port, rotator.PortSpeed, newpos.Azimuth)
						_, err := exec.Command("bash", "-c", cmdargs).Output()
						if err != nil {
							/* we have to ignore this due to a bug in hamlib */
							if err.Error() != "exit status 2" {
								fmt.Println("Error while writing position")
								errc <- err
							}
						}
					}
				}
			}
		}(rotator)
	}

	go func() {
		/* Listen for rotor commands on UDP port*/
		rxport := ":" + conf.Network.RotorRx
		RxAddr, err := net.ResolveUDPAddr("udp", rxport)
		if err != nil {
			fmt.Println(err)
			errc <- err
		}
		RxConn, err := net.ListenUDP("udp", RxAddr)
		if err != nil {
			fmt.Println(err)
			errc <- err
		}
		defer RxConn.Close()
		buf := make([]byte, 1024)
		var azI float64 = 0.0
		for {
			_, _, err := RxConn.ReadFromUDP(buf)
			if err != nil {
				fmt.Println("UDP RX Error: ", err)
				errc <- err
			}
			/* we got a command - parse and send */
			instr := string(buf)
			if *verbose {
				fmt.Println("Pkt Received ", instr)
			}
			/* our parsing is pretty simple */
			if strings.HasPrefix(instr, "<N1MMRotor><rotor>") {
				rotor := extractTag(instr, "rotor")
				azi := extractTag(instr, "goazi")
				offset := extractTag(instr, "offset")
				bi := extractTag(instr, "bidirectional")
				freq := extractTag(instr, "freqband")

				if *verbose {
					fmt.Println("           Rotor: ", rotor)
					fmt.Println("         Azimuth: ", azi)
					fmt.Println("          Offset: ", offset)
					fmt.Println("   Bidirectional: ", bi)
					fmt.Println("        Freqband: ", freq)
				}
				azI, _ = strconv.ParseFloat(azi, 64)
				cmdpos <- Rinfo{azI, rotor}
			}
			select {
			case <-errc:
				return
			case <-quit:
				return
			default:
			}
		}
	}()

	/* TODO Make the netmask a config variable */
	txport := "255.255.255.255:" + conf.Network.RotorTx
	TxAddr, err := net.ResolveUDPAddr("udp", txport)
	if err != nil {
		fmt.Println(err)
		errc <- err
	}
	LocalAddr, err := net.ResolveUDPAddr("udp", ":0")
	if err != nil {
		fmt.Println(err)
		errc <- err
	}

	TxConn, err := net.DialUDP("udp", LocalAddr, TxAddr)
	if err != nil {
		fmt.Println(err)
		errc <- err
	}
	defer TxConn.Close()

	/* main action loop */
	for {
		select {
		case <-errc:
			fmt.Printf("Quitting . . . \n")
			close(quit)
			return
		case p := <-readpos:
			/* Send the UDP packet with rotator position */
			outstr := fmt.Sprintf("%s @ %d", p.Name, int(p.Azimuth*10))
			if *verbose {
				fmt.Printf("Sending UDP: <%s>\n", outstr)
			}
			_, err := TxConn.Write([]byte(outstr))
			if err != nil {
				errc <- err
			}
		case p := <-cmdpos:
			/* validate and send to all attached rotors */
			if *verbose {
				fmt.Printf("Received UDP command: <%s> to %f\n", p.Name, p.Azimuth)
			}
			writepos <- p
		}
	}
}
