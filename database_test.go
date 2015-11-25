package main

import (
	"database/sql"
	"reflect"
	"sort"
	"testing"
	"time"
)

type addTestElement struct {
	ID      string
	SHA1    string
	SHA256  string
	Size    int64
	Version payloadVersion
}

func TestDBAdding(t *testing.T) {
	testData := []addTestElement{
		{"", "", "", 0, payloadVersion{build: 766, branch: 4, patch: 1, timestamp: time.Unix(0, 0).UTC()}},
		{"", "", "", 6236325, payloadVersion{build: 766, branch: 4, patch: 1, timestamp: time.Unix(0, 0).UTC()}},
		{"fq435r34qd34r", "235r3a2q23r3fa32", "af3fa32fa3", 0, payloadVersion{build: 766, branch: 4, patch: 1, timestamp: time.Unix(0, 0).UTC()}},
		{"fq435r34qd34r", "235r3a2q23r3fa32", "af3fa32fa3", 135235413242, payloadVersion{}},
	}

	db, err := newUserDB(":memory:")
	if err != nil {
		t.Errorf("newUserDB: %v", err.Error())
	}

	for _, datum := range testData {
		err = db.AddPayload(datum.ID, datum.SHA1, datum.SHA256, datum.Size, datum.Version)
	}

}

func TestDBListChannels1(t *testing.T) {
	db, err := newUserDB(":memory:")
	if err != nil {
		t.Errorf("newUserDB: %v", err.Error())
	}

	// 1 image per channel
	db.AddPayload("foo", "bar", "foobar", 1234, payloadVersion{build: 766, branch: 4, patch: 1, timestamp: time.Unix(0, 0).UTC()})
	db.AttachPayloadToChannel("foo", "channel1")
	db.AddPayload("xyz", "abc", "uvw", 7423, payloadVersion{build: 800, branch: 1, patch: 2, timestamp: time.Unix(0, 0).UTC()})
	db.AttachPayloadToChannel("xyz", "channel2")

	chans, err := db.ListChannels()
	if err != nil {
		t.Errorf("ListChannels: %v", err.Error())
	}
	if n := len(chans); n != 2 {
		t.Errorf("Expected 2 channels, got %v", n)
	}

	sort.Strings(chans)
	expectedChans := []string{"channel1", "channel2"}
	if !reflect.DeepEqual(chans, expectedChans) {
		t.Errorf("Expected channels %+v, got %+v", expectedChans, chans)
	}
}

func TestDBListChannels2(t *testing.T) {
	db, err := newUserDB(":memory:")
	if err != nil {
		t.Errorf("newUserDB: %v", err.Error())
	}

	// 2 images per channel
	db.AddPayload("foo", "bar", "foobar", 1234, payloadVersion{build: 766, branch: 4, patch: 1, timestamp: time.Unix(0, 0).UTC()})
	db.AttachPayloadToChannel("foo", "channel1")
	db.AddPayload("4r12f", "da23d", "d21c", 6143, payloadVersion{})
	db.AttachPayloadToChannel("4r12f", "channel1")
	db.AddPayload("xyz", "abc", "uvw", 7423, payloadVersion{build: 800, branch: 1, patch: 2, timestamp: time.Unix(0, 0).UTC()})
	db.AttachPayloadToChannel("xyz", "channel2")
	db.AddPayload("d41234d321", "12d34", "1234", 533453, payloadVersion{build: 412, branch: 4, patch: 2143, timestamp: time.Unix(2142, 0).UTC()})
	db.AttachPayloadToChannel("d41234d321", "channel2")

	chans, err := db.ListChannels()
	if err != nil {
		t.Errorf("ListChannels: %v", err.Error())
	}
	if n := len(chans); n != 2 {
		t.Errorf("Expected 2 channels, got %v", n)
	}

	sort.Strings(chans)
	expectedChans := []string{"channel1", "channel2"}
	if !reflect.DeepEqual(chans, expectedChans) {
		t.Errorf("Expected channels %+v, got %+v", expectedChans, chans)
	}
}

