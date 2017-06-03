package sensors

import (
	"github.com/Che4ter/rpi_brain/configuration"
	"fmt"
	"github.com/kidoman/embd"
	_ "github.com/kidoman/embd/host/rpi" // This loads the RPi driver
	"time"
)

type Button struct {
	config     configuration.Configuration
	echoPin    embd.DigitalPin
}

var StartButton Button

func initStartButton(config configuration.Configuration) {
	fmt.Println("init Startbutton")
	fmt.Println("init startbutton pin on ", config.StartButtonPin)

	echoPin, err := embd.NewDigitalPin(config.StartButtonPin)
	if err != nil {
		panic(err)
	}
	StartButton.echoPin = echoPin
	StartButton.echoPin.SetDirection(embd.In)
	time.Sleep(100* time.Millisecond)

	_, err = StartButton.echoPin.Read()
	if err != nil {
		panic(err)
	}
}

func GetButtonStatus() int{
	val, err := StartButton.echoPin.Read()
	if err != nil {
		panic(err)
	}
	return val
}
