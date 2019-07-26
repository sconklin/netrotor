package main

import (
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/sconklin/rotor-network/internal/config"
)

func N1MMInit() error {
	return nil
}

func N1MMHandler(quitc <-chan bool, errc chan<- error, Azc <-chan Rinfo, Spc chan<- Rinfo, conf *config.Config) {

	var azI float64
	var lastAz float64
	var deltap float64

	buf := make([]byte, 1024)

	timeLast := time.Now()

	// UDP RX setup for rotor commands
	rxport := ":" + conf.Network.RotorRx
	RxAddr, err := net.ResolveUDPAddr("udp", rxport)
	if err != nil {
		errc <- err
	}

	RxConn, err := net.ListenUDP("udp", RxAddr)
	if err != nil {
		errc <- err
	}

	defer RxConn.Close()

	// UDP TX setup
	// TODO Make the netmask a config variable
	txport := "255.255.255.255:" + conf.Network.RotorTx

	TxAddr, err := net.ResolveUDPAddr("udp", txport)
	if err != nil {
		errc <- err
	}

	LocalAddr, err := net.ResolveUDPAddr("udp", ":0")
	if err != nil {
		errc <- err
	}

	TxConn, err := net.DialUDP("udp", LocalAddr, TxAddr)
	if err != nil {
		errc <- err
	}
	defer TxConn.Close()

	// Start a handler loop for receive
	go func() {
		for {
			_, _, err := RxConn.ReadFromUDP(buf)
			if err != nil {
				log.Info("UDP RX Error\n")
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

				log.Debug("           Rotor: ", rotor)
				log.Debug("         Azimuth: ", azi)
				log.Debug("          Offset: ", offset)
				log.Debug("   Bidirectional: ", bi)
				log.Debug("        Freqband: ", freq)

				azI, _ = strconv.ParseFloat(azi, 64)
				Spc <- Rinfo{azI, rotor}
			} else {
				log.Debug("Odd Pkt Received ", instr)
			}
			select {
			case <-quitc:
				log.Info("N1MMHandler RX Loop Quit\n")
				return
			default:
			}
		}
	}()

	// Start a handler loop for transmit
	go func() {
		for {
			select {
			case <-quitc:
				log.Info("N1MMHandler TX Loop Quit\n")
				return
			case p := <-Azc:
				// We got a position report
				azI = p.Azimuth
				deltap = azI - lastAz
				if deltap < 0 {
					deltap = deltap * -1
				}
				// Send position every 15 seconds or when it changes
				if (deltap > 1) || (time.Now().Sub(timeLast) > (15 * time.Second)) {
					lastAz = azI
					timeLast = time.Now()
					outstr := fmt.Sprintf("%s @ %d", conf.Rotator.Name, int(azI*10))
					log.Info("Sending UDP: <%s>\n", outstr)
					_, err := TxConn.Write([]byte(outstr))
					if err != nil {
						errc <- err
					}
				}
			}
		}
	}()
}
