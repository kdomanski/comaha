package main

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	flags "github.com/jessevdk/go-flags"
	"github.com/julienschmidt/httprouter"
	_ "github.com/mattn/go-sqlite3"
	"math/rand"
	"net/http"
	"os"
	"path"
	"time"
)

const coreOSAppID = "{e96281a6-d1af-4bde-9a0a-97b76e56dc57}"

var db userDB
var fileBE fileBackend

var opts struct {
	ListenAddr        string `short:"l" long:"listenaddr" default:"0.0.0.0" description:"address to listen on"`
	Port              int    `short:"P" long:"port" default:"8080" description:"port to listen on"`
	DisableTimestamps bool   `short:"t" long:"disabletimestamps" description:"disable timestamps in logs (useful when using journald)"`
	Debug             bool   `short:"d" long:"debug" description:"run in debug mode"`
}

func main() {
	_, err := flags.Parse(&opts)
	if err != nil {
		os.Exit(1)
	}

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
	db, err = newSqliteDB("users.sqlite")
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

	router := httprouter.New()

	router.GET("/file", fileHandler)
	router.POST("/update", updateHandler)
	//http.HandleFunc("/admin/add_group", addGroupHandler)
	router.POST("/admin/add_payload", addPayloadHandler)
	router.GET("/admin/delete_payload", deletePayloadHandler)
	router.GET("/admin/channel/:channel/force_downgrade", channelForceDowngradeGetHandler)
	router.POST("/admin/channel/:channel/force_downgrade", channelForceDowngradePostHandler)
	router.GET("/panel", panelHandler)
	//http.HandleFunc("/admin/add_user", addUserHandler)
	router.GET("/", homeHandler)

	listenString := fmt.Sprintf("%v:%v", opts.ListenAddr, opts.Port)
	http.ListenAndServe(listenString, router)
}
