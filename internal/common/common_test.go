package common

import (
	"bytes"
	"testing"
)

func TestImeiStringToBytes(t *testing.T) {
	expectedIMEIbytes := []byte{4, 9, 0, 1, 5, 4, 2, 0, 3, 2, 3, 7, 5, 1, 8}
	imei := "490154203237518"
	actualIMEIbytes, err := ImeiStringToBytes(&imei)

	if err != nil {
		t.Error(err)
	}

	if !bytes.Equal(expectedIMEIbytes, actualIMEIbytes[:]) {
		t.Errorf("expecting imei bytes  %v but got %v", expectedIMEIbytes, actualIMEIbytes)
	}
}
