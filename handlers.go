package main

import (
	"crypto/sha1"
	"crypto/sha256"
	"encoding/base64"
	"encoding/xml"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/coreos/go-omaha/omaha"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
)

const noupdateResponse = `
<?xml version="1.0" encoding="UTF-8"?>
<response protocol="3.0" server="update.core-os.net">
<daystart elapsed_seconds="0"></daystart>
<app appid="e96281a6-d1af-4bde-9a0a-97b76e56dc57" status="ok">
<updatecheck status="noupdate"></updatecheck>
</app>
</response>`

func homeHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	http.Error(w, "", 404)

	log.Infof("Someone tried to access '%s'", r.URL.String())
}

func fileHandler(w http.ResponseWriter, r *http.Request) {
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
	versionString := r.URL.Query().Get("version")
	if versionString == "" {
		http.Error(w, "Missing parameter 'version'", 400)
		return
	}
	data := make([]byte, size)
	rcvsize, err := io.ReadFull(r.Body, data)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	versionData, err := parseVersionString(versionString)
	if err != nil {
		s := fmt.Sprintf("Could not parse 'version': %v", err.Error())
		http.Error(w, s, 400)
		return
	}

	log.Debugf("addPayloadHandler: received size is %v", rcvsize)

	rawSha1 := sha1.Sum(data)
	calculatedSha1 := base64.StdEncoding.EncodeToString(rawSha1[:])
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
	err = db.AddPayload(id, calculatedSha1, calculatedSha256, size, versionData)
	if err != nil {
		log.Errorf("addPayloadHandler: adding payload to db: %v", err.Error())
	}
}

func shutdownHandler(w http.ResponseWriter, r *http.Request) {
	os.Exit(0)
}

func updateHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	log.Infof("Handling an update request from %v", r.RemoteAddr)

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), 500)
		log.Error(err.Error())
	}

	var reqStructure omaha.Request
	log.Debugf("%v", string(body[:len(body)]))
	err = xml.Unmarshal(body, &reqStructure)

	if n := len(reqStructure.Apps); n != 1 {
		log.Errorf("Client '%v' tried to update %v services", r.RemoteAddr, n)
		http.Error(w, "I can handle only 1 app update.", 400)
		return
	}

	log.Debugf("%#v", reqStructure)
	applicationUpdate(w, r, reqStructure.Apps[0])

}

func applicationUpdate(w http.ResponseWriter, r *http.Request, app *omaha.App) {
	if app.Id != coreOSAppID {
		s := fmt.Sprintf("Unknown app ID '%v'.", app.Id)
		log.Errorf("Client '%v' tried to update service '%v'.", r.RemoteAddr, app.Id)
		http.Error(w, s, 400)
		return
	}

	payload := backend.GetPayload("user", "channel", "version")

	size := payload.Size
	sha1 := payload.SHA1
	sha256 := payload.SHA256
	id := payload.Url

	resp := omaha.NewResponse("coreos-update.protonet.info")
	newApp := resp.AddApp(coreOSAppID)
	newApp.Status = "ok"
	updateCheck := newApp.AddUpdateCheck()
	updateCheck.Status = "ok"
	updateCheck.AddUrl("http://10.0.2.2:8080/file?id=")
	//updateCheck.AddUrl("http://coreos-update.protorz.net:8080/coreos:latest/")
	manifest := updateCheck.AddManifest("1.0.2")
	manifest.AddPackage(sha1, id, strconv.FormatInt(size, 10), true)
	action := manifest.AddAction("postinstall")
	action.Sha256 = sha256
	action.DisablePayloadBackoff = true
	//action.MetadataSignatureRsa = "ixi6Oebo"
	//action.MetadataSize = "190"

	//data, err := xml.Marshal(resp)
	data, err := xml.MarshalIndent(resp, "", "  ")
	if err != nil {
		log.Error(err.Error())
		http.Error(w, "An internal error occured while marshalling a response", 500)
	}

	w.Header().Set("Content-Type", "application/xml")
	w.Write(data)
}