func TestDBListChannels3(t *testing.T) {
	db, err := newUserDB(":memory:")
	if err != nil {
		t.Errorf("newUserDB: %v", err.Error())
	}

	// 2 images without channel
	db.AddPayload("foo", "bar", "foobar", 1234, payloadVersion{build: 766, branch: 4, patch: 1, timestamp: time.Unix(0, 0).UTC()})
	db.AttachPayloadToChannel("foo", "channel1")
	db.AddPayload("xyz", "abc", "uvw", 7423, payloadVersion{build: 800, branch: 1, patch: 2, timestamp: time.Unix(0, 0).UTC()})
	db.AttachPayloadToChannel("xyz", "channel2")
	db.AddPayload("4r12f", "da23d", "d21c", 6143, payloadVersion{})
	db.AddPayload("d41234d321", "12d34", "1234", 533453, payloadVersion{build: 412, branch: 4, patch: 2143, timestamp: time.Unix(2142, 0).UTC()})

	chans, err := db.ListChannels()
	if err != nil {
		t.Errorf("ListChannels: %v", err.Error())
	}
	if n := len(chans); n != 2 {
		t.Errorf("Expected 2 channels, got %v", n)
	}

	sort.Strings(chans)
	expectedChans := []string{"channel1", "channel2"}
	if !reflect.DeepEqual(chans, expectedChans) {
		t.Errorf("Expected channels %+v, got %+v", expectedChans, chans)
	}
}

func TestDBListImages(t *testing.T) {
	db, err := newUserDB(":memory:")
	if err != nil {
		t.Errorf("newUserDB: %v", err.Error())
	}

	// 2 images in channel 1, 1 in channel 2, fourth without channel
	db.AddPayload("foo", "bar", "foobar", 1234, payloadVersion{build: 766, branch: 4, patch: 1, timestamp: time.Unix(0, 0).UTC()})
	db.AttachPayloadToChannel("foo", "channel1")
	db.AddPayload("xyz", "abc", "uvw", 7423, payloadVersion{build: 800, branch: 1, patch: 2, timestamp: time.Unix(0, 0).UTC()})
	db.AttachPayloadToChannel("xyz", "channel1")
	db.AddPayload("4r12f", "da23d", "d21c", 6143, payloadVersion{timestamp: time.Unix(0, 0).UTC()})
	db.AttachPayloadToChannel("4r12f", "channel2")
	db.AddPayload("d41234d321", "12d34", "1234", 533453, payloadVersion{build: 412, branch: 4, patch: 2143, timestamp: time.Unix(2142, 0).UTC()})

	imgs1, err := db.ListImages("channel1")
	if err != nil {
		t.Errorf("ListImages: %v", err.Error())
	}
	if n := len(imgs1); n != 2 {
		t.Errorf("Expected 2 images, got %v", n)
	}

	imgs2, err := db.ListImages("channel2")
	if err != nil {
		t.Errorf("ListImages: %v", err.Error())
	}
	if n := len(imgs2); n != 1 {
		t.Errorf("Expected 1 image, got %v", n)
	}
	if testEl := (payload{ID: "4r12f", Version: "0.0.0", SHA1: "da23d", SHA256: "d21c", Size: 6143}); imgs2[0] != testEl {
		t.Errorf("Expected image %+v, got %+v", testEl, imgs2[0])
	}

	chans, err := db.ListChannels()
	if err != nil {
		t.Errorf("ListChannels: %v", err.Error())
	}
	if n := len(chans); n != 2 {
		t.Errorf("Expected 2 channels, got %v", n)
	}

	sort.Strings(chans)
	expectedChans := []string{"channel1", "channel2"}
	if !reflect.DeepEqual(chans, expectedChans) {
		t.Errorf("Expected channels %+v, got %+v", expectedChans, chans)
	}
}

