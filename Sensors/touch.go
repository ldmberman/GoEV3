package Sensors

import (
	"fmt"
	"time"

	"github.com/ldmberman/GoEV3/utilities"
)

// TouchSensor type.
type TouchSensor struct {
	port        InPort
	isListening bool
	chStop      chan bool
	channels    []chan uint8
}

// FindTouchSensor provides access to a touch sensor at the given port.
func FindTouchSensor(port InPort) *TouchSensor {
	findSensor(port, TypeTouch)

	s := new(TouchSensor)
	s.isListening = false
	s.port = port
	s.chStop = make(chan bool)

	return s
}

func sendEvent(ch chan uint8, val uint8) {
	ch <- val
}

// StartListening starts another go routine that listens for changes in the touch sensor
func (sensor *TouchSensor) StartListening() {
	if sensor.isListening {
		return
	}
	go func() {
		snr := findSensor(sensor.port, TypeTouch)
		path := fmt.Sprintf("%s/%s", baseSensorPath, snr)
		curVal := utilities.ReadUInt8Value(path, "value0")

		for {
			select {
			case <-sensor.chStop:
				return
			default:
				value := utilities.ReadUInt8Value(path, "value0")
				if value != curVal {
					for _, ch := range sensor.channels {
						go sendEvent(ch, value)
					}
					curVal = value
				}
			}
			time.Sleep(time.Millisecond * 50)
		}
	}()
	sensor.isListening = true
}

// StopListening stops listening for changes in the touch sensor
func (sensor *TouchSensor) StopListening() {
	if sensor.isListening {
		sensor.chStop <- true
		sensor.isListening = false
	}
}

// Notify adds the chanel to a list of channels to send Touch data.
func (sensor *TouchSensor) Notify(ch chan uint8) {
	if sensor.indexOf(ch) == -1 {
		sensor.channels = append(sensor.channels, ch)
	}
}

// StopNotify removes the chanel from the list of channels to send Touch data.
func (sensor *TouchSensor) StopNotify(ch chan uint8) {
	if i := sensor.indexOf(ch); i != -1 {
		sensor.channels = append(sensor.channels[:i], sensor.channels[i+1:]...)
	}
}

func (sensor *TouchSensor) indexOf(ch chan uint8) int {
	for i, val := range sensor.channels {
		if val == ch {
			return i
		}
	}
	return -1
}
