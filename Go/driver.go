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
		// TODO timeout
	}

	if status&errorBits != 0 {
		driver.device.Write(control, clearCommand)
		return DeviceError{status & errorBits, address, data}
	}

	driver.device.Read(address)

	return nil
}

// DeviceError is the error caused by hardware errors.
type DeviceError struct {
	errorBits byte
	address   uint32
	data      byte
}

func (error DeviceError) Error() string {
	return fmt.Sprintf("%s at 0x%X", error.cause(), error.address)
}

func (error DeviceError) cause() string {
	if error.errorBits&hardwareErrorBit != 0 {
		return "Hardware Error"
	}
	if error.errorBits&internalErrorBit != 0 {
		return "Internal Error"
	}
	if error.errorBits&protectionErrorBit != 0 {
		return "Protected Block Error"
	}
	return "Unknown Error"
}
