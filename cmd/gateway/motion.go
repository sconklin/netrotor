package main

import (
	"errors"
	"fmt"
	"time"

	relay "github.com/sconklin/go-dockerpi-relay"
	i2c "github.com/sconklin/go-i2c"
)

/*
 * Modes: There are three modes.
 *
 *   ManualControl is entered when we detect that the user is moving the rotator using the original controller front panel switch(es).
 *   The only way we can detect this is that the rotor is moving and it's not under our control/
 *   ManualControl exits when the rotator has stopped moving for a period of time
 *
 *   SwControl is the 'normal' mode, where motion control is under sogtware control. The desired azimuth (setpoint) could come from
 *   several sources, including receipt of an N1MM UDP packet.
 *
 *   Stuck is entered when we have commanded the rotator to move but the azimuth does not change. This could be due to a stuck brake
 *   or ice on the rotor. An attempt is made to unstick the rotator.
 *
 *
 *
 *
 *
 *
 *
 *
 *
 *
 *
 *
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

func MotionHandler(errc chan<- error, setpointc <-chan Rinfo, lcdc chan<- LcdMsg) {

	var setpoint float64
	var mlaz float64
	var state ControlState
	var mode ControlMode

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
			setpoint = sp.Azimuth
			lcdc <- LcdMsg{LcdMsgSp, fmt.Sprintf("%03.1f", setpoint)}
			lcdc <- LcdMsg{LcdMsgSrc, sp.Source}

		default:
		}

		admutex.Lock()
		mlaz = azvalue
		admutex.Unlock()

		time.Sleep(1 * time.Second)

		// Now we start the motion control loop. We need to detect
		// when there is motion now commanded by us (front panel control)
		switch mode {
		case ModeManualControl:
			switch state {
			case StateBraked:
			case StateUnbraked:
			case StateMovingCw:
			case StateMovingCCW:
			case StateCoasting:
			default:
				errstr := fmt.Sprintf("Unexpected state %d in motion control", state)
				errc <- errors.New(errstr)
			}
		case ModeSwControl:
			switch state {
			case StateBraked:
			case StateUnbraked:
			case StateMovingCw:
			case StateMovingCCW:
			case StateCoasting:
			default:
				errstr := fmt.Sprintf("Unexpected state %d in motion control", state)
				errc <- errors.New(errstr)
			}
		case ModeStuck:
			switch state {
			case StateBraked:
				if mlaz > 0 {
					log.Info("mlaz gt zero")

				}
			case StateUnbraked:
			case StateMovingCw:
			case StateMovingCCW:
			case StateCoasting:
			default:
				errstr := fmt.Sprintf("Unexpected state %d in motion control", state)
				errc <- errors.New(errstr)
			}
		default:
			errstr := fmt.Sprintf("Unexpected mode %d in motion control", mode)
			errc <- errors.New(errstr)
		}
	}
}
