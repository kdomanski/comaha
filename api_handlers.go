package main

import (
	"database/sql"
	"fmt"
	"github.com/Sirupsen/logrus"
	"github.com/coreos/go-omaha/omaha"
	"strconv"
)

// parse an 'app' tag of request and generate a corresponding 'app' tag of response
func handleApiApp(logContext *logrus.Entry, appRequest, appResponse *omaha.App) {
	logContext = logContext.WithFields(logrus.Fields{
		"machineId": appRequest.MachineID,
	})

	if appRequest.Id != coreOSAppID {
		appResponse.Status = "error-unknownApplication"
	} else {
		appResponse.Status = "ok"
	}

	// <UpdateCheck> tag
	if appRequest.UpdateCheck != nil {
		logContext.Debug("Handling UpdateCheck")
		ucResp := appResponse.AddUpdateCheck()

		appVersion, err := parseVersionString(appRequest.Version)
		if err != nil {
			logContext.Errorf("Could not parse client's version string: %v", err.Error())
			ucResp.Status = "error-invalidVersionString"
		} else {
			handleApiUpdateCheck(logContext, appVersion, appRequest.Track, appRequest.UpdateCheck, ucResp)
		}
	}

	// <ping> tag
	if appRequest.Ping != nil {
		// TODO register info from the ping

		// response is always "ok" according to the specs
		responsePing := appResponse.AddPing()
		responsePing.Status = "ok"
	}

	// <Event> tag
	for _, event := range appRequest.Events {
		logContext.Warnf("Event to handle: '%v', resultv '%v'", event.Type, event.Result)
		// TODO handle event
	}
}

// parse an 'UpdateCheck' tag of request and generate a corresponding 'UpdateCheck' tag of response
func handleApiUpdateCheck(logContext *logrus.Entry, appVersion payloadVersion, channel string, ucRequest, ucResp *omaha.UpdateCheck) {
	payload, err := db.GetNewerPayload(appVersion, channel)
	if err != nil {
		if err == sql.ErrNoRows {
			logContext.Infof("Client already up-to-date")
			ucRequest.Status = "noupdate"
		} else {
			logContext.Errorf("Failed checking for newer payload: %v", err.Error())
			ucRequest.Status = "error-internal"
		}
	} else {
		logContext.Infof("Found update to version '%v' (id %v)", "1.2.3.4.5.6", payload.Url)

		ucResp.Status = "ok"
		// TODO implement protocol configuration (maybe use info from request URL?)
		url := fmt.Sprintf("http://%v:%v/file?id=", opts.Hostname, opts.Port)
		ucResp.AddUrl(url)

		manifest := ucResp.AddManifest("1.0.2")
		manifest.AddPackage(payload.SHA1, payload.Url, strconv.FormatInt(payload.Size, 10), true)
		action := manifest.AddAction("postinstall")
		action.Sha256 = payload.SHA256
		action.DisablePayloadBackoff = true
	}
}
