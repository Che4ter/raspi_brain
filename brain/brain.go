package brain

import (
	"fmt"
	"github.com/Che4ter/rpi_brain/arduino"
	"github.com/Che4ter/rpi_brain/configuration"
	"github.com/Che4ter/rpi_brain/sensors"
	"time"
)

type Logic struct {
	currentstate int
}

//virutal serial http://www.sagunpandey.com/setup-virtual-serial-ports-using-tty0tty-in-linux/
const (
	INITIALIZE = 1 + iota
	START
	DRIVESTRAIGHT_ONE
	DRIVECURVE
	OBSTACLESTAIR
	OBSTACLEENTANGLEMENT
	OBSTACLECROSSBARS
	OBSTACLERAMP
	PRESSBUTTON
	DONE
	IDLE
)

var states = [...]string{
	"INITIALIZE",
	"START",
	"DRIVESTRAIGHT_ONE",
	"DRIVECURVE",
	"OBSTACLESTAIR",
	"OBSTACLEENTANGLEMENT",
	"OBSTACLERAMP",
	"OBSTACLECROSSBARS",
	"PRESSBUTTON",
	"DONE",
	"IDLE"}

type State int

const resendtimeoutduration = 10
const waitforstarttimeoutduration = 60

// String() function will return the english name
// that we want out constant State be recognized as
func (state State) String() string {
	return states[state-1]
}

type brainStruct struct {
	brainBridge            chan int
	ipcBridge              chan string
	config                 configuration.Configuration
	currentState           State
	arduinoSendingBridge   chan arduino.ArduinoPacket
	arduinoReceivingBridge chan arduino.ArduinoPacket
}

var brainData brainStruct

func StartBrain(brainBridge chan int, ipcBridge chan string, config configuration.Configuration, doneBridge chan bool, arduinoReceivingBridge chan arduino.ArduinoPacket, arduinoSendingBridge chan arduino.ArduinoPacket) {

	brainData.brainBridge = brainBridge
	brainData.ipcBridge = ipcBridge
	brainData.config = config
	brainData.currentState = IDLE
	brainData.arduinoSendingBridge = arduinoSendingBridge
	brainData.arduinoReceivingBridge = arduinoReceivingBridge
	switchState(INITIALIZE)
	//https://doc.getqor.com/plugins/transition.html
	startTime := time.Now()
	// Define initial state
	for {
		switch brainData.currentState {

		case INITIALIZE:
			sensors.InitializeSensors(config)
			orientations := sensors.GetOrientations()
			for orientations.X == 0 && orientations.Y == 0 && orientations.Z == 0 {
				fmt.Println("Wait for Accel Sensor Data")
				orientations = sensors.GetOrientations()
			}

			startTime = time.Now()
			sendCommandSwitchState(configuration.STATE_START)
			switchState(START)
		case START:

			select {
			case datafromcamera := <-brainData.ipcBridge:
				fmt.Println("rtest: ", datafromcamera)

				if datafromcamera == "start" {
					fmt.Println("received startsignal: ", datafromcamera)
					switchState(DRIVESTRAIGHT_ONE)
				}
			default:
				timeoutTime := time.Now()
				diffTime := timeoutTime.Sub(startTime)
				if diffTime.Seconds() > waitforstarttimeoutduration {
					fmt.Println("start signal timeout exceeded")
					fmt.Println("start without traffic light signal...")
					switchState(DRIVESTRAIGHT_ONE)
				}
			}

		case DRIVESTRAIGHT_ONE:
			//todo: send start command
			sendCommandSwitchState(configuration.STATE_DRIVE_STRAIGHT)

			for sensors.GetDistanceFront() > 10 {
				time.Sleep(200)
			}

			switchState(OBSTACLESTAIR)
		case OBSTACLESTAIR:
			//todo: send start command
			sendCommandSwitchState(configuration.STATE_OBSTACLE_STAIR)
			orientations := sensors.GetOrientations()
			fmt.Println("wait for stair")

			for orientations.Y < 30 {
				time.Sleep(100 * time.Millisecond)
				orientations = sensors.GetOrientations()

			}

			fmt.Println("on stair up")

			for orientations.Y > 60 {
				time.Sleep(100 * time.Millisecond)
				orientations = sensors.GetOrientations()

			}
			fmt.Println("on top of the stair")

		case IDLE:
		case DONE:
			doneBridge <- true

		}
		checkForData()
	}
}

func checkForData() {
	select {
	case receivingPacket := <-brainData.arduinoReceivingBridge:
		fmt.Println("received data", receivingPacket.ID)
	default:

	}
}

func switchState(newState State) {

	fmt.Println("Old State:", brainData.currentState)
	brainData.currentState = newState
	fmt.Println("New State:", brainData.currentState)
}

func sendCommandSwitchState(STATEID int) {
	packet := arduino.ArduinoPacket{}
	packet.SOH = configuration.SOH
	packet.ID = configuration.SWITCH_STATE
	packet.TYPE = configuration.REQUEST
	packet.LENGTH = 1
	packet.DATA = make([]int, 1)
	packet.DATA[0] = STATEID
	packet.CHECKSUM = 0
	sendPacket(packet)
}

func sendPacket(packet arduino.ArduinoPacket) bool {
	brainData.arduinoSendingBridge <- packet

	startTime := time.Now()
	var resendCounter = 0
	for {
		select {
		case result := <-brainData.arduinoReceivingBridge:
			fmt.Println("received response")
			if result.ID == packet.ID && result.TYPE == configuration.RESPONSE {
				fmt.Println("response correct")
				return true
			}
		default:
			fmt.Println("wait for timeout")
			timeoutTime := time.Now()
			diffTime := timeoutTime.Sub(startTime)
			if diffTime.Seconds() > resendtimeoutduration {
				fmt.Println("resend")
				brainData.arduinoSendingBridge <- packet
			}
		}
		if resendCounter >= 3 {
			fmt.Println("sending failed")

			return false
		}
		time.Sleep(30)
	}
}