// behind the latest version from 'channel1'
func TestDBLGetNewerPayload1(t *testing.T) {
	db, err := newUserDB(":memory:")
	if err != nil {
		t.Errorf("newUserDB: %v", err.Error())
	}

	// 2 images in channel 1, 1 in channel 2, fourth without channel
	db.AddPayload("foo", "bar", "foobar", 1234, payloadVersion{build: 766, branch: 4, patch: 1, timestamp: time.Unix(0, 0).UTC()})
	db.AttachPayloadToChannel("foo", "channel1")
	db.AddPayload("xyz", "abc", "uvw", 7423, payloadVersion{build: 800, branch: 1, patch: 2, timestamp: time.Unix(0, 0).UTC()})
	db.AttachPayloadToChannel("xyz", "channel1")
	db.AddPayload("4r12f", "da23d", "d21c", 6143, payloadVersion{build: 820, branch: 1, patch: 2, timestamp: time.Unix(0, 0).UTC()})
	db.AttachPayloadToChannel("4r12f", "channel2")
	db.AddPayload("d41234d321", "12d34", "1234", 533453, payloadVersion{build: 1000, branch: 4, patch: 2143, timestamp: time.Unix(2142, 0).UTC()})

	ver, err := parseVersionString("766.6.0")
	if err != nil {
		t.Errorf("parseVersionString: %v", err.Error())
	}
	pl, err := db.GetNewerPayload(ver, "channel1")
	if err != nil {
		t.Errorf("GetNewerPayload: %v", err.Error())
	}

	if testPl := (payload{SHA1: "abc", SHA256: "uvw", Size: 7423, ID: "xyz"}); pl != testPl {
		t.Errorf("Expected payload %+v, got %+v", testPl, pl)
	}

}

// at the latest version from 'channel1'
func TestDBLGetNewerPayload2(t *testing.T) {
	db, err := newUserDB(":memory:")
	if err != nil {
		t.Errorf("newUserDB: %v", err.Error())
	}

	// 2 images in channel 1, 1 in channel 2, fourth without channel
	db.AddPayload("foo", "bar", "foobar", 1234, payloadVersion{build: 766, branch: 4, patch: 1, timestamp: time.Unix(0, 0).UTC()})
	db.AttachPayloadToChannel("foo", "channel1")
	db.AddPayload("xyz", "abc", "uvw", 7423, payloadVersion{build: 800, branch: 1, patch: 2, timestamp: time.Unix(0, 0).UTC()})
	db.AttachPayloadToChannel("xyz", "channel1")
	db.AddPayload("4r12f", "da23d", "d21c", 6143, payloadVersion{build: 820, branch: 1, patch: 2, timestamp: time.Unix(0, 0).UTC()})
	db.AttachPayloadToChannel("4r12f", "channel2")
	db.AddPayload("d41234d321", "12d34", "1234", 533453, payloadVersion{build: 1000, branch: 4, patch: 2143, timestamp: time.Unix(2142, 0).UTC()})

	ver, err := parseVersionString("800.1.2")
	if err != nil {
		t.Errorf("parseVersionString: %v", err.Error())
	}
	_, err = db.GetNewerPayload(ver, "channel1")
	if err != sql.ErrNoRows {
		t.Errorf("GetNewerPayload should have failed with sql.ErrNoRows")
	}
}

