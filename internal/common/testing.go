package common

import (
	"testing"
	"time"
)

// ShouldPanic assert that `f` panics during execution
func ShouldPanic(t *testing.T, f func()) {
	defer func() { recover() }()
	f()
	t.Errorf("should have panicked")
}

func FrozenInTime() time.Time {
	loc, err := time.LoadLocation("EST")
	if err != nil {
		panic("Could not load EST location")
	}
	//Dragon splashdown historic event time
	return time.Date(2020, 8, 2, 14, 48, 0, 0, loc)
}
