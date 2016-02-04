// Package Motor Provides APIs for interacting with EV3's motors.
package Motor

import (
	"log"
	"os"
	"path"

	"github.com/ldmberman/GoEV3/utilities"
)

// OutPort Constants for output ports.
type OutPort string

// RunCommand Constants for motor commands
type RunCommand string

// StopMode Constants for how the motor will stop
type StopMode string

// OutPort Constants for output ports.
const (
	OutPortA OutPort = "A"
	OutPortB         = "B"
	OutPortC         = "C"
	OutPortD         = "D"
)

const (
	runForeverCommand  RunCommand = "run-forever"
	runToAbsPosCommand            = "run-to-abs-pos"
	runToRelPosCommand            = "run-to-rel-pos"
	runTimedCommand               = "run-timed"
	runDirectCommand              = "run-direct"
)

// Stop Modes
const (
	Coast StopMode = "coast"
	Brake          = "brake"
	Hold           = "hold"
)

// Names of files which constitute the low-level motor API
const (
	rootMotorPath = "/sys/class/tacho-motor"
	// File descriptors for getting/setting parameters
	portFD            = "port_name"
	regulationModeFD  = "speed_regulation"
	speedGetterFD     = "speed"
	speedSetterFD     = "speed_sp"
	powerGetterFD     = "duty_cycle"
	powerSetterFD     = "duty_cycle_sp"
	runFD             = "command"
	stopModeFD        = "stop_command"
	positionFD        = "position"
	desiredPositionFD = "position_sp"
	timeFD            = "time_sp"
	countPerRotFD     = "count_per_rot"
)

func findFolder(port OutPort) string {
	if _, err := os.Stat(rootMotorPath); os.IsNotExist(err) {
		log.Fatal("There are no motors connected")
	}

	rootMotorFolder, _ := os.Open(rootMotorPath)
	defer rootMotorFolder.Close()
	motorFolders, _ := rootMotorFolder.Readdir(-1)
	if len(motorFolders) == 0 {
		log.Fatal("There are no motors connected")
	}

	for _, folderInfo := range motorFolders {
		folder := folderInfo.Name()
		motorPort := utilities.ReadStringValue(path.Join(rootMotorPath, folder), portFD)
		if motorPort == "out"+string(port) {
			return path.Join(rootMotorPath, folder)
		}
	}

	log.Fatal("No motor is connected to port ", port)
	return ""
}

func setSpeed(folder string, speed int16) {
	regulationMode := utilities.ReadStringValue(folder, regulationModeFD)

	switch regulationMode {
	case "on":
		utilities.WriteIntValue(folder, speedSetterFD, int64(speed))
	case "off":
		if speed > 100 || speed < -100 {
			log.Fatal("The speed must be in range [-100, 100]")
		}
		utilities.WriteIntValue(folder, powerSetterFD, int64(speed))
	}
}

func setAngle(folder string, angle int16) {
	utilities.WriteIntValue(folder, desiredPositionFD, int64(angle))
}

func setTime(folder string, seconds int32) {
	utilities.WriteIntValue(folder, timeFD, int64(seconds))
}

func run(folder string, speed int16, command RunCommand) {
	setSpeed(folder, speed)
	utilities.WriteStringValue(folder, runFD, string(command))
}

// RunForever runs the motor at the given port.
// The meaning of `speed` parameter depends on whether the regulation mode is turned on or off.
//
// When the regulation mode is off (by default) `speed` ranges from -100 to 100 and
// it's absolute value indicates the percent of motor's power usage. It can be roughly interpreted as
// a motor speed, but deepending on the environment, the actual speed of the motor
// may be lower than the target speed.
//
// When the regulation mode is on (has to be enabled by EnableRegulationMode function) the motor
// driver attempts to keep the motor speed at the `speed` value you've specified
// which ranges from about -1000 to 1000. The actual range depends on the type of the motor - see ev3dev docs.
//
// Negative values indicate reverse motion regardless of the regulation mode.
func RunForever(port OutPort, speed int16) {
	folder := findFolder(port)
	run(folder, speed, runForeverCommand)
}

// Rotate moves the motor to the specified angle relative to the current position in degrees at the specified speed
func Rotate(port OutPort, angle, speed int16) {
	folder := findFolder(port)
	setAngle(folder, angle)
	run(folder, speed, runToRelPosCommand)
}

// RotateTo moves the motor to the given angle relative to the current position in degrees at the specified speed
func RotateTo(port OutPort, angle, speed int16) {
	folder := findFolder(port)
	setAngle(folder, angle)
	run(folder, speed, runToAbsPosCommand)
}

// RunFor runs the motor at the given port for the given time in seconds at the given speed
func RunFor(port OutPort, time int32, speed int16) {
	folder := findFolder(port)
	setTime(folder, time)
	run(folder, speed, runTimedCommand)
}

// Stop stops the motor at the given port.
func Stop(port OutPort) {
	utilities.WriteStringValue(findFolder(port), runFD, "stop")
}

// CurrentSpeed reads the operating speed of the motor at the given port.
func CurrentSpeed(port OutPort) int16 {
	return utilities.ReadInt16Value(findFolder(port), speedGetterFD)
}

// CurrentPower reads the operating power of the motor at the given port.
func CurrentPower(port OutPort) int16 {
	return utilities.ReadInt16Value(findFolder(port), powerGetterFD)
}

// EnableRegulationMode enables regulation mode, causing the motor at the given port to compensate
// for any resistance and maintain its target speed.
func EnableRegulationMode(port OutPort) {
	utilities.WriteStringValue(findFolder(port), regulationModeFD, "on")
}

// DisableRegulationMode disables regulation mode. Regulation mode is off by default.
func DisableRegulationMode(port OutPort) {
	utilities.WriteStringValue(findFolder(port), regulationModeFD, "off")
}

// SetStopMode sets the brake mode to ether Coast, Brake or Hold.
func SetStopMode(port OutPort, mode StopMode) {
	utilities.WriteStringValue(findFolder(port), stopModeFD, string(mode))
}

// CurrentPosition reads the position of the motor at the given port.
func CurrentPosition(port OutPort) int32 {
	return utilities.ReadInt32Value(findFolder(port), positionFD)
}

// InitializePosition sets the position of the motor at the given port.
func InitializePosition(port OutPort, value int32) {
	utilities.WriteIntValue(findFolder(port), positionFD, int64(value))
}
