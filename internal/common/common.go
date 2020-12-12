package common

import (
	"errors"
	"fmt"
	"strconv"
)

// ErrNotImplemented is raised throughout the codebase of the challenge to
// denote implementations to be done by the candidate.
var ErrNotImplemented = errors.New("not implemented")

// ImeiStringToBytes converts an IMEI string to bytes array. This is not only a character to bytes convertion, it parses each char as number
func ImeiStringToBytes(imei *string) ([15]byte, error) {
	var buf [15]byte

	if len(*imei) < 15 {
		return buf, errors.New("IMEI string should have at least 15 characters")
	}
	for i := 0; i < len(*imei); i++ {
		s := (*imei)[i]
		digit, err := strconv.Atoi(string(s))
		if err == nil {
			buf[i] = byte(digit)
		} else {
			return buf, fmt.Errorf("IMEI has an invalid digit character at %d", i)
		}
	}
	return buf, nil
}
