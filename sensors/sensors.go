package sensors

import (
	"fmt"
	"github.com/Che4ter/rpi_brain/configuration"
	"github.com/kidoman/embd"
	"sync"
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

	initUltrasonic(config)
	initAccel(config)
}
