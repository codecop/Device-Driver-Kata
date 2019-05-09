package codekata

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRead(t *testing.T) {
	hardware := makeMockHardware(t)
	hardware.expectRead(0xFF, 3)
	driver := DeviceDriver{hardware}

	data, err := driver.Read(0xFF)

	assert.EqualValues(t, 3, data, "read value")
	assert.NoError(t, err)
	hardware.verifyAllInteractions()
}

func TestSuccessfulWriteReadyAtFirstCheck(t *testing.T) {
	hardware := makeMockHardware(t)
	hardware.expectWriteProgramCommand()
	hardware.expectWrite(0xAB, 42)
	hardware.expectReadStatus(0x80) // ready bit set, success bits
	driver := DeviceDriver{hardware}

	err := driver.Write(0xAB, 42)

	assert.NoError(t, err)
	hardware.verifyAllInteractions()
}

func TestSuccessfulWriteReadyAtThirdCheck(t *testing.T) {
	hardware := makeMockHardware(t)
	hardware.expectWriteProgramCommand()
	hardware.expectWrite(0x42, 21)
	hardware.expectReadStatus(0x00) // not ready
	hardware.expectReadStatus(0x00) // not ready
	hardware.expectReadStatus(0x80)
	driver := DeviceDriver{hardware}

	err := driver.Write(0x42, 21)

	assert.NoError(t, err)
	hardware.verifyAllInteractions()
}

func TestFailedWriteWithHardwareError(t *testing.T) {
	hardware := makeMockHardware(t)
	hardware.expectWriteProgramCommand()
	hardware.expectWrite(0xAB, 42)
	hardware.expectReadStatus(0x88) // ready bit set, error bit
	hardware.expectWrite(0x0, 0xFF) // reset
	driver := DeviceDriver{hardware}

	err := driver.Write(0xAB, 42)

	assert.EqualError(t, err, "Hardware Error at 0xAB")
	hardware.verifyAllInteractions()
}

func TestFailedWriteWithProtectedBlockError(t *testing.T) {
	hardware := makeMockHardware(t)
	hardware.expectWriteProgramCommand()
	hardware.expectWrite(0xAB, 42)
	hardware.expectReadStatus(0xA0) // ready bit set, error bit
	hardware.expectWrite(0x0, 0xFF) // reset
	driver := DeviceDriver{hardware}

	err := driver.Write(0xAB, 42)

	assert.EqualError(t, err, "Protected Block Error at 0xAB")
	hardware.verifyAllInteractions()
}

func TestSuccessfulWriteWithRetryAfterInternalError(t *testing.T) {
	hardware := makeMockHardware(t)
	hardware.expectWriteProgramCommand()
	hardware.expectWrite(0xAB, 42)
	hardware.expectReadStatus(0x90) // ready bit set, error bit
	hardware.expectWrite(0x0, 0xFF) // reset
	// retry
	hardware.expectWriteProgramCommand()
	hardware.expectWrite(0xAB, 42)
	hardware.expectReadStatus(0x80)
	driver := DeviceDriver{hardware}

	err := driver.Write(0xAB, 42)

	assert.NoError(t, err)
	hardware.verifyAllInteractions()
}

func TestSuccessfulWriteWith3RetriesAfterInternalError(t *testing.T) {
	hardware := makeMockHardware(t)
	hardware.expectWriteProgramCommand()
	hardware.expectWrite(0xAB, 42)
	hardware.expectReadStatus(0x90) // ready bit set, error bit
	hardware.expectWrite(0x0, 0xFF) // reset
	// retry 1
	hardware.expectWriteProgramCommand()
	hardware.expectWrite(0xAB, 42)
	hardware.expectReadStatus(0x90) // ready bit set, error bit
	hardware.expectWrite(0x0, 0xFF) // reset
	// retry 2
	hardware.expectWriteProgramCommand()
	hardware.expectWrite(0xAB, 42)
	hardware.expectReadStatus(0x90) // ready bit set, error bit
	hardware.expectWrite(0x0, 0xFF) // reset
	// retry 3
	hardware.expectWriteProgramCommand()
	hardware.expectWrite(0xAB, 42)
	hardware.expectReadStatus(0x80)
	driver := DeviceDriver{hardware}

	err := driver.Write(0xAB, 42)

	assert.NoError(t, err)
	hardware.verifyAllInteractions()
}

