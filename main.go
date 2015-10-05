package main

import (
	"flag"
	"fmt"
	log "github.com/Sirupsen/logrus"
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

var myHostname string      // local hostname advertised in responses
var listenAddr string      // address to listen on
var listenPort int         // port to listen on
var disableTimestamps bool // disable timestamps in logs

func main() {
	var err error
	flag.StringVar(&myHostname, "hostname", "", "hostname advertised when using local file backend")
	flag.StringVar(&listenAddr, "listenaddr", "0.0.0.0", "address to listen on")
	flag.IntVar(&listenPort, "port", 8080, "port to listen on")
	flag.BoolVar(&disableTimestamps, "disabletimestamps", false, "disable timestamps in logs")
	flag.Parse()
	if myHostname == "" {
		log.Error("You must set the 'hostname' parameter when using local file backend.")
		os.Exit(1)
	}

	log.SetFormatter(&log.TextFormatter{DisableTimestamp: disableTimestamps})
	log.SetLevel(log.DebugLevel)
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

	listenString := fmt.Sprintf("%v:%v", listenAddr, listenPort)
	http.ListenAndServe(listenString, nil)
}
