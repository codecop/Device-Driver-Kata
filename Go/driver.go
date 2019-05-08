package codekata

const control uint32 = 0x0

const programCommand byte = 0x40
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

	driver.device.Read(address)

	return nil
}
