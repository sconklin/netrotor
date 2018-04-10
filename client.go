package rotor

import (
	"fmt"
	"net"
	"time"
)

/*
 * N1MM broadcasts Rotator commands from port 12040
 * Rotator status we send are sent from port 13010
 * https://stackoverflow.com/questions/16465705/how-to-handle-configuration-in-go
 */

func Client() error {
	ServerAddr, err := net.ResolveUDPAddr("udp", "255.255.255.255:13010")

	if err != nil {
		return err
	}

	LocalAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:")

	if err != nil {
		return err
	}

	Conn, err := net.DialUDP("udp", LocalAddr, ServerAddr)

	if err != nil {
		return err
	}

	defer Conn.Close()
	i := 0
	for {
		msg := fmt.Sprintf("hexbeam-rotor@%d", i)
		i += 2
		if i > 365 {
			i = 0
		}
		buf := []byte(msg)
		_, err := Conn.Write(buf)
		if err != nil {
			fmt.Println(msg, err)
		}
		time.Sleep(time.Second * 1)
	}

	return nil
}
