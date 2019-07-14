package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"rotor-network/internal/config"
)

func main() {
	var verbose = flag.Bool("v", false, "Enable verbose output")
	var azimuth = flag.Float64("a", -1, "Azimuth to turn to")
	var name = flag.String("n", "", "Name of rotator")

	flag.Parse()

	conf, err := config.ReadConfig("rotorconf.json")
	if err != nil {
		fmt.Println("Error: ", err)
		os.Exit(1)
	}

	if *verbose {
		config.DumpConfig(conf)
	}

	if *name == "" {
		log.Fatal("Must specify a rotor name\n")
	}

	if (*azimuth < 0.0) || (*azimuth > 360.0) {
		log.Fatal("Must specify an azimuth between 0 and 360 degrees\n")
	}

	/* TODO Make the netmask a config variable */
	txport := "255.255.255.255:" + conf.Network.RotorRx
	TxAddr, err := net.ResolveUDPAddr("udp", txport)
	if err != nil {
		log.Fatal(err)
	}
	LocalAddr, err := net.ResolveUDPAddr("udp", ":0")
	if err != nil {
		log.Fatal(err)
	}

	TxConn, err := net.DialUDP("udp", LocalAddr, TxAddr)
	if err != nil {
		log.Fatal(err)
	}
	defer TxConn.Close()

	outstr := fmt.Sprintf("<N1MMRotor><rotor>%s</rotor><goazi>%2.1f</goazi><offset>0.0</offset><bidirectional>0</bidirectional><freqband>14</freqband></N1MMRotor>", *name, *azimuth)

	if *verbose {
		fmt.Printf("Sending UDP: \n%s>\n", outstr)
	}
	_, err = TxConn.Write([]byte(outstr))
	if err != nil {
		log.Fatal(err)
	}
}
