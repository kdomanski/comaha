package main

import (
	"crypto/sha1"
	"crypto/sha256"
	"encoding/base64"
	"encoding/xml"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/coreos/go-omaha/omaha"
	"github.com/julienschmidt/httprouter"
	"html/template"
	"io"
	"io/ioutil"
	"net/http"
	"path"
	"runtime"
)

const noupdateResponse = `
<?xml version="1.0" encoding="UTF-8"?>
<response protocol="3.0" server="update.core-os.net">
<daystart elapsed_seconds="0"></daystart>
<app appid="e96281a6-d1af-4bde-9a0a-97b76e56dc57" status="ok">
<updatecheck status="noupdate"></updatecheck>
</app>
</response>`

func homeHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	defer r.Body.Close()
	http.Error(w, "", 404)

	log.Infof("Someone tried to access '%s'", r.URL.String())
}

func fileHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	defer r.Body.Close()

	log.Infof("Someone tried to access '%s'", r.URL.String())

	fileid := r.URL.Query().Get("id")
	log.Infof("Handling request for %v", fileid)
	http.ServeFile(w, r, path.Join("storage", fileid))
}

func addPayloadHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	defer runtime.GC()
	defer r.Body.Close()

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
	channel := r.URL.Query().Get("channel")
	if channel == "" {
		http.Error(w, "Missing parameter 'channel'", 400)
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

	err = db.AttachPayloadToChannel(id, channel)
	if err != nil {
		log.Errorf("addPayloadHandler: adding payload to channel: %v", err.Error())
	}
}

func deletePayloadHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	defer r.Body.Close()

	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "Missing parameter 'id'", 400)
		return
	}

	err := db.DeletePayload(id)
	if err != nil {
		log.Errorf("deletePayloadHandler: removing DB entry for '%v': %v", id, err.Error())
		http.Error(w, err.Error(), 500)
	}

	err = fileBE.Delete(id)
	if err != nil {
		log.Errorf("deletePayloadHandler: removing file for '%v': %v", id, err.Error())
		http.Error(w, err.Error(), 500)
	}
}

func channelForceDowngradeGetHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	channel := ps.ByName("channel")
	value, err := db.GetChannelForceDowngrade(channel)
	if err != nil {
		log.Errorf("channelForceDowngradeGetHandler: getting FD value for channel '%v': %v", channel, err.Error())
		http.Error(w, err.Error(), 500)
		return
	}

	if value {
		fmt.Fprint(w, 1)
	} else {
		fmt.Fprint(w, 0)
	}
}

func channelForceDowngradePostHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	channel := ps.ByName("channel")

	defer r.Body.Close()
	body, err := ioutil.ReadAll(r.Body)
	value := string(body)

	var boolValue bool
	switch value {
	case "1":
		boolValue = true
	case "0":
		boolValue = false
	default:
		s := fmt.Sprintf("Invalid value '%v'", value)
		http.Error(w, s, http.StatusBadRequest)
	}

	err = db.SetChannelForceDowngrade(channel, boolValue)
	if err != nil {
		log.Error(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func updateHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
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

	logContext := log.WithFields(log.Fields{
		"remoteAddr": r.RemoteAddr,
	})

	// protocol and hostname for local storage URL building
	scheme := "http"
	if h := r.Header.Get("X-Forwarded-Proto"); h != "" {
		scheme = h
	}
	localUrl := fmt.Sprintf("%v://%v", scheme, r.Host)

	resp := omaha.NewResponse(r.Host)
	for _, appReq := range reqStructure.Apps {
		appResponse := resp.AddApp(appReq.Id)
		handleApiApp(logContext, localUrl, appReq, appResponse)
	}

	data, err := xml.MarshalIndent(resp, "", "  ")
	if err != nil {
		log.Error(err.Error())
		http.Error(w, "An internal error occured while marshalling a response", 500)
	}

	w.Header().Set("Content-Type", "application/xml")
	w.Write(data)
}

func panelHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	defer r.Body.Close()

	// TODO log panel access

	data, err := ioutil.ReadFile("static/images.html")
	if err != nil {
		http.Error(w, err.Error(), 500)
		log.Error(err.Error())
		return
	}

	funcMap := template.FuncMap{
		"toMB": func(i int64) string {
			divided := float32(i) / 1048576
			formatted := fmt.Sprintf("%.1f", divided)
			return formatted
		},
	}

	t, err := template.New("images").Funcs(funcMap).Parse(string(data))
	if err != nil {
		http.Error(w, err.Error(), 500)
		log.Error(err.Error())
		return
	}

	channels, err := db.ListChannels()
	if err != nil {
		log.Error(err.Error())
		http.Error(w, "Failed to retrieve list of channels", 500)
		return
	}

	var chosenChannel string
	var forceDowngrade bool
	var images []payload
	var events []Event

	if _, ok := r.URL.Query()["events"]; ok {
		events, err = db.GetEvents()
		if err != nil {
			log.Error(err.Error())
			http.Error(w, "Failed to retrieve events from the database", 500)
			return
		}
	} else {
		chosenChannel = r.URL.Query().Get("channel")
		if chosenChannel == "" && len(channels) > 0 {
			chosenChannel = channels[0]
		}

		forceDowngrade, err = db.GetChannelForceDowngrade(chosenChannel)
		if err != nil {
			log.Error(err.Error())
			http.Error(w, "Failed to retrieve force_downgrade option for the channel", 500)
			return
		}

		images, err = db.ListImages(chosenChannel)
		if err != nil {
			log.Error(err.Error())
			http.Error(w, "Failed to retrieve images for the channel", 500)
			return
		}
	}

	panelData := struct {
		Images         []payload
		Events         []Event
		Channels       []string
		CurrentChannel string
		ForceDowngrade bool
	}{
		images,
		events,
		channels,
		chosenChannel,
		forceDowngrade,
	}

	err = t.Execute(w, panelData)
	if err != nil {
		log.Errorf("Error parsing panel template: %v", err)
		http.Error(w, "Error parsing panel template", 500)
		return
	}
}
