package main

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	flags "github.com/jessevdk/go-flags"
	_ "github.com/mattn/go-sqlite3"
	"math/rand"
	"net/http"
	"os"
	"path"
	"time"
)

const coreOSAppID = "{e96281a6-d1af-4bde-9a0a-97b76e56dc57}"

var db *userDB
var fileBE fileBackend

var opts struct {
	ListenAddr        string `short:"l" long:"listenaddr" default:"0.0.0.0" description:"address to listen on"`
	Port              int    `short:"P" long:"port" default:"8080" description:"port to listen on"`
	DisableTimestamps bool   `short:"t" long:"disabletimestamps" description:"disable timestamps in logs (useful when using journald)"`
	Debug             bool   `short:"d" long:"debug" description:"run in debug mode"`
}

func main() {
	var err error
	flags.Parse(&opts)

	log.SetFormatter(&log.TextFormatter{DisableTimestamp: opts.DisableTimestamps})
	if opts.Debug {
		log.SetLevel(log.DebugLevel)
	} else {
		log.SetLevel(log.InfoLevel)
	}
	log.Info("COmaha update server starting")

	// seed the RNG
	rand.Seed(time.Now().UnixNano())

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

	http.HandleFunc("/file", fileHandler)
	http.HandleFunc("/update", updateHandler)
	http.HandleFunc("/shutdown", shutdownHandler)
	//http.HandleFunc("/admin/add_group", addGroupHandler)
	http.HandleFunc("/admin/add_payload", addPayloadHandler)
	http.HandleFunc("/panel", panelHandler)
	//http.HandleFunc("/admin/add_user", addUserHandler)
	http.HandleFunc("/", homeHandler)

	listenString := fmt.Sprintf("%v:%v", opts.ListenAddr, opts.Port)
	http.ListenAndServe(listenString, nil)
}
