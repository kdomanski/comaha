package main

import (
	"io/ioutil"
	"os"
	"path"
	"testing"
)

var testdata1 = []byte("g273urgdb2397gv237e8dv823xg12397evg129p8egp32098dn43712f975y234-f-249fyb r8f34y298fy2b42dgd2h38rgd423brccf02834rty0234ftcb pg9bc34tr0203r023p r'bc347-tr374t-03y49try493p34")
var testdata2 = []byte("f-2- b735t-f43b5t423785tb723r0834dxrtvdox24rtv2o8346rto834tfvw9o5tl9w3ltlcb498tfb8w349cxtw44xctwcrp3 9bwrtv cpwc934fyc-3w4\t[wcu tn-v3n4v-t4390ty3w479rtc9pw34pt97wr43pt8rq238ctr8b2q3tq9r923432423")

func TestStorage(t *testing.T) {
	tempdir, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempdir)

	b := newLocalFileBackend(tempdir)

	// first file
	id1, err := b.Store(testdata1)
	if err != nil {
		t.Fatal(err)
	}
	path1 := path.Join(tempdir, id1)

	stat1, err := os.Stat(path1)
	if os.IsNotExist(err) {
		t.Fatalf("Should exist but doesn't: '%v'", path1)
	}
	if stat1.Size() != int64(len(testdata1)) {
		t.Fatalf("File '%v' len: %v!=%v", path1, stat1.Size(), len(testdata1))
	}

	// second file
	id2, err := b.Store(testdata2)
	if err != nil {
		t.Fatal(err)
	}
	path2 := path.Join(tempdir, id2)

	stat2, err := os.Stat(path2)
	if os.IsNotExist(err) {
		t.Fatalf("Should exist but doesn't: '%v'", path2)
	}
	if stat2.Size() != int64(len(testdata2)) {
		t.Fatalf("File '%v' len: %v!=%v", path2, stat2.Size(), len(testdata2))
	}

	// delete first, second should remain
	err = b.Delete(id1)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(path1); !os.IsNotExist(err) {
		t.Fatalf("Shouldn't exist but does: '%v'", path1)
	}

	if _, err := os.Stat(path2); os.IsNotExist(err) {
		t.Fatalf("Should exist but doesn't: '%v'", path2)
	}
}
