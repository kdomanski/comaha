package main

import (
	"database/sql"
	"encoding/xml"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/coreos/go-omaha/omaha"
	sqlite3 "github.com/mattn/go-sqlite3"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
)

const coreOSAppID = "{e96281a6-d1af-4bde-9a0a-97b76e56dc57}"

var backend *singleFileBackend
var db *userDB

func main() {
	log.SetLevel(log.DebugLevel)
	log.Info("COmaha update server starting")
	var err error

	sql.Register("sqlite3", &sqlite3.SQLiteDriver{})

	db, err = newUserDB("users.sqlite")
	if err != nil {
		log.Errorf("Could not open database: %v", err.Error())
		os.Exit(1)
	}

	backend, err = NewSingleFileBackend("payload-list.json")
	if err != nil {
		log.Errorf("Failed to load simple backend from 'payload-list.json': %v", err.Error())
		os.Exit(1)
	}

	http.HandleFunc("/file", coreosHandler)
	http.HandleFunc("/update", rootHandler)
	//http.HandleFunc("/admin/add_group", addGroupHandler)
	http.HandleFunc("/admin/add_payload", addPayloadHandler)
	//http.HandleFunc("/admin/add_user", addUserHandler)
	http.HandleFunc("/", homeHandler)
	http.ListenAndServe(":8080", nil)
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
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