// ahead of the latest version from 'channel1'
func TestDBLGetNewerPayload3(t *testing.T) {
	db, err := newUserDB(":memory:")
	if err != nil {
		t.Errorf("newUserDB: %v", err.Error())
	}

	// 2 images in channel 1, 1 in channel 2, fourth without channel
	db.AddPayload("foo", "bar", "foobar", 1234, payloadVersion{build: 766, branch: 4, patch: 1, timestamp: time.Unix(0, 0).UTC()})
	db.AttachPayloadToChannel("foo", "channel1")
	db.AddPayload("xyz", "abc", "uvw", 7423, payloadVersion{build: 800, branch: 1, patch: 2, timestamp: time.Unix(0, 0).UTC()})
	db.AttachPayloadToChannel("xyz", "channel1")
	db.AddPayload("4r12f", "da23d", "d21c", 6143, payloadVersion{build: 820, branch: 1, patch: 2, timestamp: time.Unix(0, 0).UTC()})
	db.AttachPayloadToChannel("4r12f", "channel2")
	db.AddPayload("d41234d321", "12d34", "1234", 533453, payloadVersion{build: 1000, branch: 4, patch: 2143, timestamp: time.Unix(2142, 0).UTC()})

	ver, err := parseVersionString("812.0.0")
	if err != nil {
		t.Errorf("parseVersionString: %v", err.Error())
	}
	_, err = db.GetNewerPayload(ver, "channel1")
	if err != sql.ErrNoRows {
		t.Errorf("GetNewerPayload should have failed with sql.ErrNoRows")
	}
}

func TestDBDeleting1(t *testing.T) {
	db, err := newUserDB(":memory:")
	if err != nil {
		t.Errorf("newUserDB: %v", err.Error())
	}

	db.AddPayload("foo", "bar", "foobar", 1234, payloadVersion{build: 766, branch: 4, patch: 1, timestamp: time.Unix(0, 0).UTC()})
	db.AttachPayloadToChannel("foo", "channel1")
	db.AddPayload("4r12f", "da23d", "d21c", 6143, payloadVersion{})
	db.AttachPayloadToChannel("4r12f", "channel1")
	db.AddPayload("xyz", "abc", "uvw", 7423, payloadVersion{build: 800, branch: 1, patch: 2, timestamp: time.Unix(0, 0).UTC()})
	db.AttachPayloadToChannel("xyz", "channel2")
	db.AddPayload("d41234d321", "12d34", "1234", 533453, payloadVersion{build: 412, branch: 4, patch: 2143, timestamp: time.Unix(2142, 0).UTC()})
	db.AttachPayloadToChannel("d41234d321", "channel2")

	db.DeletePayload("d41234d321")

	pds, err := db.ListImages("channel2")
	if err != nil {
		t.Errorf("ListImages: %v", err.Error())
	}
	if n := len(pds); n != 1 {
		t.Errorf("Expected 1 image, got %v", n)
	}

	if leftID := pds[0].ID; leftID != "xyz" {
		t.Error("Expected image 'xyz', got %v", leftID)
	}
}

func TestDBDeleting2(t *testing.T) {
	db, err := newUserDB(":memory:")
	if err != nil {
		t.Errorf("newUserDB: %v", err.Error())
	}

	db.AddPayload("foo", "bar", "foobar", 1234, payloadVersion{build: 766, branch: 4, patch: 1, timestamp: time.Unix(0, 0).UTC()})
	db.AttachPayloadToChannel("foo", "channel1")
	db.AddPayload("4r12f", "da23d", "d21c", 6143, payloadVersion{})
	db.AttachPayloadToChannel("4r12f", "channel1")
	db.AddPayload("xyz", "abc", "uvw", 7423, payloadVersion{build: 800, branch: 1, patch: 2, timestamp: time.Unix(0, 0).UTC()})
	db.AttachPayloadToChannel("xyz", "channel2")
	db.AddPayload("d41234d321", "12d34", "1234", 533453, payloadVersion{build: 412, branch: 4, patch: 2143, timestamp: time.Unix(2142, 0).UTC()})
	db.AttachPayloadToChannel("d41234d321", "channel2")

	db.DeletePayload("d41234d321")
	db.DeletePayload("xyz")

	chans, err := db.ListChannels()
	if err != nil {
		t.Errorf("ListChannels: %v", err.Error())
	}
	if n := len(chans); n != 1 {
		t.Errorf("Expected 1 channel, got %v", n)
	}

	expectedChans := []string{"channel1"}
	if !reflect.DeepEqual(chans, expectedChans) {
		t.Errorf("Expected channels %+v, got %+v", expectedChans, chans)
	}
}
