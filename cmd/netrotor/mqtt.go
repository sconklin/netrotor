package main

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"
	"time"

	"github.com/sconklin/netrotor/internal/config"
	"github.com/yosssi/gmq/mqtt"
	"github.com/yosssi/gmq/mqtt/client"
)

func azimuthValid(az float64) bool {
	if (az >= 0.0) && (az <= 360.0) {
		return true
	} else {
		return false
	}
}

func MqttHandler(errc chan<- error, mqttsetc chan<- Rinfo, conf *config.Config) {

	var lastAz float64
	var deltap float64

	timeLast := time.Now()

	log.Debug("MQTT New")
	mqttClient := client.New(&client.Options{
		ErrorHandler: func(err error) {
			errc <- err
		},
	})

	defer mqttClient.Terminate()

	log.Debug("MQTT Cert")
	// Read the certificate file.
	b, err := ioutil.ReadFile("/etc/ssl/certs/ca-certificates.crt")
	if err != nil {
		errc <- err
	}

	roots := x509.NewCertPool()
	if ok := roots.AppendCertsFromPEM(b); !ok {
		errc <- errors.New("failed to parse root certificate")
	}

	tlsConfig := &tls.Config{
		RootCAs: roots,
	}

	ipstr := conf.MqttI.Broker + ":" + conf.MqttI.BrokerPort
	log.Infof(" MQTT Server: %s", ipstr)

	log.Debug("MQTT Connect")
	// Connect to the MQTT Server.
	err = mqttClient.Connect(&client.ConnectOptions{
		Network:   "tcp",
		Address:   ipstr,
		ClientID:  []byte("netrotor"),
		UserName:  []byte(conf.MqttI.BrokerUser),
		Password:  []byte(conf.MqttI.BrokerPass),
		TLSConfig: tlsConfig,
	})
	if err != nil {
		errc <- err
	}

	// Subscribe to two MQTT topic. Accept setpoint commands on one and
	// publish Azimuth on the other

	log.Infof("MQTT Subscribe to <%s>", conf.MqttI.TopicSet)
	err = mqttClient.Subscribe(&client.SubscribeOptions{
		SubReqs: []*client.SubReq{
			&client.SubReq{
				TopicFilter: []byte(conf.MqttI.TopicSet),
				QoS:         mqtt.QoS0,
				// Define the processing of the message handler.
				Handler: func(topicName, message []byte) {
					log.Debugf("MQTT RX <%s><%s>", string(topicName), string(message))
					if strings.ToUpper(string(message)) == "STOP" {
						log.Info(" MQTT RX Stop")
						mqttsetc <- Rinfo{0.0, "", "MQT", "Stop"}
					} else {
						var azimuth float64
						azimuth, _ = strconv.ParseFloat(string(message), 64)
						if azimuthValid(azimuth) {
						} else {
							log.Warnf(" MQTT RX Bad Azimuth: <%5.1f>", azimuth)
						}
						mqttsetc <- Rinfo{azimuth, conf.Rotator.Name, "MQT", "Move"}
					}
				},
			},
		},
	})
	if err != nil {
		errc <- err
	}

	for {
		select {
		case <-time.After(1 * time.Second):
			admutex.Lock()
			azI := azvalue
			admutex.Unlock()

			timeLast = time.Now()
			deltap = azI - lastAz
			if deltap < 0 {
				deltap = deltap * -1
			}

			// Send position every 60 seconds or when it changes
			if (deltap > 2) || (time.Now().Sub(timeLast) > (60 * time.Second)) {
				lastAz = azI
				timeLast = time.Now()
				/* Publish the azimuth */
				textinfo := fmt.Sprintf("%5.1f", azI)
				log.Infof(" MQTT Server: publishing %s", textinfo)
				err = mqttClient.Publish(&client.PublishOptions{
					QoS:       mqtt.QoS1,
					TopicName: []byte(conf.MqttI.TopicRead),
					Message:   []byte(textinfo),
				})
				if err != nil {
					errc <- err
				}
			}
		}

	}
}
