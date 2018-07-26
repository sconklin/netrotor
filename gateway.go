package main

import (
	"./config"
	"flag"
	"fmt"
	"github.com/tarm/serial"
	"io"
	"net"
	"os"
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
			buf := make([]byte, 128)

			speed, err := strconv.Atoi(rotator.PortSpeed)
			if err != nil {
				if *verbose {
					fmt.Println(err)
				}
				errc <- err
			}

			c := &serial.Config{Name: rotator.Port, Baud: speed, ReadTimeout: time.Second * 1}
			s, err := serial.OpenPort(c)
			if err != nil {
				if *verbose {
					fmt.Println(err)
				}
				errc <- err
			}
			defer s.Close()

			/*
			 * "AM1;" - Start Turning
			 * "AP1xxx;" - set azimuth, does not start turning
			 * "AI1;" - send me your current position
			 * ";" - stop
			 */
			for {
				/* Write command then read position */
				n, err := s.Write([]byte("AI1;"))
				if err != nil {
					if *verbose {
						fmt.Println(err)
					}
					errc <- err
				}

				time.Sleep(250 * time.Millisecond)
				n, err = s.Read(buf)

				if err != nil {
					if *verbose {
						fmt.Println("Error on Serial Read 1")
						fmt.Println(err)
					}
					errc <- err
				} else {
					result := string(buf[:n])

					if strings.HasPrefix(result, ";") && (len(result) == 4) {
						azint, err := strconv.Atoi(strings.TrimLeft(result, ";"))
						if err != nil {
							if *verbose {
								fmt.Println(err)
							}
							errc <- err
						}

						azI = float64(azint)
						deltap = azI - posLast
						if deltap < 0 {
							deltap = deltap * -1
						}
						if (deltap > 1) || (time.Now().Sub(tLast) > (15 * time.Second)) {
							readpos <- Rinfo{azI, rotator.Name}
							posLast = azI
							tLast = time.Now()
						}
					} else {
						if *verbose {
							fmt.Printf("Ignoring rotor response of <%s>", result)
						}
					}

				}
				select {
				case <-errc:
					fmt.Printf("got an errc\n")
					return
				case <-quit:
					return
				case <-time.After(1 * time.Second):
				case newpos := <-writepos:
					if strings.Compare(newpos.Name, rotator.Name) == 0 {
						fmt.Printf("Here's where we send a position\n")
						if newpos.Azimuth > 0.0 && newpos.Azimuth < 360.0 {
							/* it's good! */
							sendaz := int(newpos.Azimuth + 0.5)
							if sendaz == 360 {
								sendaz = 0
							}

							cmd := fmt.Sprintf("AP1%03d;", sendaz)
							if *verbose {
								fmt.Printf("Sending command <%s>\n", cmd)
							}
							n, err := s.Write([]byte(cmd))
							if err != nil {
								if *verbose {
									fmt.Println(err)
								}
								errc <- err
							}
							if n == 0 {
								fmt.Printf("Serial write returned zero length\n")
							}

							cmd = fmt.Sprintf("AM1;")
							if *verbose {
								fmt.Printf("Sending command <%s>\n", cmd)
							}
							n, err = s.Write([]byte(cmd))
							if err != nil {
								if *verbose {
									fmt.Println(err)
								}
								errc <- err
							}
							if n == 0 {
								fmt.Printf("Serial write returned zero length\n")
							}

							fmt.Printf("Done sending\n")
						} else if newpos.Azimuth < 0.0 {
							/* negative values mean stop */
							n, err := s.Write([]byte(";"))
							if err != nil {
								if *verbose {
									fmt.Println(err)
								}
								errc <- err
							}
							if n == 0 {
								fmt.Printf("Serial write returned zero length (2)\n")
							}
						} else {
							if *verbose {
								fmt.Printf("Ignoring invalid azimuth %f\n", newpos.Azimuth)
							}
						}

						/* These can generate responses we don't care about - especially if you send a stop while the rotor is not turning */
						n, err = s.Read(buf)
						if err != nil && err != io.EOF {
							if *verbose {
								fmt.Println("Error on Serial Read 2")
								fmt.Println(err)
							}
							errc <- err
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
			} else {
				if *verbose {
					fmt.Println("Odd Pkt Received ", instr)
				}
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
