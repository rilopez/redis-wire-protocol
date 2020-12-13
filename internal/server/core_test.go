package server

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/rilopez/redis-wire-protocol/internal/common"
	"github.com/rilopez/redis-wire-protocol/internal/device"
)

func TestNewCore(t *testing.T) {
	core := newCore(common.FrozenInTime, uint(1337), 2)
	expectedClientsLen := 0
	actualClientsLen := core.numConnectedDevices()
	if actualClientsLen != expectedClientsLen {
		t.Errorf("expected len(core.client) to equal %d but got %d", expectedClientsLen, actualClientsLen)
	}
}

func TestOutputReading(t *testing.T) {
	expectedRecord := "1257894000000000000,490154203237518,67.770000,2.635550,33.410000,44.400000,0.256660"
	actualRecord := formatReadingOutput(490154203237518, 1257894000000000000, &device.Reading{
		Temperature:  67.77,
		Altitude:     2.63555,
		Latitude:     33.41,
		Longitude:    44.4,
		BatteryLevel: 0.25666,
	})

	if expectedRecord != actualRecord {
		t.Errorf("expected %s got %s", expectedRecord, actualRecord)
	}

}

func TestCore_HandleReading(t *testing.T) {
	//Setup
	expectedLastReadingEpoch := common.FrozenInTime().UnixNano()
	expectedPayload := device.CreateRandReadingBytes()

	core := newCore(common.FrozenInTime, uint(1337), 2)
	expectedClientIMEI := uint64(448324242329542)
	dev := &connectedClient{}
	core.clients[expectedClientIMEI] = dev

	//Exercise
	core.handleSET(expectedClientIMEI, expectedPayload[:])

	if dev.lastReadingEpoch != expectedLastReadingEpoch {
		t.Errorf("expected LastReadingEpoch to equal %d but got %d",
			expectedLastReadingEpoch,
			dev.lastReadingEpoch)
	}
	expectedReading := &device.Reading{}
	expectedReading.Decode(expectedPayload[:])
	if !reflect.DeepEqual(expectedReading, dev.lastReading) {
		t.Errorf("expected LastReading to equal %v but got %v",
			expectedReading,
			dev.lastReading)
	}
}

func TestCore_HandleReading_UnknownClient(t *testing.T) {
	//Setup
	core := newCore(common.FrozenInTime, uint(1337), 2)

	//Exercise

	unknownIMEI := uint64(123)
	err := core.handleSET(unknownIMEI, []byte{1, 2})
	if err == nil {
		t.Errorf("expected get an error for unknown client %d", unknownIMEI)
	}
}

func TestCore_HandleReading_InvalidPayload(t *testing.T) {
	//Setup
	core := newCore(common.FrozenInTime, uint(1337), 2)
	expectedClientIMEI := uint64(448324242329542)
	dev := &connectedClient{}
	core.clients[expectedClientIMEI] = dev

	//Exercise bound check panic
	errBoundCheckPanic := core.handleSET(expectedClientIMEI, []byte{1, 2})
	if errBoundCheckPanic == nil {
		t.Errorf("expected get an error for unknown client %d", expectedClientIMEI)
	}

	invalidPayload := device.NewPayload(9999999, 9999999, 9999999, 9999999, 9999999)

	errInvalidPayload := core.handleSET(expectedClientIMEI, invalidPayload[:])
	if errInvalidPayload == nil {
		t.Errorf("expected get an error for unknown client %d", expectedClientIMEI)
	}
}

func TestCore_Register(t *testing.T) {
	core := newCore(common.FrozenInTime, uint(1337), 2)
	expectedIMEI := uint64(448324242329542)
	callBackChannel := make(chan common.Command, 1)
	err := core.register(expectedIMEI, callBackChannel)
	if err != nil {
		t.Errorf("unexpected error %v", err)
	}

	_, exists := core.clientByID(expectedIMEI)
	if !exists {
		t.Errorf("clients map should contain an entry for IMEI: %d", expectedIMEI)
	}
	cmd := <-callBackChannel
	if cmd.ID != common.WELCOME {
		t.Errorf("Expected callback channel to receive a WELCOME cmd but got %v", cmd.ID)
	}
}

func TestCore_Register_ExistingClient(t *testing.T) {
	// Setup
	core := newCore(common.FrozenInTime, uint(1337), 2)
	expectedIMEI := uint64(448324242329542)
	callBackChannel := make(chan common.Command, 2)

	//Exercise
	err := core.register(expectedIMEI, callBackChannel)
	if err != nil {
		t.Errorf("Unexpected err (%v) while trying to register %d", err, expectedIMEI)
	}
	<-callBackChannel //ignore welcome cmd

	err = core.register(expectedIMEI, callBackChannel)
	if err == nil {
		t.Errorf("An error is expected when trying to register an existing client ")
	}

	cmd := <-callBackChannel
	if cmd.ID != common.KILL {
		t.Error("Expecting a kill command from the back channel ")
	}

}

func TestCore_Deregister_ExistingClient(t *testing.T) {
	// Setup
	core := newCore(common.FrozenInTime, uint(1337), 2)
	imei := uint64(448324242329542)
	callBackChannel := make(chan common.Command, 1)

	//Exercise
	err := core.register(imei, callBackChannel)
	if err != nil {
		t.Errorf("Unexpected err (%v)while trying to register %d", err, imei)
	}

	err = core.deregister(imei)
	if err != nil {
		t.Errorf("Unexpected error trying to deregister an existing client %v ", err)
	}
}

func TestCore_Deregister_UnknownClient(t *testing.T) {
	// Setup
	core := newCore(common.FrozenInTime, uint(1337), 2)
	expectedClientIMEI := uint64(448324242329542)

	//Exercise

	err := core.deregister(expectedClientIMEI)
	if err == nil {
		t.Errorf("An error is expected when trying to deregister an unknown client")
	}
}

func ExampleCore_handleReading() {
	//Setup

	expectedPayload := device.NewPayload(9.127577, 12545.598440, -51.432503, -42.963412, 31.805817)

	core := newCore(common.FrozenInTime, uint(1337), 2)
	expectedIMEI := uint64(448324242329542)
	callBackChannel := make(chan common.Command, 1)

	//Exercise
	err := core.register(expectedIMEI, callBackChannel)
	if err != nil {
		fmt.Printf("Unexpected err (%v) while trying to register %d", err, expectedIMEI)
	}
	//Exercise

	core.handleSET(expectedIMEI, expectedPayload[:])

	// Output: 1596397680000000000,448324242329542,9.127577,12545.598440,-51.432503,-42.963412,31.805817
}

func BenchmarkCore_HandleReading(b *testing.B) {
	//Setup
	expectedPayload := device.CreateRandReadingBytes()

	core := newCore(common.FrozenInTime, uint(1337), 2)

	expectedClientIMEI := uint64(448324242329542)

	callBackChannel := make(chan common.Command, 1)

	err := core.register(expectedClientIMEI, callBackChannel)
	if err != nil {
		b.Error(err)
	}

	//Exercise

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		fmt.Printf("reading %d of %d readings", i, b.N)
		err := core.handleSET(expectedClientIMEI, expectedPayload[:])
		if err != nil {
			b.Fail()
		}
	}
	b.StopTimer()

}
