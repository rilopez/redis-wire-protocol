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

func ExpectNoError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Errorf("expected no error , got %v", err)
	}
}

func AssertEquals(t *testing.T, got interface{}, want interface{}) {
	t.Helper()
	if got != want {
		t.Errorf("expecting %s , got %s", want, got)
	}
}
