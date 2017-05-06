package sensors

import (
	"github.com/Che4ter/rpi_brain/configuration"
	"github.com/Che4ter/rpi_brain/utilities"
	"github.com/kidoman/embd"
	_ "github.com/kidoman/embd/host/rpi" // This loads the RPi driver
	"github.com/kidoman/embd/sensor/us020"
)

type Ultrasonic struct {
	config     configuration.Configuration
	echoPin    embd.DigitalPin
	triggerPin embd.DigitalPin
	sensor     us020.US020
}

var UltrasonicSensor Ultrasonic

func GetDistanceFront() float64 {
	mutex.Lock()
	distance, err := UltrasonicSensor.sensor.Distance()
	if err != nil {
		panic(err)
	}
	mutex.Unlock()
	return utilities.Round(float64(distance), 1)
}
