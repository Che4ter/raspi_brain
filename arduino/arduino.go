package arduino

import (
	"github.com/Che4ter/rpi_brain/configuration"
	"github.com/tarm/serial"
	"io/ioutil"
	"log"
	"strings"
	"time"
)

type Arduino struct {
	arduinoReceivingBridge chan ArduinoPacket
	arduinoSendingBridge   chan ArduinoPacket
	config                 configuration.Configuration
	serial                 *serial.Port
}

type ArduinoPacket struct {
	SOH      int
	ID       int
	TYPE     int
	LENGTH   int
	DATA     []int
}

const timeoutduration = 60

var arduinoSerial Arduino

// findArduino looks for the file that represents the arduino serial connection. Returns the fully qualified path
// to the device if we are able to find a likely candidate for an arduino, otherwise an empty string if unable to
// find an arduino device.
func findArduino() string {
	contents, _ := ioutil.ReadDir("/dev")

	// Look for the arduino device
	for _, f := range contents {
		if strings.Contains(f.Name(), "tty.usbserial") ||
			strings.Contains(f.Name(), "ttyUSB") || strings.Contains(f.Name(), "ttyACM0") {
			return "/dev/" + f.Name()
		}
	}

	// Have not been able to find the device.
	return ""
}

func RunArduinoServer(arduinoSendingBridge chan ArduinoPacket, arduinoReceivingBridge chan ArduinoPacket, config configuration.Configuration) {
	// Find the device that represents the arduino serial connection.
	arduinoport := ""
	if config.SerialPortAutoDetect {
		arduinoport = findArduino()
	}

	if arduinoport == "" {
		arduinoport = config.SerialPort
	}

	c := &serial.Config{Name: arduinoport, Baud: config.SerialBaudRate}
	s, err := serial.OpenPort(c)
	if err != nil {
		log.Fatal(err)
	}

	// When connecting to an older revision arduino, you need to wait a little while it resets.
	time.Sleep(1 * time.Second)

	arduinoSerial = Arduino{arduinoReceivingBridge, arduinoSendingBridge, config, s}

	go read()

	for {
		arduinoPacket := <-arduinoSerial.arduinoSendingBridge
		sendPacket(arduinoPacket)
	}
	//Write(arduinoPackage,[]byte{0x00, 0x01, 0x02, 0x03})
}

func read() {
	var curLocation = 0
	var receivingPacket = ArduinoPacket{}
	for {
		buf := make([]byte, 1)

		_, err := arduinoSerial.serial.Read(buf)

		if err != nil {
			log.Fatal(err)
		}

		if buf[0] == configuration.SOH && curLocation == 0 {
			receivingPacket.SOH = int(buf[0])
			curLocation++

		} else if curLocation == 1 {
			receivingPacket.ID = int(buf[0])
			curLocation++
		} else if curLocation == 2 {
			receivingPacket.TYPE = int(buf[0])
			curLocation++
		} else if curLocation == 3 {
			receivingPacket.LENGTH = int(buf[0])
			receivingPacket.DATA = make([]int, receivingPacket.LENGTH)
			curLocation++

		} else if curLocation >= 4 && curLocation < (4+receivingPacket.LENGTH) {
			receivingPacket.DATA[curLocation-4] = int(buf[0])
			curLocation++
		} else if curLocation == (4 + receivingPacket.LENGTH) {
			arduinoSerial.arduinoReceivingBridge <- receivingPacket

			curLocation = 0
			receivingPacket = ArduinoPacket{}
		}
	}
}

func verifyChecksum(packet ArduinoPacket) bool {
	return true
}

func write(b []byte) {
	//n, err := s.Write([]byte("test"))
	n, err := arduinoSerial.serial.Write(b)

	if err != nil {
		log.Fatal(err)
	}
	log.Printf("%q", n)
}

func sendPacket(packet ArduinoPacket) {
	data := make([]int, 5+packet.LENGTH)
	data[0] = packet.SOH
	data[1] = packet.ID
	data[2] = packet.TYPE
	data[3] = packet.LENGTH

	for i := 0; i < int(packet.LENGTH); i++ {
		data[i+4] = packet.DATA[i]
	}

	bs := make([]byte, len(data))

	for i := 0; i < int(len(data)); i++ {
		bs[i] = byte(data[i])
	}

	write(bs)
}
