package main

import (
	"fmt"
	"net"
	"os"
)

/*
 * N1MM broadcasts Rotator commands to port 12040
 * Rotator status is sent to port 13010
 */

func main() {
	ServerAddr, err := net.ResolveUDPAddr("udp", ":13010")
	if err != nil {
		os.Exit(1)
	}

	ServerConn, err := net.ListenUDP("udp", ServerAddr)
	if err != nil {
		os.Exit(1)
	}

	defer ServerConn.Close()

	buf := make([]byte, 1024)

	fmt.Println("Entering read loop")
	for {
		n, addr, err := ServerConn.ReadFromUDP(buf)
		fmt.Println("Received ", string(buf[0:n]), " from ", addr)

		if err != nil {
			fmt.Println("Error: ", err)
		}
	}
}
