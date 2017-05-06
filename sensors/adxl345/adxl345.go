package adxl345

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/golang/glog"
	"github.com/kidoman/embd"
	"sync"
	"time"
)

const (
	ADXL345_ADDRESS          = 0x53
	ADXL345_REG_DEVID        = 0x00 // Device ID
	ADXL345_REG_DATAX0       = 0x32 // X-axis data 0 (6 bytes for X/Y/Z)
	ADXL345_REG_POWER_CTL    = 0x2D // Power-saving features control
	ADXL345_REG_DATA_FORMAT  = 0x31
	ADXL345_REG_BW_RATE      = 0x2C
	ADXL345_DATARATE_0_10_HZ = 0x00
	ADXL345_DATARATE_0_20_HZ = 0x01
	ADXL345_DATARATE_0_39_HZ = 0x02
	ADXL345_DATARATE_0_78_HZ = 0x03
	ADXL345_DATARATE_1_56_HZ = 0x04
	ADXL345_DATARATE_3_13_HZ = 0x05
	ADXL345_DATARATE_6_25HZ  = 0x06
	ADXL345_DATARATE_12_5_HZ = 0x07
	ADXL345_DATARATE_25_HZ   = 0x08
	ADXL345_DATARATE_50_HZ   = 0x09
	ADXL345_DATARATE_100_HZ  = 0x0A // (default)
	ADXL345_DATARATE_200_HZ  = 0x0B
	ADXL345_DATARATE_400_HZ  = 0x0C
	ADXL345_DATARATE_800_HZ  = 0x0D
	ADXL345_DATARATE_1600_HZ = 0x0E
	ADXL345_DATARATE_3200_HZ = 0x0F
	ADXL345_RANGE_2_G        = 0x00 // +/-  2g (default)
	ADXL345_RANGE_4_G        = 0x01 // +/-  4g
	ADXL345_RANGE_8_G        = 0x02 // +/-  8g
	ADXL345_RANGE_16_G       = 0x03 // +/- 16g

	pollDelay = 5 //In Microseconds
)

// Range represents a L3GD20 range setting.
type Range struct {
	value byte
}

// The three range settings supported by L3GD20.
var (
	RANGE_2_G  = &Range{value: ADXL345_RANGE_2_G}
	RANGE_4_G  = &Range{value: ADXL345_RANGE_4_G}
	RANGE_8_G  = &Range{value: ADXL345_RANGE_8_G}
	RANGE_16_G = &Range{value: ADXL345_RANGE_16_G}
)

type Orientation struct {
	X, Y, Z int16
}

// ADXL345 represents a ADXL345 accelerometre.
type ADXL345 struct {
	Bus   embd.I2CBus
	Range *Range

	initialized bool
	mu          sync.RWMutex

	orientations chan Orientation
	closing      chan chan struct{}
}

// New creates a new ADXL345 interface. The bus variable controls
// the I2C bus used to communicate with the device.
func New(bus embd.I2CBus, Range *Range) *ADXL345 {
	return &ADXL345{
		Bus:   bus,
		Range: Range,
	}
}

func (d *ADXL345) setup() error {
	d.mu.RLock()
	if d.initialized {
		d.mu.RUnlock()
		return nil
	}
	d.mu.RUnlock()

	d.mu.Lock()
	defer d.mu.Unlock()
	d.orientations = make(chan Orientation)

	data, err := d.Bus.ReadByteFromReg(ADXL345_ADDRESS, ADXL345_REG_DEVID)
	if err != nil {
		return err
	}

	if data == 0xE5 {
		if err := d.Bus.WriteByteToReg(ADXL345_ADDRESS, ADXL345_REG_POWER_CTL, 0x08); err != nil {

			return err
		}
	} else {
		return err
	}

	d.Range = RANGE_2_G
	err = set_range(d)
	if err != nil {
		panic(err)
	}

	d.initialized = true

	return nil
}

func set_range(d *ADXL345) error {

	data, err := d.Bus.ReadByteFromReg(ADXL345_ADDRESS, ADXL345_REG_DATA_FORMAT)
	if err != nil {
		return err
	}
	data = data & 0x0F
	data |= d.Range.value
	data |= 0x08
	if err := d.Bus.WriteByteToReg(ADXL345_ADDRESS, ADXL345_REG_DATA_FORMAT, data); err != nil {

		return err
	}
	return nil
}

// Orientations returns a channel which will have the current orientation reading.
func (d *ADXL345) Orientations() (<-chan Orientation, error) {
	if err := d.setup(); err != nil {
		return nil, err
	}

	return d.orientations, nil
}

// Start starts the data acquisition loop.
func (d *ADXL345) Start() error {
	if err := d.setup(); err != nil {
		return err
	}

	d.closing = make(chan chan struct{})

	go func() {
		var x, y, z int16
		var orientations chan Orientation

		timer := time.Tick(time.Duration(pollDelay * time.Microsecond))

		for {
			select {
			case <-timer:
				dx, dy, dz, err := d.measureOrientation()
				if err != nil {
					glog.Errorf("adxl345: %v", err)
				} else {
					x = dx
					y = dy
					z = dz

					orientations = d.orientations
				}
			case orientations <- Orientation{x, y, z}:
				orientations = nil
			case waitc := <-d.closing:
				waitc <- struct{}{}
				close(d.orientations)
				return
			}

		}
	}()

	return nil
}

func (d *ADXL345) measureOrientation() (int16, int16, int16, error) {
	if err := d.setup(); err != nil {
		fmt.Println("xx1")
		return 0, 0, 0, err
	}
	data := make([]byte, 6)

	if err := d.Bus.ReadFromReg(ADXL345_ADDRESS, ADXL345_REG_DATAX0, data); err != nil {
		fmt.Println("xx2")

		return 0, 0, 0, err
	}

	var x, y, z int16

	buf := bytes.NewReader(data)
	err := binary.Read(buf, binary.LittleEndian, &x)
	if err != nil {
		x = 0
	}
	err = binary.Read(buf, binary.LittleEndian, &y)
	if err != nil {
		y = 0
	}
	err = binary.Read(buf, binary.LittleEndian, &z)
	if err != nil {
		z = 0
	}

	return x, y, z, nil
}

// Stop the data acquisition loop.
func (d *ADXL345) Stop() error {
	if d.closing != nil {
		waitc := make(chan struct{})
		d.closing <- waitc
		<-waitc
		d.closing = nil
	}
	if err := d.Bus.WriteByteToReg(ADXL345_ADDRESS, ADXL345_REG_POWER_CTL, 0x00); err != nil {

		return err
	}
	d.initialized = false
	return nil
}

// Close.
func (d *ADXL345) Close() error {
	return d.Stop()
}
