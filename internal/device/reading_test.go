package device

import (
	"runtime"
	"testing"

	"github.com/rilopez/redis-wire-protocol/internal/common"
)

func TestReading_Decode(t *testing.T) {
	expectedTemperature := float64(38)
	expectedAltitude := float64(10)
	expectedLatitude := float64(21.033643)
	expectedLongitude := float64(-89.5969049)
	expectedBatteryLevel := float64(45)

	payload := NewPayload(
		expectedTemperature,
		expectedAltitude,
		expectedLatitude,
		expectedLongitude,
		expectedBatteryLevel)

	reading := Reading{}
	reading.Decode(payload[:])

	if reading.Temperature != expectedTemperature {
		t.Errorf("expected Temperature: %f got %f", expectedTemperature, reading.Temperature)
	}

	if reading.Altitude != expectedAltitude {
		t.Errorf("expected Altitude: %f got %f", expectedAltitude, reading.Altitude)
	}

	if reading.Latitude != expectedLatitude {
		t.Errorf("expected Latitude: %f got %f", expectedLatitude, reading.Latitude)
	}

	if reading.Longitude != expectedLongitude {
		t.Errorf("expected Longitude: %f got %f", expectedLongitude, reading.Longitude)
	}

	if reading.BatteryLevel != expectedBatteryLevel {
		t.Errorf("expected BatteryLevel: %f got %f", expectedBatteryLevel, reading.BatteryLevel)
	}

}

func TestDecodePanicAtLeast40byteslong(t *testing.T) {
	//it panics if b isn't at least 15 bytes long.
	var only39 [39]byte
	reading := Reading{}

	common.ShouldPanic(t, func() {
		reading.Decode(only39[:])
	})
}

func TestReading_Decode_Allocations(t *testing.T) {
	defer runtime.GOMAXPROCS(runtime.GOMAXPROCS(1))
	var start, end runtime.MemStats
	reading := Reading{}
	payload := NewPayload(30, 5, 21, -89, 99)
	runtime.GC()
	runtime.ReadMemStats(&start)

	reading.Decode(payload[:])

	runtime.ReadMemStats(&end)
	alloc := end.TotalAlloc - start.TotalAlloc
	if alloc > 0 {
		t.Errorf("Decode should NOT allocate under any condition, it allocated %d bytes", alloc)
	}
}

func BenchmarkReading_Decode(b *testing.B) {
	randomReading := CreateRandReadingBytes()
	reading := Reading{}
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		if ok := reading.Decode(randomReading[:]); !ok {
			b.Fail()
		}
	}
	b.StopTimer()
}
