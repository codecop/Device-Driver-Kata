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

// DeviceDriver is used by the operating system to interact with the hardware 'FlashMemoryDevice'.
type DeviceDriver struct {
	device FlashMemoryDevice
}

func (driver DeviceDriver) Read(address uint32) (byte, error) {
	return driver.device.Read(address), nil
}

func (driver DeviceDriver) Write(address uint32, data byte) error {
	driver.device.Write(control, programCommand)
	driver.device.Write(address, data)

	var status byte
	for status&readyBit == 0 {
		status = driver.device.Read(control)
	}

	if status&errorBits != 0 {
		driver.device.Write(control, clearCommand)
		return DeviceError{address, data, status & errorBits}
	}
	return nil
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
