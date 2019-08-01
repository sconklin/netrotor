package main

import (
	"time"

	"github.com/sconklin/go-ads1115"
	"github.com/sconklin/go-i2c"
)

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

	err = sensor.SetPgaMode(ads.PGA_0_256)
	if err != nil {
		errc <- err
	}
	log.Debugf("  Configured for +/- 128 mV Full Scale")

	err = sensor.SetConversionMode(ads.MODE_CONTINUOUS)
	if err != nil {
		errc <- err
	}
	log.Debugf("  Configured for continuous sampling")

	err = sensor.SetDataRate(ads.RATE_8)
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

	for {
		time.Sleep(100 * time.Millisecond)
		val, err := sensor.ReadConversion()
		if err != nil {
			errc <- err
		}
		admutex.Lock()
		advalue = val
		admutex.Unlock()
		log.Infof("A/D value: 0x%x", val)
	}
}
