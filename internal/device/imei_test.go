package device

import (
	"runtime"
	"testing"

	"github.com/rilopez/redis-wire-protocol/internal/common"
)

func TestDecode(t *testing.T) {
	b := []byte{4, 9, 0, 1, 5, 4, 2, 0, 3, 2, 3, 7, 5, 1, 8}
	expectedIMEI := uint64(490154203237518)
	actualIMEI, _ := decodeIMEI(b)

	if actualIMEI != expectedIMEI {
		t.Errorf("expecting ID %d but got %d", expectedIMEI, actualIMEI)
	}
}

func TestDecodeErrCheckSum(t *testing.T) {
	b := []byte{4, 9, 0, 1, 5, 4, 2, 0, 3, 2, 3, 7, 5, 1, 1}
	_, err := decodeIMEI(b)
	if err != errIMEIChecksum {
		t.Errorf("expecting ErrChecksum but got  %v", err)
	}
}

func TestDecodeWithNonDigitsErrInvalid(t *testing.T) {
	b := []byte{4, 9, 'A', 'B', 5, 4, 2, 0, 3, 2, 3, 7, 5, 1, 8}
	_, err := decodeIMEI(b)
	if err != errIMEIInvalid {
		t.Errorf("expecting ErrInvalid but got %s", err)
	}
}

func TestDecodePanicAtLeast15byteslong(t *testing.T) {
	//it panics if b isn't at least 15 bytes long.
	onlyTwoBytes := []byte{1, 2}

	common.ShouldPanic(t, func() {
		_, _ = decodeIMEI(onlyTwoBytes)
	})
}

func TestDecodeIMEI_Allocations(t *testing.T) {
	defer runtime.GOMAXPROCS(runtime.GOMAXPROCS(1))
	var start, end runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&start)
	_, _ = decodeIMEI([]byte{4, 9, 0, 1, 5, 4, 2, 0, 3, 2, 3, 7, 5, 1, 8})
	runtime.ReadMemStats(&end)
	alloc := end.TotalAlloc - start.TotalAlloc
	if alloc > 0 {
		t.Errorf("Decode should NOT allocate under any condition, it allocated %d bytes", alloc)
	}
}

func BenchmarkDecodeIMEI(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = decodeIMEI([]byte{4, 9, 0, 1, 5, 4, 2, 0, 3, 2, 3, 7, 5, 1, 8})
	}
	b.StopTimer()
}
