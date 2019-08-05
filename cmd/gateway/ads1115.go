package main

import (
	"math"
	"time"

	"github.com/sconklin/go-ads1115"
	"github.com/sconklin/go-i2c"
)

// AdToAz converts the A/D voltage reading to azimuth
// This is done now using a linear fit from meeasuring
// the rotor, but should probably be replaced with
// interpolation using values gathered during a cal routine.
func AdToAz(adval int16) float64 {
	rawAz := float64(adval-100) / 79.87
	// Raw azimuth is 0 = south
	return math.Mod((rawAz + 180), 360)
}

func AdsHandler(errc chan<- error) {
	i2c, err := i2c.NewI2C(0x48, 1)
	if err != nil {
		errc <- err
	}
	defer i2c.Close()

	sensor, err := ads.NewADS(ads.ADS1115, i2c)

	if err != nil {
		errc <- err
	}

	err = sensor.SetMuxMode(ads.MUX_SINGLE_0)
	if err != nil {
		errc <- err
	}
	log.Debugf("  Configured for Single Ended Channel 0")

	err = sensor.SetPgaMode(ads.PGA_4_096)
	if err != nil {
		errc <- err
	}
	log.Debugf("  Configured for +/- 4.096 V Full Scale")

	err = sensor.SetConversionMode(ads.MODE_CONTINUOUS)
	if err != nil {
		errc <- err
	}
	log.Debugf("  Configured for continuous sampling")

	err = sensor.SetDataRate(ads.RATE_16)
	if err != nil {
		errc <- err
	}
	log.Debugf("  Configured for 8 Samples per Second")

	err = sensor.SetComparatorMode(ads.COMP_MODE_TRADITIONAL)
	if err != nil {
		errc <- err
	}
	log.Debugf("  Configured for traditional comparator mode")

	err = sensor.SetComparatorPolarity(ads.COMP_POL_ACTIVE_LOW)
	if err != nil {
		errc <- err
	}
	log.Debugf("  Configured comparator active low")

	err = sensor.SetComparatorLatch(ads.COMP_LAT_OFF)
	if err != nil {
		errc <- err
	}
	log.Debugf("  Configured comparator latch off")

	err = sensor.SetComparatorQueue(ads.COMP_QUE_DISABLE)
	if err != nil {
		errc <- err
	}
	log.Debugf("  Configured comparator queue disabled")

	err = sensor.WriteConfig()
	if err != nil {
		errc <- err
	}
	log.Debugf("  Wrote new Config to A/D")

	config, err := sensor.ReadConfig()
	if err != nil {
		errc <- err
	}
	log.Debugf("This A/D has final config: 0x%x", config)

	const loopcount = 10
	var vals [loopcount]int16
	var vi uint16
	// populate array
	for vi = 0; vi < loopcount; vi++ {
		val, err := sensor.ReadConversion()
		if err != nil {
			errc <- err
		}
		vals[vi] = val
	}
	for {
		time.Sleep(100 * time.Millisecond)
		val, err := sensor.ReadConversion()
		if err != nil {
			errc <- err
		}
		vals[vi] = val
		vi = vi + 1
		if vi >= loopcount {
			vi = 0
		}

		// average the readings
		var sum int32
		for i := 0; i < loopcount; i++ {
			sum = sum + int32(vals[i])
		}
		sum = sum / loopcount
		val = int16(sum)

		adfl := AdToAz(val)
		admutex.Lock()
		azvalue = adfl
		admutex.Unlock()
		log.Debugf("Azimuth value: %05d  , %03.3f", val, azvalue)
	}
}
