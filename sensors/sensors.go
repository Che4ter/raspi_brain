package sensors

import (
	"fmt"
	"github.com/Che4ter/rpi_brain/configuration"
	"github.com/kidoman/embd"
	"github.com/kidoman/embd/sensor/us020"
	"sync"
	"time"
)

var mutex = &sync.Mutex{}

func InitializeSensors(config configuration.Configuration) {
	UltrasonicSensor = Ultrasonic{}
	UltrasonicSensor.config = config

	fmt.Println("init embd")
	if err := embd.InitGPIO(); err != nil {
		panic(err)
	}
	defer embd.CloseGPIO()

	//initUltrasonic(config)
	initAccel(config)

}

func initUltrasonic(config configuration.Configuration) {
	fmt.Print("init Ultrasonic Sensor")
	fmt.Println("init echo pin on ", config.UltrasonicInPin)

	echoPin, err := embd.NewDigitalPin(config.UltrasonicInPin)
	if err != nil {
		panic(err)
	}
	UltrasonicSensor.echoPin = echoPin

	fmt.Println("init trigger pin on ", config.UltrasonicOutPin)
	triggerPin, err := embd.NewDigitalPin(config.UltrasonicOutPin)
	if err != nil {
		panic(err)
	}
	UltrasonicSensor.triggerPin = triggerPin

	fmt.Println("rpi delay")
	time.Sleep(2 * time.Second)

	fmt.Println("setup ultrasonic driver")

	rf := us020.New(echoPin, triggerPin, nil)
	UltrasonicSensor.sensor = *rf
	defer rf.Close()

	fmt.Println("delay")
	time.Sleep(1 * time.Second)
}
