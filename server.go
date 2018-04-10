package rotor

import (
	"fmt"
	"net"
)

/*
 * N1MM broadcasts Rotator commands from port 12040
 * Rotator status we send are sent from port 13010
 * https://stackoverflow.com/questions/16465705/how-to-handle-configuration-in-go
 */

func Server() error {
	/* Lets prepare a address to listen from any address sending at port 12040*/
	ServerAddr, err := net.ResolveUDPAddr("udp", ":12040")

	if err != nil {
		return err
	}

	/* Now listen at selected port */
	ServerConn, err := net.ListenUDP("udp", ServerAddr)

	if err != nil {
		return err
	}

	defer ServerConn.Close()

	buf := make([]byte, 1024)

	for {
		n, addr, err := ServerConn.ReadFromUDP(buf)
		fmt.Println("Received ", string(buf[0:n]), " from ", addr)

		if err != nil {
			fmt.Println("Error: ", err)
		}
	}

	return nil
}
