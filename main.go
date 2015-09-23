package main

import (
  "fmt"
  "net/http"
  "io/ioutil"
  log "github.com/Sirupsen/logrus"
  "github.com/coreos/go-omaha/omaha"
  "encoding/xml"
)

const CoreOSAppId = "{e96281a6-d1af-4bde-9a0a-97b76e56dc57}"

func main() {
  log.SetLevel(log.DebugLevel)

  http.HandleFunc("/coreos:latest", coreosHandler)
  http.HandleFunc("/update", rootHandler)
  http.ListenAndServe(":8080", nil)
}

func coreosHandler(w http.ResponseWriter, r *http.Request) {
  defer r.Body.Close()
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
    log.Errorf("Client '%v' tried to %v services", r.RemoteAddr, n)
    http.Error(w, "I can handle only 1 app update.", 400)
  }

  applicationUpdate(w, r, reqStructure.Apps[0])

  log.Debugf("%#v", reqStructure)

  http.Error(w, "ok", 200)
  w.Header().Set("Content-Type", "application/xml")
}

func applicationUpdate(w http.ResponseWriter, r *http.Request, app *omaha.App) {
  if app.Id != CoreOSAppId {
    s := fmt.Sprintf("Unknown app ID '%v'.", app.Id)
    log.Errorf("Client '%v' tried to update service '%v'.", r.RemoteAddr, app.Id)
    http.Error(w, s, 400)
  }

  resp := omaha.NewResponse("coreos-update.protonet.info")
  newApp := resp.AddApp(CoreOSAppId)
  updateCheck := newApp.AddUpdateCheck()
  updateCheck.Status = "ok"
  updateCheck.AddUrl("http://coreos-update.protonet.info/coreos:latest")
  manifest := updateCheck.AddManifest("1.0.2")

  data, err := xml.Marshal(resp)
  if err != nil {
    log.Error(err.Error())
    http.Error(w, "An internal error occured while marshalling a response", 500)
  }

  w.Header().Set("Content-Type", "application/xml")
  w.Write(data)
}
