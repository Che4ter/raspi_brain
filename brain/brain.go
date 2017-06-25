package brain

import (
	"fmt"
	"github.com/Che4ter/rpi_brain/arduino"
	"github.com/Che4ter/rpi_brain/configuration"
	"github.com/Che4ter/rpi_brain/sensors"
	"strconv"
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
	DRIVESTRAIGHT_BEFORE_CURVE
	SEARCH_FOR_END
	OBSTACLESTAIR
	DONE
	IDLE
	RESET
	WAITFORBUTTON
)

var states = [...]string{
	"INITIALIZE",
	"START",
	"DRIVESTRAIGHT_ONE",
	"DRIVESTRAIGHT_BEFORE_CURVE",
	"SEARCH_FOR_END",
	"OBSTACLESTAIR",
	"DONE",
	"IDLE",
	"RESET",
	"WAITFORBUTTON"}

type State int

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
	startTime := time.Now()
	stairPosition := 0

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
			direction := sensors.GetDirection()
			fmt.Println("*****************************")
			if direction == 0 {
				fmt.Println("Direction: Right Parcour")
			} else {
				fmt.Println("Direction: Left Parcour")
			}
			fmt.Println("*****************************")
			sendCommandSetDirection(direction)
			sendCommandSetSpeedParcour(brainData.config.SpeedParcour)
			time.Sleep(1 * time.Second)
			sendCommandSetSpeedStair(brainData.config.SpeedStair)
			time.Sleep(1 * time.Second)
			switchState(WAITFORBUTTON)

		case WAITFORBUTTON:
			if sensors.GetButtonStatus() == 0 {

				fmt.Println("Startbutton pressed")
				sendCommandSwitchState(configuration.STATE_START)
				time.Sleep(1 * time.Second)
				switchState(START)
			}
		case START:
			if firstTime {
				firstTime = false
				startTime = time.Now()
			}

			select {
			case datafromcamera := <-brainData.ipcBridge:
				fmt.Println("received ipcv: ", datafromcamera)

				if datafromcamera == "start" {
					fmt.Println("received startsignal: ", datafromcamera)
					switchState(DRIVESTRAIGHT_ONE)
				}
			default:
				timeoutTime := time.Now()
				diffTime := timeoutTime.Sub(startTime)
				if diffTime.Seconds() > brainData.config.WaitForSignalTimeout {
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
			if firstTime {
				firstTime = false
			}

		case IDLE:
		case DONE:
			doneBridge <- true

		case RESET:
			sendCommandReset()
			time.Sleep(4 * time.Second)
			firstTime = false
			switchState(INITIALIZE)
		}
		checkButton()
		checkForNumber()
	}
}

func checkForNumber() {
	if brainData.currentState != START {
		select {
		case datafromcamera := <-brainData.ipcBridge:
			i, _ := strconv.Atoi(datafromcamera)
			if i > 0 && i <= 5 {
				sendCommandSetDigit(i)
			}
		default:
		}
	}
}

func checkButton() {
	if sensors.GetButtonStatus() == 0 && brainData.currentState != WAITFORBUTTON {
		fmt.Println("button detected")
		if brainData.currentState == RESET {
			switchState(INITIALIZE)

		} else {
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

func sendCommandSetDigit(digit int) {
	fmt.Println("Send Digit to Arduino:", digit)
	packet := arduino.ArduinoPacket{
		SOH:    configuration.SOH,
		ID:     configuration.SET_DIGIT,
		TYPE:   configuration.REQUEST,
		LENGTH: 1,
		DATA:   make([]int, 1)}
	packet.DATA[0] = digit
	sendPacket(packet)
}

func sendCommandSetSpeedStair(speed int) {
	fmt.Println("Send Packet to Arduino: set Stair speed to: ", speed)
	packet := arduino.ArduinoPacket{
		SOH:    configuration.SOH,
		ID:     configuration.SET_SPEED_STAIR,
		TYPE:   configuration.REQUEST,
		LENGTH: 1,
		DATA:   make([]int, 1)}
	packet.DATA[0] = speed
	sendPacket(packet)
}

func sendCommandSetSpeedParcour(speed int) {
	fmt.Println("Send Packet to Arduino: set Parcour speed to: ", speed)
	packet := arduino.ArduinoPacket{
		SOH:    configuration.SOH,
		ID:     configuration.SET_SPEED_PARCOUR,
		TYPE:   configuration.REQUEST,
		LENGTH: 1,
		DATA:   make([]int, 1)}
	packet.DATA[0] = speed
	sendPacket(packet)
}

func sendCommandSetDirection(direction int) {
	fmt.Println("Send Packet to Arduino: set Direction to: ", direction)
	packet := arduino.ArduinoPacket{
		SOH:    configuration.SOH,
		ID:     configuration.SET_DIRECTION,
		TYPE:   configuration.REQUEST,
		LENGTH: 1,
		DATA:   make([]int, 1)}
	packet.DATA[0] = direction
	sendPacket(packet)
}

func sendCommandReset() {
	fmt.Println("Send Packet to Arduino: reset")

	packet := arduino.ArduinoPacket{
		SOH:    configuration.SOH,
		ID:     configuration.RESET,
		TYPE:   configuration.REQUEST,
		LENGTH: 0,
		DATA:   make([]int, 0)}
	sendPacket(packet)
}

func sendCommandStop() {
	fmt.Println("Send Packet to Arduino: stop")
	packet := arduino.ArduinoPacket{
		SOH:    configuration.SOH,
		ID:     configuration.STOP,
		TYPE:   configuration.REQUEST,
		LENGTH: 0,
		DATA:   make([]int, 0)}
	sendPacket(packet)
}

func sendPacket(packet arduino.ArduinoPacket) bool {
	brainData.arduinoSendingBridge <- packet

	return true
}
