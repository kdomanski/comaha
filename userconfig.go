package main

import (
	"database/sql"
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

	_, err = database.Exec("CREATE TABLE IF NOT EXISTS payloads(id TEXT, size INTEGER, sha1 TEXT, sha256 TEXT, version TEXT)")
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

func (u *userDB) AddPayload(id, sha1, sha256 string, size int64, version string) error {
	u.mutex.Lock()
	defer u.mutex.Unlock()
	q, err := u.db.Prepare("INSERT INTO payloads (id,size,sha1,sha256,version) VALUES (?, ?, ?, ?, ?);")
	if err != nil {
		return err
	}
	_, err = q.Exec(id, size, sha1, sha256, version)
	if err != nil {
		return err
	}

	log.Debugf("DB: added payload '%v', size=%v, version=%v, sha1=%v, sha256=%v,", id, size, version, sha1, sha256)

	return nil
}

func (u *userDB) ListImages(channel string) {

}
