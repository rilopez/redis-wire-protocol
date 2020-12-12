package device

import (
	"errors"
	"math"
)

var (
	errIMEIInvalid  = errors.New("ID: invalid ")
	errIMEIChecksum = errors.New("ID: invalid checksum")
)

/*decodeIMEI implements an IMEI decoder.

returns the IMEI code contained in the first 15 bytes of b. In case b isn't strictly
composed of digits, the returned error will be ErrIMEIInvalid.

In case b's checksum is wrong, the returned error will be ErrIMEIInvalid.
DecodeIMEI does NOT allocate under any condition. Additionally, it panics if b
isn't at least 15 bytes long.

NOTE: for more information about IMEI codes and their structure you may
consult with:
https://en.wikipedia.org/wiki/International_Mobile_Equipment_Identity.
*/
func decodeIMEI(b []byte) (code uint64, err error) {
	_ = b[14] // nice trick to hint bound checks to the compiler | https://medium.com/@brianblakewong/optimizing-go-bounds-check-elimination-f4be681ba030

	place := 14
	var checksum byte
	var imeiAsFloat float64

	for i := 0; i < 15; i++ {
		currDigit := b[i]
		digitForChecksum := currDigit
		if currDigit > 9 {
			// In case b isn't strictly composed of digits, the returned error will be
			// ErrInvalid.
			return 0, errIMEIInvalid
		}
		if (i+1)%2 == 0 {
			digitForChecksum *= 2
			if digitForChecksum > 9 {
				digitForChecksum = (digitForChecksum % 10) + 1
			}
		}
		checksum += digitForChecksum
		imeiAsFloat += float64(currDigit) * math.Pow10(place)
		place--
	}

	if checksum%10 != 0 {
		return 0, errIMEIChecksum
	}

	return uint64(imeiAsFloat), nil
}
