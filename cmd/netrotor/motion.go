package main

import (
	"errors"
	"fmt"
	"math"
	"time"

	relay "github.com/sconklin/go-dockerpi-relay"
	i2c "github.com/sconklin/go-i2c"
)

/*
 * Modes: There may eventually be three modes. Only SwControl is implemented.
 *
 *   SwControl is the 'normal' mode, where motion control is under sogtware control. The desired azimuth (setpoint) could come from
 *   several sources, including receipt of an N1MM UDP packet.
 *
 *   ManualControl is entered when we detect that the user is moving the rotator using the original controller front panel switch(es).
 *   The only way we can detect this is that the rotor is moving and it's not under our control/
 *   ManualControl exits when the rotator has stopped moving for a period of time
 *
 *   Stuck is entered when we have commanded the rotator to move but the azimuth does not change. This could be due to a stuck brake
 *   or ice on the rotor. An attempt is made to unstick the rotator.
 *
 */

type ControlMode int

const (
	ModeManualControl = iota
	ModeSwControl
	ModeStuck
)

type ControlState int

const (
	StateBraked = iota
	StateUnbraked
	StateMovingCw
	StateMovingCCW
	StateCoasting
)

const (
	BrakeRelay = 1
	CwRelay    = 2
	CcwRelay   = 3
)

const DeadBand = 2.0

// Oddly, the amount of coast if different in each direction
// were both 3.0
const CwCoastBand = 7.9
const CcwCoastBand = 4.6

const (
	Clockwise = iota
	CounterClockwise
)

func within(one, two, delta float64) bool {
	if math.Abs(one-two) <= math.Abs(delta) {
		return true
	} else {
		return false
	}
}

func clampAz(az float64) float64 {

	if az > 180.0 {
		az = az - 360.0
	}
	if az > 180.0 {
		return 180.0
	}
	if az < -180.0 {
		return -180.0
	}
	return az
}

func infoString(mode ControlMode, state ControlState) string {
	var retstr string
	switch mode {
	case ModeManualControl:
		retstr = "M"
	case ModeSwControl:
		retstr = " "
	case ModeStuck:
		retstr = "S"
	}

	switch state {
	case StateBraked:
		retstr += "B"
	case StateUnbraked:
		retstr += "U"
	case StateMovingCw:
		retstr += ">"
	case StateMovingCCW:
		retstr += "<"
	case StateCoasting:
		retstr += "C"
	}
	retstr += "  "
	return retstr
}

func MotionHandler(errc chan<- error, setpointc <-chan Rinfo, lcdc chan<- LcdMsg) {

	// In this module, setpoint and azimuth are kept as a range from -180 to 180 degrees
	// This is because this is the actual motion range of the rotator
	var setpoint float64
	var azimuth float64
	var state ControlState
	var mode ControlMode
	var coastStartTime = time.Now()
	var spReceived bool = false
	var updateInfo = false

	i2c, err := i2c.NewI2C(0x10, 1)
	if err != nil {
		errc <- err
	}
	defer i2c.Close()
	rly, err := relay.NewRelay(i2c)
	if err != nil {
		errc <- err
	}

	// Make sure all relays are off in case we restarted
	for i := uint8(1); i <= uint8(4); i++ {
		err = rly.Off(i)
		if err != nil {
			errc <- err
		}
	}

	// set initial state and mode
	mode = ModeSwControl
	state = StateBraked

	for {
		select {
		case sp := <-setpointc:
			/* we received a new setpoint */
			spReceived = true
			setpoint = clampAz(sp.Azimuth)
			spstr := fmt.Sprintf("%5.1f", sp.Azimuth)
			log.Infof("Motion received setpoint: %s", spstr)
			lcdc <- LcdMsg{LcdMsgSp, spstr}
			lcdc <- LcdMsg{LcdMsgSrc, sp.Source}
		case <-time.After(100 * time.Millisecond):
			break
		}

		admutex.Lock()
		azimuth = azvalue
		admutex.Unlock()
		azimuth = clampAz(azimuth)

		// Now we start the motion control logic.
		switch mode {
		case ModeManualControl:
			errstr := fmt.Sprintf("Unexpected unimplemented manual mode in motion control")
			errc <- errors.New(errstr)
		case ModeSwControl:
			if !spReceived {
				break
			}
			switch state {
			case StateBraked:
				// See if we need to move
				if within(setpoint, azimuth, DeadBand) {
					break // nothing to do
				} else {
					// unbrake and prepare to turn
					rly.On(BrakeRelay) // relay on releases brake
					state = StateUnbraked
					updateInfo = true
				}
			case StateUnbraked:
				if setpoint >= azimuth {
					// Turn Clockwise
					state = StateMovingCw
					rly.On(CwRelay)
					updateInfo = true
				} else {
					// Turn CCW
					state = StateMovingCCW
					rly.On(CcwRelay)
					updateInfo = true
				}
			case StateMovingCw:
				if (setpoint < azimuth) || within(setpoint, azimuth, CwCoastBand) {
					rly.Off(CwRelay)
					state = StateCoasting
					coastStartTime = time.Now()
					updateInfo = true
				}
			case StateMovingCCW:
				if (setpoint > azimuth) || within(setpoint, azimuth, CcwCoastBand) {
					rly.Off(CcwRelay)
					state = StateCoasting
					coastStartTime = time.Now()
					updateInfo = true
				}
			case StateCoasting:
				tdelta := time.Now().Sub(coastStartTime)
				if tdelta > (2 * time.Second) {
					rly.Off(BrakeRelay)
				}
				rly.Off(BrakeRelay)
				state = StateBraked
				spReceived = false // TODO This is only for testing
				updateInfo = true
			default:
				errstr := fmt.Sprintf("Unexpected state %d in motion control", state)
				errc <- errors.New(errstr)
			}
		case ModeStuck:
			errstr := fmt.Sprintf("Unexpected unimplemented stuck mode in motion control")
			errc <- errors.New(errstr)
		default:
			errstr := fmt.Sprintf("Unexpected mode %d in motion control", mode)
			errc <- errors.New(errstr)
		}
		if updateInfo {
			lcdc <- LcdMsg{LcdMsgInf, infoString(mode, state)}
			updateInfo = false
		}
	}
}
