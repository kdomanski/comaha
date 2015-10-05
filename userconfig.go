package main

import (
	"database/sql"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"sync"
)

// sqlite3 cannot use a single connection concurrently - thus the mutex
type userDB struct {
	db    *sql.DB
	mutex sync.Mutex
}

func newUserDB(filename string) (*userDB, error) {
	database, err := sql.Open("sqlite3", filename)
	if err != nil {
		return nil, err
	}

	_, err = database.Exec("CREATE TABLE IF NOT EXISTS payloads(id TEXT, size INTEGER, sha1 TEXT, sha256 TEXT, ver_build INTEGER, ver_branch INTEGER, ver_patch INTEGER, ver_timestamp INTEGER)")
	if err != nil {
		return nil, err
	}

	_, err = database.Exec("CREATE TABLE IF NOT EXISTS channel_payload_rel(payload TEXT, channel TEXT)")
	if err != nil {
		return nil, err
	}

	_, err = database.Exec("CREATE TABLE IF NOT EXISTS client(id TEXT, name TEXT)")
	if err != nil {
		return nil, err
	}

	_, err = database.Exec("CREATE TABLE IF NOT EXISTS channel_client_rel(client TEXT, channel TEXT)")
	if err != nil {
		return nil, err
	}

	return &userDB{db: database}, nil
}

func (u *userDB) Close() error {
	return u.db.Close()
}

func (u *userDB) AddPayload(id, sha1, sha256 string, size int64, version payloadVersion) error {
	u.mutex.Lock()
	defer u.mutex.Unlock()
	q, err := u.db.Prepare("INSERT INTO payloads (id,size,sha1,sha256,ver_build,ver_branch,ver_patch,ver_timestamp) VALUES (?, ?, ?, ?, ?, ?, ?, ?);")
	if err != nil {
		return err
	}
	_, err = q.Exec(id, size, sha1, sha256, version.build, version.branch, version.patch, version.timestamp.Unix())
	if err != nil {
		return err
	}

	log.Debugf("DB: added payload '%v', size=%v, version=%v.%v.%v+%v, sha1=%v, sha256=%v,", id, size, version.build, version.branch, version.patch, version.timestamp.Unix(), sha1, sha256)

	return nil
}

func (u *userDB) GetNewerPayload(currentVersion payloadVersion) (p payload, err error) {
	u.mutex.Lock()
	defer u.mutex.Unlock()

	q, err := u.db.Prepare("SELECT id,size,sha1,sha256 FROM payloads WHERE (ver_build > ?) OR (ver_build = ? AND ver_branch > ?) OR (ver_build = ? AND ver_branch = ? AND ver_patch > ?) OR (ver_build = ? AND ver_branch = ? AND ver_patch = ? AND ver_timestamp > ?) ORDER BY ver_build, ver_branch, ver_patch, ver_timestamp LIMIT 1;")
	if err != nil {
		return
	}

	result := q.QueryRow(currentVersion.build, currentVersion.build, currentVersion.branch, currentVersion.build, currentVersion.branch, currentVersion.patch, currentVersion.build, currentVersion.branch, currentVersion.patch, currentVersion.timestamp.Unix())

	err = result.Scan(&p.Url, &p.Size, &p.SHA1, &p.SHA256)

	return
}

type imageListElement struct {
	Id      string
	Version string
	Sha1    string
	Sha256  string
	Size    int64
}

func (u *userDB) ListImages(channel string) []imageListElement {
	u.mutex.Lock()
	defer u.mutex.Unlock()

	// TODO channels
	result, err := u.db.Query("SELECT id,ver_build,ver_branch,ver_patch,ver_timestamp,sha1,sha256,size FROM payloads ORDER BY ver_build, ver_branch, ver_patch, ver_timestamp;")
	if err != nil {
		return []imageListElement{}
	}

	out := []imageListElement{}

	for result.Next() {
		var image imageListElement

		var verBuild int
		var verBranch int
		var verPatch int
		var verTimestamp int

		err = result.Scan(&image.Id, &verBuild, &verBranch, &verPatch, &verTimestamp, &image.Sha1, &image.Sha256, &image.Size)
		if err != nil {
			return []imageListElement{}
		}

		// TODO timestamp formatting
		image.Version = fmt.Sprintf("%v.%v.%v+%v", verBuild, verBranch, verPatch, verTimestamp)
		out = append(out, image)
	}

	return out
}
