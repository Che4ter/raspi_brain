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
	INITIALIZE           = 1 + iota
	START
	DRIVESTRAIGHT_ONE
	DRIVESTRAIGHT_BEFORE_CURVE
	SEARCH_FOR_END
	DRIVECURVE
	OBSTACLESTAIR
	OBSTACLEENTANGLEMENT
	OBSTACLECROSSBARS
	OBSTACLERAMP
	PRESSBUTTON
	DONE
	IDLE
	RESET
)

var states = [...]string{
	"INITIALIZE",
	"START",
	"DRIVESTRAIGHT_ONE",
	"DRIVESTRAIGHT_BEFORE_CURVE",
	"SEARCH_FOR_END",
	"DRIVECURVE",
	"OBSTACLESTAIR",
	"OBSTACLEENTANGLEMENT",
	"OBSTACLERAMP",
	"OBSTACLECROSSBARS",
	"PRESSBUTTON",
	"DONE",
	"IDLE",
	"RESET"}

type State int

const resendtimeoutduration = 30
const waitforstarttimeoutduration = 1

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
var firstTime = false

func StartBrain(brainBridge chan int, ipcBridge chan string, config configuration.Configuration, doneBridge chan bool, arduinoSendingBridge chan arduino.ArduinoPacket, arduinoReceivingBridge chan arduino.ArduinoPacket) {

	brainData.brainBridge = brainBridge
	brainData.ipcBridge = ipcBridge
	brainData.config = config
	brainData.currentState = IDLE
	brainData.arduinoSendingBridge = arduinoSendingBridge
	brainData.arduinoReceivingBridge = arduinoReceivingBridge

	switchState(INITIALIZE)
	//https://doc.getqor.com/plugins/transition.html
	startTime := time.Now()
	stairPosition := 0
	//sendCommandSwitchState(configuration.STATE_DRIVE_CURVE_LEFT)

	// Define initial state
	for {
		switch brainData.currentState {

		case INITIALIZE:
			fmt.Println("Start Initializing")
			sensors.InitializeSensors(config)
			orientations := sensors.GetOrientations()
			for orientations.X == 0 && orientations.Y == 0 && orientations.Z == 0 {
				fmt.Println("Wait for Accel Sensor Data")
				orientations = sensors.GetOrientations()
			}
			sendCommandSwitchState(configuration.STATE_START)
			switchState(START)
		case START:
			if firstTime {
				firstTime = false
				startTime = time.Now()
			}
			select {
			case datafromcamera := <-brainData.ipcBridge:
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
			if firstTime {
				sendCommandSwitchState(configuration.STATE_DRIVE1)
				firstTime = false
			}

			/*if sensors.GetDistanceFront() > 10 {
				time.Sleep(200)
			} else {
				switchState(OBSTACLESTAIR)
			}*/
			switchState(OBSTACLESTAIR)


		case OBSTACLESTAIR:
			if firstTime {
				sendCommandSwitchState(configuration.STATE_OBSTACLE_STAIR)
				fmt.Println("stair...")

				firstTime = false
				stairPosition = 0

			}

			orientations := sensors.GetOrientations()
			//fmt.Println(orientations.X,orientations.Y,orientations.Z)
			if stairPosition == 0 && orientations.Y < -200 && orientations.X > 10 {
				time.Sleep(1 * time.Second)
				switchState(DRIVESTRAIGHT_BEFORE_CURVE)
			}

		case DRIVESTRAIGHT_BEFORE_CURVE:
			if firstTime {
				sendCommandSwitchState(configuration.STATE_BEFORE_CURVE)
				firstTime = false
				//time.Sleep(20 * time.Second)
			}
		case SEARCH_FOR_END:

		case IDLE:
		case DONE:
			doneBridge <- true

		case RESET:
			if firstTime {
				sendCommandStop()
				time.Sleep(3 * time.Second)
				firstTime = false
			} else{
				switchState(INITIALIZE)

			}


		}
		checkForData()
		checkButton()
	}
}

func checkForData() {
	select {
	case receivingPacket := <-brainData.arduinoReceivingBridge:
		fmt.Println("received data", receivingPacket.ID)
	default:

	}
}

func checkButton() {
	if sensors.GetButtonStatus() == 0 {
		fmt.Println("button detected")
		if brainData.currentState == RESET {
			switchState(INITIALIZE)

		} else
		{
		switchState(RESET)

		}
	}
}

func switchState(newState State) {

	fmt.Println("Old State:", brainData.currentState)
	brainData.currentState = newState
	fmt.Println("New State:", brainData.currentState)
	firstTime = true
}

func sendCommandSwitchState(STATEID int) {
	packet := arduino.ArduinoPacket{
		SOH:    configuration.SOH,
		ID:     configuration.SWITCH_STATE,
		TYPE:   configuration.REQUEST,
		LENGTH: 1,
		DATA:   make([]int, 1)}
	packet.DATA[0] = STATEID
	sendPacket(packet)
}

func sendCommandReset() {
	packet := arduino.ArduinoPacket{
		SOH:    configuration.SOH,
		ID:     configuration.RESET,
		TYPE:   configuration.REQUEST,
		LENGTH: 0,
		DATA:   make([]int, 0)}
	sendPacket(packet)
}

func sendCommandStop() {
	packet := arduino.ArduinoPacket{
		SOH:    configuration.SOH,
		ID:     configuration.STOP,
		TYPE:   configuration.REQUEST,
		LENGTH: 0,
		DATA:   make([]int, 0)}
	sendPacket(packet)
}

func sendPacket(packet arduino.ArduinoPacket) bool {
	fmt.Println("try to send packet")
	brainData.arduinoSendingBridge <- packet
	return true

	startTime := time.Now()
	var resendCounter = 0
	for {
		select {
		case result := <-brainData.arduinoReceivingBridge:
			fmt.Println("received response")
			fmt.Println("packet id " ,result.ID)
			fmt.Println("packet type " ,result.TYPE)


			if result.ID == packet.ID && result.TYPE == configuration.RESPONSE {
				fmt.Println("response correct")


				return true
			}
		default:
			//fmt.Println("wait for timeout")
			timeoutTime := time.Now()
			diffTime := timeoutTime.Sub(startTime)
			if diffTime.Seconds() > resendtimeoutduration {
				fmt.Println("resend")
				brainData.arduinoSendingBridge <- packet
				startTime = time.Now()
				resendCounter++
			}
		}
		if resendCounter >= 2 {
			fmt.Println("sending failed")

			return false
		}
		time.Sleep(50 * time.Millisecond)
	}
}
