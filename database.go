package main

import (
	"database/sql"
	log "github.com/Sirupsen/logrus"
	"sync"
	"time"
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

	_, err = database.Exec("CREATE TABLE IF NOT EXISTS events(client TEXT, type INTEGER, result INTEGER, timestamp INTEGER)")
	if err != nil {
		return nil, err
	}

	return &userDB{db: database}, nil
}

func (u *userDB) Close() error {
	return u.db.Close()
}

func (u *userDB) AttachPayloadToChannel(id, channel string) error {
	u.mutex.Lock()
	defer u.mutex.Unlock()

	q, err := u.db.Prepare("INSERT INTO channel_payload_rel (payload,channel) VALUES (?, ?);")
	if err != nil {
		return err
	}

	_, err = q.Exec(id, channel)
	if err != nil {
		return err
	}

	return nil
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

func (u *userDB) GetNewerPayload(currentVersion payloadVersion, channel string) (p payload, err error) {
	u.mutex.Lock()
	defer u.mutex.Unlock()

	q, err := u.db.Prepare("SELECT id,size,sha1,sha256 FROM payloads AS P JOIN channel_payload_rel AS R ON P.id=R.payload WHERE R.channel=? AND ((ver_build > ?) OR (ver_build = ? AND ver_branch > ?) OR (ver_build = ? AND ver_branch = ? AND ver_patch > ?) OR (ver_build = ? AND ver_branch = ? AND ver_patch = ? AND ver_timestamp > ?)) ORDER BY ver_build DESC, ver_branch DESC, ver_patch DESC, ver_timestamp DESC LIMIT 1;")
	if err != nil {
		return
	}

	result := q.QueryRow(channel, currentVersion.build, currentVersion.build, currentVersion.branch, currentVersion.build, currentVersion.branch, currentVersion.patch, currentVersion.build, currentVersion.branch, currentVersion.patch, currentVersion.timestamp.Unix())

	err = result.Scan(&p.ID, &p.Size, &p.SHA1, &p.SHA256)

	return
}

func (u *userDB) ListChannels() ([]string, error) {
	u.mutex.Lock()
	defer u.mutex.Unlock()

	result, err := u.db.Query("SELECT DISTINCT channel FROM channel_payload_rel;")
	if err != nil {
		return nil, err
	}

	channels := []string{}

	for result.Next() {
		var chanName string
		err = result.Scan(&chanName)
		if err != nil {
			return nil, err
		}
		channels = append(channels, chanName)
	}

	return channels, nil
}

func (u *userDB) ListImages(channel string) ([]payload, error) {
	u.mutex.Lock()
	defer u.mutex.Unlock()

	q, err := u.db.Prepare("SELECT id,ver_build,ver_branch,ver_patch,ver_timestamp,sha1,sha256,size FROM payloads AS P JOIN channel_payload_rel AS R ON P.id=R.payload WHERE R.channel=? ORDER BY ver_build, ver_branch, ver_patch, ver_timestamp;")
	if err != nil {
		return nil, err
	}

	result, err := q.Query(channel)
	if err != nil {
		return nil, err
	}

	out := []payload{}

	for result.Next() {
		var image payload

		var ver payloadVersion
		var timestamp int64

		err = result.Scan(&image.ID, &ver.build, &ver.branch, &ver.patch, &timestamp, &image.SHA1, &image.SHA256, &image.Size)
		if err != nil {
			return nil, err
		}

		ver.timestamp = time.Unix(timestamp, 0).UTC()
		image.Version = ver.String()
		out = append(out, image)
	}

	return out, nil
}

func (u *userDB) LogEvent(client string, evType, evResult int) error {
	u.mutex.Lock()
	defer u.mutex.Unlock()

	q, err := u.db.Prepare("INSERT INTO events (client,type,result,timestamp) VALUES (?, ?, ?, ?);")
	if err != nil {
		return err
	}

	_, err = q.Exec(client, evType, evResult, time.Now().UTC().Unix())

	return err
}
