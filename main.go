package main

import (
	"encoding/xml"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/coreos/go-omaha/omaha"
	"io/ioutil"
	"net/http"
)

const coreOSAppID = "{e96281a6-d1af-4bde-9a0a-97b76e56dc57}"

func main() {
	log.SetLevel(log.DebugLevel)

	http.HandleFunc("/coreos:latest/update.gz", coreosHandler)
	http.HandleFunc("/update", rootHandler)
	http.HandleFunc("/", homeHandler)
	http.ListenAndServe(":8080", nil)
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	http.Error(w, "", 404)

	log.Infof("Someone tried to access '%s'", r.URL.String())
}

func coreosHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	http.ServeFile(w, r, "update.gz")
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

	resp := omaha.NewResponse("coreos-update.protonet.info")
	newApp := resp.AddApp(coreOSAppID)
	newApp.Status = "ok"
	updateCheck := newApp.AddUpdateCheck()
	updateCheck.Status = "ok"
	updateCheck.AddUrl("http://172.17.8.12:8080/coreos:latest/")
	//updateCheck.AddUrl("http://coreos-update.protorz.net:8080/coreos:latest/")
	manifest := updateCheck.AddManifest("1.0.2")
	manifest.AddPackage("328d2d14facf805b3508afc4d315f784c41e62c4", "update.gz", "123456", true)
	action := manifest.AddAction("postinstall")
	action.Sha256 = "b602d630f0a081840d0ca8fc4d35810e42806642b3127bb702d65c3df227d0f5"
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