func TestFailedWriteWithInternalError(t *testing.T) {
	hardware := makeMockHardware(t)
	hardware.expectWriteProgramCommand()
	hardware.expectWrite(0xCA, 42)
	hardware.expectReadStatus(0x90) // ready bit set, error bit
	hardware.expectWrite(0x0, 0xFF) // reset
	// retry 1
	hardware.expectWriteProgramCommand()
	hardware.expectWrite(0xCA, 42)
	hardware.expectReadStatus(0x90) // ready bit set, error bit
	hardware.expectWrite(0x0, 0xFF) // reset
	// retry 2
	hardware.expectWriteProgramCommand()
	hardware.expectWrite(0xCA, 42)
	hardware.expectReadStatus(0x90) // ready bit set, error bit
	hardware.expectWrite(0x0, 0xFF) // reset
	// retry 3
	hardware.expectWriteProgramCommand()
	hardware.expectWrite(0xCA, 42)
	hardware.expectReadStatus(0x90) // ready bit set, error bit
	hardware.expectWrite(0x0, 0xFF) // reset
	driver := DeviceDriver{hardware}

	err := driver.Write(0xCA, 42)

	assert.EqualError(t, err, "Internal Error at 0xCA")
	hardware.verifyAllInteractions()
}

type mockOperation struct {
	kind    string // read or write
	address uint32
	value   byte
}

type mockHardware struct {
	t          *testing.T
	operations []mockOperation
	replay     int
}

func makeMockHardware(t *testing.T) *mockHardware {
	return &mockHardware{t, make([]mockOperation, 0), 0}
}

func (mock *mockHardware) addOperation(operation mockOperation) {
	mock.operations = append(mock.operations, operation)
}

func (mock *mockHardware) expectRead(address uint32, value byte) {
	mock.addOperation(mockOperation{"read", address, value})
}

func (mock *mockHardware) expectReadStatus(status byte) {
	mock.expectRead(0x00, status)
}

func (mock *mockHardware) expectWrite(address uint32, value byte) {
	mock.addOperation(mockOperation{"write", address, value})
}

func (mock *mockHardware) expectWriteProgramCommand() {
	mock.expectWrite(0x0, 0x40)
}

func (mock *mockHardware) nextOperation() mockOperation {
	operation := mock.operations[mock.replay]
	mock.replay++
	return operation
}

func (mock *mockHardware) Read(address uint32) byte {
	operation := mock.nextOperation()

	assert.EqualValues(mock.t, "read", operation.kind, "expectation")
	assert.EqualValues(mock.t, operation.address, address, "address")

	return operation.value
}

func (mock *mockHardware) Write(address uint32, value byte) {
	operation := mock.nextOperation()

	assert.EqualValues(mock.t, "write", operation.kind, "expectation")
	assert.EqualValues(mock.t, operation.address, address, "address")
	assert.EqualValues(mock.t, operation.value, value, "written value")
}

func (mock *mockHardware) verifyAllInteractions() {
	assert.EqualValues(mock.t, len(mock.operations), mock.replay, "all interactions")
}

/*
Test cases
==========

* timeout when reading status -> error

Requirements
============

In this Kata, your job is to implement the device driver that operates a flash
memory device. The protocol for talking to the hardware is outlined below. The
code you develop should allow the operating system to both read and write
binary data to and from the device.

Flash Memory Device Protocol
----------------------------

When you want to write data to the device:

  - Begin by writing the 'Program' command, 0x40 to address 0x0
  - Then make a call to write the data to the address you want to write to.
  - Then read the value in address 0x0 and check bit 7 in the returned data,
	also known as the 'ready bit'. Repeat the read operation until the ready bit is
	set to 1. This means the write operation is complete. In a typical device it
	should take around ten microseconds, but it will vary from write to write.
  - There could have been an error, so in the value from address 0x0, examine
	the contents of the other bits. If all of them are set to 0 then the operation
	was successful.
  - You should then make a read from the memory address you previously set, in
	order to check it returns the value you set.
  - If that is successful, then you can assume the whole write operation was
	successful.

In the case of an error, the device sets the one of the other bits in the data
at location 0x0, that is, bit 3, 4 or 5. The meaning of the error codes is as
follows:

  - bit 3: Vpp error. The voltage on the device was wrong and it is now
	physically damaged.
  - bit 4: internal error. Something went wrong this time but next time it might
	work.
  - bit 5: protected block error. You cannot write to that address, but other
	addresses may work.

If any of these error bits are set, you must write 0xFF to address 0x0 before
the device will accept any new write requests.
This will reset the error bits to zero. Note that until the 'ready bit' is set,
then you will not get valid values for any of the error bits.

When you want to read data from the device:

  - simply make a call to read the contents of the address. There is no need to
	begin by writing the 'Program' command.
*/
