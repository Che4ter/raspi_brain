package sensors

import (
	"fmt"
	"github.com/Che4ter/rpi_brain/configuration"
	"github.com/kidoman/embd"
	_ "github.com/kidoman/embd/host/rpi" // This loads the RPi driver
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

	initUltrasonic(config)
	go startAccel(config)
	initDirectionToggle(config)
	initStartButton(config)

	time.Sleep(2 * time.Second)
}