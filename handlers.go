package main

import (
	"crypto/sha1"
	"crypto/sha256"
	"encoding/base64"
	//"encoding/hex"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"io"
	"net/http"
	"os"
)

func homeHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	http.Error(w, "", 404)

	log.Infof("Someone tried to access '%s'", r.URL.String())
}

func coreosHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	fileid := r.URL.Query().Get("id")
	http.ServeFile(w, r, fileid)
}

func addPayloadHandler(w http.ResponseWriter, r *http.Request) {
	receivedSha1 := r.URL.Query().Get("sha1")
	if receivedSha1 == "" {
		http.Error(w, "Missing parameter 'sha1'", 400)
		return
	}
	receivedSha256 := r.URL.Query().Get("sha256")
	if receivedSha256 == "" {
		http.Error(w, "Missing parameter 'sha256'", 400)
		return
	}
	size := r.ContentLength
	version := r.URL.Query().Get("version")
	if version == "" {
		http.Error(w, "Missing parameter 'version'", 400)
		return
	}
	data := make([]byte, size)
	rcvsize, err := io.ReadFull(r.Body, data)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	log.Debugf("addPayloadHandler: received size is %v", rcvsize)

	rawSha1 := sha1.Sum(data)
	calculatedSha1 := base64.StdEncoding.EncodeToString(rawSha1[:])
	//calculatedSha1 := fmt.Sprintf("%x", rawSha1)
	//calculatedSha1 := hex.EncodeToString(rawSha1[:])
	rawSha256 := sha256.Sum256(data)
	calculatedSha256 := base64.StdEncoding.EncodeToString(rawSha256[:])

	if receivedSha1 != calculatedSha1 {
		s := fmt.Sprintf("SHA1 validation failed, '%v' != '%v'", receivedSha1, calculatedSha1)
		http.Error(w, s, 400)
		return
	}

	if receivedSha256 != calculatedSha256 {
		s := fmt.Sprintf("SHA256 validation failed, '%v' != '%v'", receivedSha256, calculatedSha256)
		http.Error(w, s, 400)
		return
	}

	id, err := fileBE.Store(data)
	if err != nil {
		log.Errorf("addPayloadHandler: storing data: %v", err.Error())
	}
	err = db.AddPayload(id, calculatedSha1, calculatedSha256, size, version)
	if err != nil {
		log.Errorf("addPayloadHandler: adding payload to db: %v", err.Error())
	}
}

func shutdownHandler(w http.ResponseWriter, r *http.Request) {
	os.Exit(0)
}
