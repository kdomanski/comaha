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

func (v payloadVersion) String() string {
	if v.timestamp.Unix() == 0 {
		return fmt.Sprintf("%v.%v.%v", v.build, v.branch, v.patch)
	} else {
		return fmt.Sprintf("%v.%v.%v+%v", v.build, v.branch, v.patch, v.timestamp.Format(timelayout))
	}
}

func (v payloadVersion) IsGreater(other payloadVersion) bool {
	switch {
	case v.build > other.build:
		return true
	case v.build < other.build:
		return false
	}

	switch {
	case v.branch > other.branch:
		return true
	case v.branch < other.branch:
		return false
	}

	switch {
	case v.patch > other.patch:
		return true
	case v.patch < other.patch:
		return false
	}

	vTime := v.timestamp.UTC().Unix()
	otherTime := other.timestamp.UTC().Unix()

	if vTime > otherTime {
		return true
	}

	return false
}

func (v payloadVersion) IsEqual(other payloadVersion) bool {
	return v.build == other.build &&
		v.branch == other.branch &&
		v.patch == other.patch &&
		v.timestamp.UTC().Unix() == other.timestamp.UTC().Unix()
}

func (v payloadVersion) IsZero(other payloadVersion) bool {
	return v.build == 0 &&
		v.branch == 0 &&
		v.patch == 0
}
