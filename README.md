A networked rotor controller
=============================================================================================

[![Build Status](https://travis-ci.org/sconklin/netrotor.svg?branch=master)](https://travis-ci.org/sconklin/netrotor)
[![Go Report Card](https://goreportcard.com/badge/github.com/sconklin/netrotor)](https://goreportcard.com/report/github.com/sconklin/netrotor)
[![GoDoc](https://godoc.org/github.com/sconklin/netrotor?status.svg)](https://godoc.org/github.com/sconklin/netrotor)
[![MIT License](http://img.shields.io/badge/License-MIT-yellow.svg)](./LICENSE)

This is an Azimuth-only rotator control system that runs on a raspberry pi and allows control of the rotor from the network.
It accepts the N1MM logger rotor control packets, and allows control via MQTT.


To make it start as a service, copy netrotor.service to /etc/systemd/system, then edit that script to point to the gateway executable.

Then run the following:

$ sudo systemctl daemon-reload
$ sudo systemctl enable rotor-n1mm.service

For N1MM, Port 12060 is used for status and 13010 is used for rotor updates

Turn Rotator:
<N1MMRotor>
     <rotor>rotor name</rotor>
     <goazi>55.0</goazi>
     <offset>0.0</offset>
     <bidirectional>0</bidirectional>
     <freqband>14</freqband>   *
</N1MMRotor>

Examples of freqband encoding are 1.8, 3.5, 7, 14, 21, 28

Stop Rotator:
<N1MMRotor>
      <stop>
            <rotor>YaesuCom9</rotor>
            <freqband>21.0</freqband>
      </stop>
</N1MMRotor>

Rotor status update messages sent from the separate N1MM Rotor program on UDP port 13010 are in this format:

rotorname @ rotorheading

They are sent approx every 16 seconds

There are spaces before and after the ‘@’, and the heading is in degrees times ten with no leading zeros, i.e.:
36 degrees - 360
146 degrees - 1460

356 degrees - 3560
