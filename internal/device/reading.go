package device

import (
	"encoding/binary"
	"math"
	"math/rand"
)

// Reading is the set of device readings.
type Reading struct {
	// Temperature denotes the temperature reading of the message.
	Temperature float64

	// Altitude denotes the altitude reading of the message.
	Altitude float64

	// Latitude denotes the latitude reading of the message.
	Latitude float64

	// Longitude denotes the longitude reading of the message.
	Longitude float64

	// BatteryLevel denotes the battery level reading of the message.
	BatteryLevel float64
}

const (
	temperatureMin  = -300
	temperatureMax  = 300
	altitudeMin     = -20000
	altitudMax      = 20000
	latitudeMin     = -90
	latitudeMax     = 90
	longitudeMin    = -180
	longitudeMax    = 180
	batteryLevelMin = 0
	batteryLevelMax = 100
)

// Decode decodes the reading message payload in the given b into r.
//
// If any of the fields are outside their valid min/max ranges ok will be unset.
//
// Decode does NOT allocate under any condition. Additionally, it panics if b
// isn't at least 40 bytes long.
func (r *Reading) Decode(b []byte) (ok bool) {
	_ = b[39] // compiler bound check hint
	temperature := math.Float64frombits(binary.BigEndian.Uint64(b[0:]))
	altitude := math.Float64frombits(binary.BigEndian.Uint64(b[8:]))
	latitude := math.Float64frombits(binary.BigEndian.Uint64(b[16:]))
	longitude := math.Float64frombits(binary.BigEndian.Uint64(b[24:]))
	batteryLevel := math.Float64frombits(binary.BigEndian.Uint64(b[32:]))

	if !allFieldsAreValid(temperature, altitude, latitude, longitude, batteryLevel) {
		return false
	}

	r.Temperature = temperature
	r.Altitude = altitude
	r.Latitude = latitude
	r.Longitude = longitude
	r.BatteryLevel = batteryLevel

	return true
}

func allFieldsAreValid(temperature, altitude, latitude, longitude, batteryLevel float64) bool {
	validTemperature := isInRange(temperature, temperatureMin, temperatureMax)
	validAltitude := isInRange(altitude, altitudeMin, altitudMax)
	validLatitude := isInRange(latitude, latitudeMin, latitudeMax)
	validLongitude := isInRange(longitude, longitudeMin, longitudeMax)
	validBatteryLevel := isInRange(batteryLevel, batteryLevelMin, batteryLevelMax)

	return validTemperature &&
		validAltitude &&
		validLatitude &&
		validLongitude &&
		validBatteryLevel

}

func isInRange(value, min, max float64) bool {
	return value >= min && value <= max
}

// CreateRandReading creates a new random bytes payload for testing
func CreateRandReadingBytes() [40]byte {
	return NewPayload(
		randFloat(temperatureMin, temperatureMax),
		randFloat(altitudeMin, altitudMax),
		randFloat(latitudeMin, latitudeMax),
		randFloat(longitudeMin, longitudeMax),
		randFloat(batteryLevelMin, batteryLevelMax),
	)

}

func randFloat(min, max float64) float64 {
	return min + rand.Float64()*(max-min)
}

// NewPayload returns a input reading values as byte array
func NewPayload(expectedTemperature float64, expectedAltitude float64, expectedLatitude float64, expectedLongitude float64, expectedBatteryLevel float64) [40]byte {
	var payload [40]byte
	binary.BigEndian.PutUint64(payload[0:], math.Float64bits(expectedTemperature))
	binary.BigEndian.PutUint64(payload[8:], math.Float64bits(expectedAltitude))
	binary.BigEndian.PutUint64(payload[16:], math.Float64bits(expectedLatitude))
	binary.BigEndian.PutUint64(payload[24:], math.Float64bits(expectedLongitude))
	binary.BigEndian.PutUint64(payload[32:], math.Float64bits(expectedBatteryLevel))
	return payload
}
