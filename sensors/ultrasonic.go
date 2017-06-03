package sensors

import (
	"github.com/Che4ter/rpi_brain/configuration"
	"github.com/Che4ter/rpi_brain/utilities"
	"github.com/kidoman/embd"
	_ "github.com/kidoman/embd/host/rpi" // This loads the RPi driver
	"github.com/kidoman/embd/sensor/us020"
	"fmt"
	"time"
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

func initUltrasonic(config configuration.Configuration) {
	fmt.Println("init Ultrasonic Sensor")
	fmt.Println("init echo pin on ", config.UltrasonicInPin)

	echoPin, err := embd.NewDigitalPin(config.UltrasonicInPin)
	if err != nil {
		panic(err)
	}

	fmt.Println("init trigger pin on ", config.UltrasonicOutPin)
	triggerPin, err := embd.NewDigitalPin(config.UltrasonicOutPin)
	if err != nil {
		panic(err)
	}

	fmt.Println("rpi delay")
	time.Sleep(2 * time.Second)

	fmt.Println("setup ultrasonic driver")

	UltrasonicSensor.sensor = *us020.New(echoPin, triggerPin, nil)
	defer UltrasonicSensor.sensor.Close()
}
