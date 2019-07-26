package main

import (
	"errors"
	"fmt"
	i2c "github.com/sconklin/go-i2c"
	device "github.com/sconklin/go-lcd-backpack"
)

/*
  01234567890123456789
 ######################
0#   AI4QR RotorNet   #
1# AZ: 123.4 SP: 123.4#
2#SRC: net INF: ABCDE #
3#                    #
 ######################
*/
const (
	LcdBanner  = "AI4QR RotorNet"
	LcdBannerL = 0
	LcdBannerC = 3
	AzLabel    = "AZ:"
	AzLabelL   = 1
	AzLabelC   = 1
	AzValL     = 1
	AzValC     = 5
	SpLabel    = "SP:"
	SpLabelL   = 1
	SpLabelC   = 11
	SpValL     = 1
	SpValC     = 15
	SrcLabel   = "SRC:"
	SrcLabelL  = 2
	SrcLabelC  = 0
	SrcValL    = 2
	SrcValC    = 5
	InfLabel   = "INF:"
	InfLabelL  = 2
	InfLabelC  = 9
	InfValL    = 2
	InfValC    = 14
	MsgL       = 3
	MsgC       = 0
)

type LcdMsgType int

const (
	LcdMsgAz = iota
	LcdMsgSp
	LcdMsgSrc
	LcdMsgInf
	LcdMsgMsg
)

type LcdMsg struct {
	Type LcdMsgType
	Text string
}

func LcdHandler(quitc <-chan bool, errc chan<- error, msgc <-chan LcdMsg) {

	i2c, err := i2c.NewI2C(0x20, 1)
	if err != nil {
		errc <- err
	}
	defer i2c.Close()
	lcd, err := device.NewLcd(i2c, device.LCD16x2)
	if err != nil {
		errc <- err
	}
	lcd.BacklightOn()
	lcd.Clear()
	lcd.SetPosition(LcdBannerL, LcdBannerC)
	fmt.Fprint(lcd, LcdBanner)
	lcd.SetPosition(AzLabelL, AzLabelC)
	fmt.Fprint(lcd, AzLabel)
	lcd.SetPosition(SpLabelL, SpLabelC)
	fmt.Fprint(lcd, SpLabel)
	lcd.SetPosition(SrcLabelL, SrcLabelC)
	fmt.Fprint(lcd, SrcLabel)
	lcd.SetPosition(InfLabelL, InfLabelC)
	fmt.Fprint(lcd, InfLabel)

	for {
		select {
		case <-quitc:
			log.Info("LcdHandler Quit\n")
			return
		case msg := <-msgc:
			/* Display this message */
			switch msg.Type {
			case LcdMsgAz:
				log.Debug("LcdHandler RX Az\n")
				lcd.SetPosition(AzValL, AzValC)
				fmt.Fprint(lcd, msg.Text[0:4])
			case LcdMsgSp:
				log.Debug("LcdHandler RX Sp\n")
				lcd.SetPosition(SpValL, SpValC)
				fmt.Fprint(lcd, msg.Text[0:4])
			case LcdMsgSrc:
				log.Debug("LcdHandler RX Src\n")
				lcd.SetPosition(SrcValL, SrcValC)
				fmt.Fprint(lcd, msg.Text[0:3])
			case LcdMsgInf:
				log.Debug("LcdHandler RX Inf\n")
				lcd.SetPosition(InfValL, InfValC)
				fmt.Fprint(lcd, msg.Text[0:5])
			case LcdMsgMsg:
				log.Debug("LcdHandler RX Msg\n")
				lcd.SetPosition(MsgL, MsgC)
			default:
				errc <- errors.New("Invalid LCD Msg Type")
			}
		}
	}
}
