package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rilopez/redis-wire-protocol/internal/common"
	"github.com/rilopez/redis-wire-protocol/internal/device"
)

func TestHttpd_StatsHandler(t *testing.T) {
	core := newCore(common.FrozenInTime, uint(1337), 2)
	httpd := newHttpd(core, 80)
	req, err := http.NewRequest("GET", "/stats", nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(httpd.statsHandler)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	jsonMap, errJSONParse := parseJSONAsMap(rr.Body.String())
	if errJSONParse != nil {
		t.Error(errJSONParse)
	}

	assertJSONMapHasField(t, jsonMap, "numConnectedClients")
	assertJSONMapHasField(t, jsonMap, "numCpu")
	assertJSONMapHasField(t, jsonMap, "numGoroutine")
	assertJSONMapHasField(t, jsonMap, "memStats")
}

func TestHttpd_StatusHandler(t *testing.T) {
	core := newCore(common.FrozenInTime, uint(1337), 2)
	httpd := newHttpd(core, 80)
	expectedIMEI := uint64(448324242329542)
	callBackChannel := make(chan common.Command, 1)
	err := core.register(expectedIMEI, callBackChannel)
	if err != nil {
		t.Error(err)
	}

	url := fmt.Sprintf("/status/%d", expectedIMEI)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(httpd.statusHandler)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	jsonMap, errJSONParse := parseJSONAsMap(rr.Body.String())
	if errJSONParse != nil {
		t.Error(errJSONParse)
	}

	assertJSONMapHasField(t, jsonMap, "online")

}

func TestHttpd_ReadingHandler(t *testing.T) {
	core := newCore(common.FrozenInTime, uint(1337), 2)
	httpd := newHttpd(core, 80)
	expectedIMEI := uint64(448324242329542)
	reading := &device.Reading{}
	randomReadingBytes := device.CreateRandReadingBytes()
	reading.Decode(randomReadingBytes[:])

	callBackChannel := make(chan common.Command, 1)
	err := core.register(expectedIMEI, callBackChannel)
	if err != nil {
		t.Error(err)
	}
	dev, _ := core.clientByID(expectedIMEI)
	dev.lastReading = reading

	url := fmt.Sprintf("/readings/%d", expectedIMEI)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(httpd.readingsHandler)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	jsonMap, errJSONParse := parseJSONAsMap(rr.Body.String())
	if errJSONParse != nil {
		t.Error(errJSONParse)
	}

	assertJSONMapHasField(t, jsonMap, "reading")
	assertJSONMapHasField(t, jsonMap, "timestampEpoch")

}

func TestImeiFromPath(t *testing.T) {

	expectedIMEI := uint64(448324242329542)

	actualIMEI, err := imeiFromPath("/clients/448324242329542", "/clients/")
	if err != nil {
		t.Fatal(err)
	}
	if expectedIMEI != actualIMEI {
		t.Errorf("expcted IMEI: %d, actual: %d", expectedIMEI, actualIMEI)
	}
}

func assertJSONMapHasField(t *testing.T, jsonMap map[string]interface{}, fieldName string) {
	if _, exists := jsonMap[fieldName]; !exists {
		t.Errorf("field %s not found in json %v", fieldName, jsonMap)
	}
}

func parseJSONAsMap(jsonString string) (map[string]interface{}, error) {
	var jsonMap map[string]interface{}
	err := json.Unmarshal([]byte(jsonString), &jsonMap)
	return jsonMap, err

}
