package main

import (
	"crypto/sha1"
	"encoding/base64"
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

	calculatedSha1 := base64.StdEncoding.EncodeToString(sha1.Sum(data))
}
