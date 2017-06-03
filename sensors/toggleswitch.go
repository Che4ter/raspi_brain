package sensors

import (
	"github.com/Che4ter/rpi_brain/configuration"
	"fmt"
	"github.com/kidoman/embd"
	_ "github.com/kidoman/embd/host/rpi" // This loads the RPi driver
	"time"
)

type Toggle struct {
	echoPin embd.DigitalPin
}

var DirectionToggle Toggle

func initDirectionToggle(config configuration.Configuration) {
	fmt.Println("init Directiontoggle")
	fmt.Println("init direction pin on ", config.ToggleSwitchPin)

	DirectionToggle.echoPin, _ = embd.NewDigitalPin(config.ToggleSwitchPin)

	DirectionToggle.echoPin.SetDirection(embd.In)
	DirectionToggle.echoPin.ActiveLow(true)
	time.Sleep(100 * time.Millisecond)
}

func GetDirection() int {
	val, err := DirectionToggle.echoPin.Read()
	if err != nil {
		panic(err)
	}
	return val
}
