package codekata

// DeviceDriver is used by the operating system to interact with the hardware 'FlashMemoryDevice'.
type DeviceDriver struct {
	device FlashMemoryDevice
}

func (driver DeviceDriver) Read(address uint32) (byte, error) {
	return driver.device.Read(address), nil
}

func (driver DeviceDriver) Write(address uint32, data byte) error {
	// write 0x40, write data, read 0x0, ready bit set, success bits, read data.
	driver.device.Write(0x0, 0x40)
	driver.device.Write(address, data)
	driver.device.Read(0x0)     // status
	driver.device.Read(address) // verify
	return nil
}
