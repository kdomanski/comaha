package main

import (
	"crypto/sha1"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"io"
	"net/http"
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
	receivedSha256 := r.URL.Query().Get("sha256")
	size := r.ContentLength
	data := make([]byte, size)
	_, err := io.ReadFull(r.Body, data)
	if err != nil {
		http.Error(w, err.Error(), 500)
	}

	rawSha1 := sha1.Sum(data)
	calculatedSha1 := base64.StdEncoding.EncodeToString(rawSha1[:])
	rawSha256 := sha256.Sum256(data)
	calculatedSha256 := base64.StdEncoding.EncodeToString(rawSha256[:])

	if receivedSha1 != calculatedSha1 {
		s := fmt.Sprintf("SHA1 validation failed, '%v'!='%v'", receivedSha1, calculatedSha1)
		http.Error(w, s, 400)
	}

	if receivedSha256 != calculatedSha256 {
		s := fmt.Sprintf("SHA256 validation failed, '%v'!='%v'", receivedSha256, calculatedSha256)
		http.Error(w, s, 400)
	}

	// TODO actually add image

	w.Write(nil)
}
