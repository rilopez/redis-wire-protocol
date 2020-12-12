package device

import (
	"log"
	"net"
	"time"

	"github.com/rilopez/redis-wire-protocol/internal/common"
)

// Randomatic implements a simple TCP client that sends `n` random readings after login
func Randomatic(clientServerAddress *string, clientImei *string, numReadings *uint, readingRate *uint) {
	baseClient(clientServerAddress, clientImei, time.Duration(*readingRate)*time.Millisecond, time.Nanosecond, numReadings)
}

// Slowmatic implements a client that will be send be disconected by the server  because it takes more than 2 seconds between msgs
func Slowmatic(clientServerAddress *string, clientImei *string, numReadings *uint) {
	baseClient(clientServerAddress, clientImei, 3*time.Second, time.Nanosecond, numReadings)
}

// TooSlowToPlayWithGrownups implements a client that is too slow to send the initial login message, so the server will disconnect the connection
func TooSlowToPlayWithGrownups(clientServerAddress *string, clientImei *string, numReadings *uint) {
	baseClient(clientServerAddress, clientImei, time.Second, 2*time.Second, numReadings)
}

func baseClient(clientServerAddress *string, clientImei *string, readingRate time.Duration, sleepBeforeLogin time.Duration, numReadings *uint) {
	log.Printf("Connecting to %s", *clientServerAddress)
	conn, err := net.Dial("tcp", *clientServerAddress)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	log.Printf("DEBUG: converting ID %s to bytes", *clientImei)
	imeiBytes, err := common.ImeiStringToBytes(clientImei)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("DEBUG: sending login ID %v to server", imeiBytes)
	time.Sleep(sleepBeforeLogin)
	n, err := conn.Write(imeiBytes[:])
	if err != nil {
		log.Fatalf("Error trying to send IMEI %v", err)
	}

	log.Printf("DEBUG: %d bytes sent", n)
	for i := uint(0); i < *numReadings || *numReadings == 0; i++ {
		randomReading := CreateRandReadingBytes()
		log.Printf("DEBUG: [%d] sending reading %v to server", i, randomReading)
		n, err := conn.Write(randomReading[:])
		log.Printf("DEBUG: [%d] %d bytes sent", i, n)
		if err != nil {
			log.Fatalf("Error trying to send reading %v", err)
		}
		time.Sleep(readingRate)
	}

}
