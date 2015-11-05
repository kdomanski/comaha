package main

import (
	"database/sql"
	"github.com/Sirupsen/logrus"
	"github.com/coreos/go-omaha/omaha"
	"strconv"
)

const (
	eventTypeDownload = 13
	eventTypeArrive   = 14
	eventTypeApply    = 3
	eventTypeSuccess  = 800

	eventResultError = 0
	eventResultOK    = 1
	eventResultDone  = 2
)

// parse an 'app' tag of request and generate a corresponding 'app' tag of response
func handleApiApp(logContext *logrus.Entry, localUrl string, appRequest, appResponse *omaha.App) {
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
			handleApiUpdateCheck(logContext, localUrl, appVersion, appRequest.Track, appRequest.UpdateCheck, ucResp)
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
	handleApiEvents(logContext, appRequest.MachineID, appRequest.Events)
}

func handleApiEvents(logContext *logrus.Entry, client string, events []*omaha.Event) error {
	for _, event := range events {
		evType, err := strconv.Atoi(event.Type)
		if err != nil {
			return err
		}

		evResult, err := strconv.Atoi(event.Result)
		if err != nil {
			return err
		}

		err = db.LogEvent(client, evType, evResult)
		if err != nil {
			logContext.Error(err)
		}
		switch evType {
		case eventTypeDownload:
			logContext.Info("Client is downloading new version.")
		case eventTypeArrive:
			logContext.Info("Client finished download.")
		case eventTypeApply:
			switch evResult {
			case eventResultOK:
				logContext.Info("Client applied package.")
			case eventResultError:
				logContext.Info("Client errored during update.")
			case eventResultDone:
				logContext.Info("Client upgraded to current version.")
			}
		case eventTypeSuccess:
			logContext.Info("Install success. Update completion prevented by instance.")
		default:
			logContext.Warn("Unknown event type %v.", evType)
		}
	}
	return nil
}

// parse an 'UpdateCheck' tag of request and generate a corresponding 'UpdateCheck' tag of response
func handleApiUpdateCheck(logContext *logrus.Entry, localUrl string, appVersion payloadVersion, channel string, ucRequest, ucResp *omaha.UpdateCheck) {
	payload, err := db.GetNewerPayload(appVersion, channel)
	if err != nil {
		if err == sql.ErrNoRows {
			logContext.Infof("Client already up-to-date")
			ucResp.Status = "noupdate"
		} else {
			logContext.Errorf("Failed checking for newer payload: %v", err.Error())
			ucResp.Status = "error-internal"
		}
	} else {
		logContext.Infof("Found update to version '%v' (id %v)", payload.Version, payload.ID)

		ucResp.Status = "ok"
		ucResp.AddUrl(localUrl + "/file?id=")

		manifest := ucResp.AddManifest("1.0.2")
		manifest.AddPackage(payload.SHA1, payload.ID, strconv.FormatInt(payload.Size, 10), true)
		action := manifest.AddAction("postinstall")
		action.Sha256 = payload.SHA256
		action.DisablePayloadBackoff = true
	}
}
