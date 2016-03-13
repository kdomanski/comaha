package main

import (
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"os"
	"path"
	"time"
)

type payloadVersion struct {
	build     int32
	branch    int32
	patch     int32
	timestamp time.Time
}

const timelayout = "2006-01-02-1504"

func (v payloadVersion) String() string {
	if v.timestamp.Unix() == 0 {
		return fmt.Sprintf("%v.%v.%v", v.build, v.branch, v.patch)
	} else {
		return fmt.Sprintf("%v.%v.%v+%v", v.build, v.branch, v.patch, v.timestamp.Format(timelayout))
	}
}

func replace(oldID, newID string, db *sql.DB) {
	tx, err := db.Begin()
	if err != nil {
		fmt.Fprintf(os.Stderr, "TX Begin: %v\n", err.Error())
		os.Exit(1)
	}

	_, err = tx.Exec("DELETE FROM payloads SET WHERE id=?", oldID)
	if err != nil {
		tx.Rollback()
		fmt.Fprintf(os.Stderr, "TX DELETE payloads: %v\n", err.Error())
		os.Exit(1)
	}

	_, err = tx.Exec("UPDATE channel_payload_rel SET payload=? WHERE payload=?", newID, oldID)
	if err != nil {
		tx.Rollback()
		fmt.Fprintf(os.Stderr, "TX UPDATE channel_payload_rel: %v\n", err.Error())
		os.Exit(1)
	}

	err = tx.Commit()
	if err != nil {
		tx.Rollback()
		fmt.Fprintf(os.Stderr, "TX COMMIT: %v\n", err.Error())
		os.Exit(1)
	}

	filePath := path.Join("storage", oldID)
	err = os.Remove(filePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "os.Remove(\"%v\"): %v\n", filePath, err.Error())
		os.Exit(1)
	}
}

func dedup(clones []string, db *sql.DB) {
	masterID := clones[0]
	dupes := clones[1:]

	for _, d := range dupes {
		replace(d, masterID, db)
	}
}

func main() {
	database, err := sql.Open("sqlite3", "users.sqlite")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer database.Close()

	rows, err := database.Query(`SELECT id,ver_build,ver_branch,ver_patch,ver_timestamp FROM payloads;`)
	if err != nil {
		fmt.Fprintf(os.Stderr, "SELECT: %v\n", err.Error())
		os.Exit(1)
	}

	allPayloads := make(map[string][]string)

	for rows.Next() {
		var id string
		var ver payloadVersion
		var timestamp int64
		err = rows.Scan(&id, &ver.build, &ver.branch, &ver.patch, &timestamp)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Scan: %v\n", err.Error())
			os.Exit(1)
		}
		ver.timestamp = time.Unix(timestamp, 0).UTC()
		vString := ver.String()
		allPayloads[vString] = append(allPayloads[vString], id)
	}

	for v, clones := range allPayloads {
		fmt.Printf("deduping %v clones of version '%v'\n", len(clones), v)
		dedup(clones, database)
	}

}
