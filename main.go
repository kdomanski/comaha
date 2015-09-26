package main

import (
	log "github.com/Sirupsen/logrus"
	_ "github.com/mattn/go-sqlite3"
	"math/rand"
	"net/http"
	"os"
	"path"
	"time"
)

const coreOSAppID = "{e96281a6-d1af-4bde-9a0a-97b76e56dc57}"

var backend *singleFileBackend
var db *userDB
var fileBE fileBackend

func main() {
	log.SetLevel(log.DebugLevel)
	log.Info("COmaha update server starting")

	// seed the RNG
	rand.Seed(time.Now().UnixNano())

	var err error

	// open db
	db, err = newUserDB("users.sqlite")
	if err != nil {
		log.Errorf("Could not open database: %v", err.Error())
		os.Exit(1)
	}
	defer db.Close()

	cwd, err := os.Getwd()
	if err != nil {
		log.Errorf("Could not cwd: %v", err.Error())
	}
	lfbe := newLocalFileBackend(path.Join(cwd, "storage"))
	fileBE = &lfbe

	backend, err = NewSingleFileBackend("payload-list.json")
	if err != nil {
		log.Errorf("Failed to load simple backend from 'payload-list.json': %v", err.Error())
		os.Exit(1)
	}

	http.HandleFunc("/file", fileHandler)
	http.HandleFunc("/update", updateHandler)
	http.HandleFunc("/shutdown", shutdownHandler)
	//http.HandleFunc("/admin/add_group", addGroupHandler)
	http.HandleFunc("/admin/add_payload", addPayloadHandler)
	//http.HandleFunc("/admin/add_user", addUserHandler)
	http.HandleFunc("/", homeHandler)
	http.ListenAndServe(":8080", nil)
}
