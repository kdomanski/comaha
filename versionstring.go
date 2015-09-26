package main

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

type payloadVersion struct {
	build     int32
	branch    int32
	patch     int32
	timestamp time.Time
}

const timelayout = "2006-01-02-1504"

func parseVersionString(input string) (payloadVersion, error) {
	result := payloadVersion{}

	mainparts := strings.Split(input, "+")
	if len(mainparts) > 2 {
		return result, errors.New("Extraneous '+' signs")
	}

	if len(mainparts) == 2 {
		t, err := time.Parse(timelayout, mainparts[1])
		if err != nil {
			return result, err
		}
		result.timestamp = t.UTC()
	} else {
		result.timestamp = time.Unix(0, 0).UTC()
	}

	intparts := strings.Split(mainparts[0], ".")
	if len(intparts) != 3 {
		s := fmt.Sprintf("Incorrect number of version segments (%v).", len(intparts))
		return result, errors.New(s)
	}

	verBuild, err := strconv.Atoi(intparts[0])
	if err != nil {
		return result, err
	}

	verBranch, err := strconv.Atoi(intparts[1])
	if err != nil {
		return result, err
	}

	verPatch, err := strconv.Atoi(intparts[2])
	if err != nil {
		return result, err
	}

	result.build = int32(verBuild)
	result.branch = int32(verBranch)
	result.patch = int32(verPatch)

	return result, nil
}
