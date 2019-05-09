package codekata

import "fmt"

const controlAddress uint32 = 0x0

const programCommand byte = 0x40
const clearCommand byte = 0xFF

// hardwareStatus is the returned bitmask
type hardwareStatus byte

const readyBit hardwareStatus = 0x80
const hardwareErrorBit hardwareStatus = 0x08
const internalErrorBit hardwareStatus = 0x10
const protectionErrorBit hardwareStatus = 0x20

func (status hardwareStatus) isNotReady() bool {
	return status&readyBit == 0
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

const retries = 3

// DeviceDriver is used by the operating system to interact with the hardware 'FlashMemoryDevice'.
type DeviceDriver struct {
	device FlashMemoryDevice
}

func (driver DeviceDriver) Read(address uint32) (byte, error) {
	return driver.device.Read(address), nil
}

func (driver DeviceDriver) Write(address uint32, data byte) error {
	var status hardwareStatus

	for try := 0; try <= retries; try++ {
		driver.writeData(address, data)

		status = driver.waitReady()
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

func (driver DeviceDriver) waitReady() hardwareStatus {
	var status hardwareStatus
	for status.isNotReady() {
		status = hardwareStatus(driver.device.Read(controlAddress))
	}
	return status
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
