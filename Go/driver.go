package codekata

import "fmt"

const control uint32 = 0x0

const programCommand byte = 0x40
const clearCommand byte = 0xFF

const hardwareErrorBit byte = 0x08
const internalErrorBit byte = 0x10
const protectionErrorBit byte = 0x20
const errorBits byte = hardwareErrorBit | internalErrorBit | protectionErrorBit
const readyBit byte = 0x80

const retries = 99

// DeviceDriver is used by the operating system to interact with the hardware 'FlashMemoryDevice'.
type DeviceDriver struct {
	device FlashMemoryDevice
}

func (driver DeviceDriver) Read(address uint32) (byte, error) {
	return driver.device.Read(address), nil
}

func (driver DeviceDriver) Write(address uint32, data byte) error {
	var status byte // TODO status object with queries for kind of errors, hide bit masking...

	for try := 0; try <= retries; try++ {
		driver.writeData(address, data)
		status = driver.waitReady()
		if status&errorBits == 0 {
			return nil
		}
		driver.reset()
		if status&internalErrorBit == 0 {
			break
		}
	}

	return DeviceError{address, data, status}
}

func (driver DeviceDriver) writeData(address uint32, data byte) {
	driver.device.Write(control, programCommand)
	driver.device.Write(address, data)
}

func (driver DeviceDriver) waitReady() byte {
	var status byte
	for status&readyBit == 0 {
		status = driver.device.Read(control)
	}
	return status
}

func (driver DeviceDriver) reset() {
	driver.device.Write(control, clearCommand)
}

// DeviceError is the error caused by hardware errors.
type DeviceError struct {
	address   uint32
	data      byte
	errorBits byte
}

func (e DeviceError) Error() string {
	return fmt.Sprintf("%s at 0x%X", e.cause(), e.address)
}

func (e DeviceError) cause() string {
	if e.errorBits&hardwareErrorBit != 0 {
		return "Hardware Error"
	}
	if e.errorBits&internalErrorBit != 0 {
		return "Internal Error"
	}
	if e.errorBits&protectionErrorBit != 0 {
		return "Protected Block Error"
	}
	return "Unknown Error"
}
