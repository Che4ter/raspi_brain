package sensors

import (
	"fmt"
	"github.com/Che4ter/rpi_brain/configuration"
	"github.com/Che4ter/rpi_brain/sensors/adxl345"
	"github.com/kidoman/embd"
	"time"
)

var _orientation adxl345.Orientation

func GetOrientations() adxl345.Orientation {
	return _orientation
}

func initAccel(config configuration.Configuration) {
	fmt.Println("init Accelerometer Sensor")
	if err := embd.InitI2C(); err != nil {
		panic(err)
	}
	defer embd.CloseI2C()

	bus := embd.NewI2CBus(1)

	acc := adxl345.New(bus, adxl345.RANGE_2_G)
	defer acc.Close()
	acc.Start()

	orientations, err := acc.Orientations()
	if err != nil {
		panic(err)
	}

	timer := time.Tick(250 * time.Millisecond)

	for {
		select {
		case <-timer:
			orientation := <-orientations
			_orientation.Y = orientation.Y
			_orientation.X = orientation.X
			_orientation.Z = orientation.Z
		}
	}

}
