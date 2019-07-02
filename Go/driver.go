package codekata

import (
	"fmt"
)

// hardwareStatus is the returned bitmask
type hardwareStatus byte

const (
	readyBit hardwareStatus = 0x80
	hardwareErrorBit hardwareStatus = 0x08
 	internalErrorBit hardwareStatus = 0x10
	protectionErrorBit hardwareStatus = 0x20
)

func (status hardwareStatus) isReady() bool {
	return status&readyBit != 0
}

func (status hardwareStatus) isSuccess() bool {
	return !(status.isHardwareError() || status.isInternalError() || status.isProtectionError())
}

func (status hardwareStatus) isNotRecoverableError() bool {
	return status.isHardwareError() || status.isProtectionError()
}

func (status hardwareStatus) isHardwareError() bool {
	return status&hardwareErrorBit != 0
}

func (status hardwareStatus) isInternalError() bool {
	return status&internalErrorBit != 0
}

func (status hardwareStatus) isProtectionError() bool {
	return status&protectionErrorBit != 0
}

type timerMilliseconds func() uint64

const controlAddress uint32 = 0x0

const (
	programCommand byte = 0x40
	clearCommand byte = 0xFF
)

const retries = 3
const timeout = 100 // milli seconds

// DeviceDriver is used by the operating system to interact with the hardware 'FlashMemoryDevice'.
type DeviceDriver struct {
	device FlashMemoryDevice
	timer  timerMilliseconds
}

func (driver DeviceDriver) Read(address uint32) (byte, error) {
	return driver.device.Read(address), nil
}

func (driver DeviceDriver) Write(address uint32, data byte) error {
	var status hardwareStatus
	var err error

	for try := 0; try <= retries; try++ {
		driver.writeData(address, data)

		status, err = driver.waitReady()
		if err != nil {
			return err
		}
		if status.isSuccess() {
			return nil
		}

		driver.reset()
		if status.isNotRecoverableError() {
			break
		}
	}

	return deviceError{address, data, status}
}

func (driver DeviceDriver) writeData(address uint32, data byte) {
	driver.writeControl(programCommand)
	driver.device.Write(address, data)
}

func (driver DeviceDriver) writeControl(data byte) {
	driver.device.Write(controlAddress, data)
}

func (driver DeviceDriver) waitReady() (hardwareStatus, error) {
	startTime := driver.timer()
	for {
		status := hardwareStatus(driver.device.Read(controlAddress))
		if status.isReady() {
			return status, nil
		}

		pastMillis := driver.timer() - startTime
		if pastMillis >= timeout {
			break
		}
	}
	return hardwareStatus(0), deviceTimeout{timeout}
}

func (driver DeviceDriver) reset() {
	driver.writeControl(clearCommand)
}

// deviceError is the error caused by hardware errors.
type deviceError struct {
	address   uint32
	data      byte
	errorBits hardwareStatus
}

func (e deviceError) Error() string {
	return fmt.Sprintf("%s at 0x%X", e.cause(), e.address)
}

func (e deviceError) cause() string {
	if e.errorBits.isHardwareError() {
		return "Hardware Error"
	}
	if e.errorBits.isInternalError() {
		return "Internal Error"
	}
	if e.errorBits.isProtectionError() {
		return "Protected Block Error"
	}
	return "Unknown Error"
}

type deviceTimeout struct {
	time uint32
}

func (e deviceTimeout) Error() string {
	return fmt.Sprint("Timeout")
}
