package main

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io/ioutil"
	"strconv"

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

func MqttHandler(errc chan<- error, mqttsetc chan<- Rinfo, mqttpubc <-chan Rinfo, conf *config.Config) {

	log.Info("MQTT New")
	mqttClient := client.New(&client.Options{
		ErrorHandler: func(err error) {
			errc <- err
		},
	})

	defer mqttClient.Terminate()

	log.Info("MQTT Cert")
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

	log.Info("MQTT Connect")
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
					// TODO make this use logging or remove it
					// fmt.Println(string(topicName), string(message))
					log.Infof("MQTT RX <%s><%s>", string(topicName), string(message))
					var azimuth float64
					azimuth, _ = strconv.ParseFloat(string(message), 64)
					if azimuthValid(azimuth) {
					} else {
						log.Infof(" MQTT RX Bad Azimuth: <%5.1f>", azimuth)
					}
					mqttsetc <- Rinfo{azimuth, "", "MQT"}
				},
			},
		},
	})
	if err != nil {
		errc <- err
	}

	for {
		select {
		case sp := <-mqttpubc:
			/* Publish the azimuth */
			textinfo := fmt.Sprintf("%5.1f", sp.Azimuth)
			log.Infof(" MQTT Server: publishing %s", textinfo)
			err = mqttClient.Publish(&client.PublishOptions{
				QoS:       mqtt.QoS1,
				TopicName: []byte(conf.MqttI.TopicRead),
				Message:   []byte(textinfo),
			})
			if err != nil {
				errc <- err
			}
			/*
				case <-time.After(100 * time.Millisecond):
					break
			*/
		}

	}
}
