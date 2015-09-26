package main

import (
	"testing"
	"time"
)

type parsePairs struct {
	toParse        string
	expectedResult payloadVersion
}

func TestParsing(t *testing.T) {
	testData := []parsePairs{
		{"766.4.1", payloadVersion{build: 766, branch: 4, patch: 1, timestamp: time.Unix(0, 0).UTC()}},
		{"0.1.2", payloadVersion{build: 0, branch: 1, patch: 2, timestamp: time.Unix(0, 0).UTC()}},
		{"11111111.22222222.33333333", payloadVersion{build: 11111111, branch: 22222222, patch: 33333333, timestamp: time.Unix(0, 0).UTC()}},
		{"802.3.0+2015-01-01-0101", payloadVersion{build: 802, branch: 3, patch: 0, timestamp: time.Unix(1420074060, 0).UTC()}},
		{"802.1.2+2015-12-31-2359", payloadVersion{build: 802, branch: 1, patch: 2, timestamp: time.Unix(1451606340, 0).UTC()}},
	}

	for _, datum := range testData {
		parsed, err := parseVersionString(datum.toParse)
		if err != nil {
			t.Errorf("While parsing '%v': %v", datum.toParse, err.Error())
		}
		if parsed != datum.expectedResult {
			t.Errorf("%v != %v", parsed, datum.expectedResult)
		}
	}
}
