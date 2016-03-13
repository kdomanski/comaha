package main

import (
	"database/sql"
	log "github.com/Sirupsen/logrus"
	"sync"
	"time"
)

// sqlite3 cannot use a single connection concurrently - thus the mutex
type sqliteDB struct {
	db    *sql.DB
	mutex sync.Mutex
}

func newSqliteDB(filename string) (*sqliteDB, error) {
	database, err := sql.Open("sqlite3", filename)
	if err != nil {
		return nil, err
	}

	err = initStructure(database)
	if err != nil {
		return nil, err
	}

	return &sqliteDB{db: database}, nil
}

func initStructure(database *sql.DB) error {
	_, err := database.Exec("CREATE TABLE IF NOT EXISTS payloads(id TEXT, size INTEGER, sha1 TEXT, sha256 TEXT, ver_build INTEGER, ver_branch INTEGER, ver_patch INTEGER, ver_timestamp INTEGER)")
	if err != nil {
		return err
	}

	_, err = database.Exec("CREATE TABLE IF NOT EXISTS channel_payload_rel(payload TEXT, channel TEXT)")
	if err != nil {
		return err
	}

	_, err = database.Exec("CREATE TABLE IF NOT EXISTS client(id TEXT, name TEXT)")
	if err != nil {
		return err
	}

	_, err = database.Exec("CREATE TABLE IF NOT EXISTS channel_client_rel(client TEXT, channel TEXT)")
	if err != nil {
		return err
	}

	_, err = database.Exec("CREATE TABLE IF NOT EXISTS events(client TEXT, type INTEGER, result INTEGER, timestamp INTEGER)")
	if err != nil {
		return err
	}

	_, err = database.Exec("CREATE TABLE IF NOT EXISTS channel_settings(channel TEXT, force_downgrade INTEGER DEFAULT 0)")
	if err != nil {
		return err
	}

	return nil
}

func (u *sqliteDB) Close() error {
	return u.db.Close()
}

func (u *sqliteDB) AttachPayloadToChannel(id, channel string) error {
	u.mutex.Lock()
	defer u.mutex.Unlock()

	q, err := u.db.Prepare(`INSERT INTO channel_payload_rel (payload,channel) SELECT ?, ?
	                        WHERE NOT EXISTS(SELECT 1 FROM channel_payload_rel WHERE payload=? AND channel=?);`)
	if err != nil {
		return err
	}

	_, err = q.Exec(id, channel, id, channel)
	if err != nil {
		return err
	}

	return nil
}

func (u *sqliteDB) AddPayload(id, sha1, sha256 string, size int64, version payloadVersion) error {
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

func (u *sqliteDB) DeletePayload(id string) error {
	u.mutex.Lock()
	defer u.mutex.Unlock()
	tx, err := u.db.Begin()
	if err != nil {
		return err
	}

	_, err = tx.Exec("DELETE from payloads WHERE id=?;", id)
	if err != nil {
		return err
	}

	_, err = tx.Exec("DELETE from channel_payload_rel WHERE payload=?;", id)
	if err != nil {
		return err
	}

	_, err = tx.Exec("DELETE from channel_settings WHERE channel=?;", id)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (u *sqliteDB) PayloadExists(id string) bool {
	row := u.db.QueryRow(`SELECT EXISTS(SELECT 1 FROM payloads WHERE id=?);`, id)
	var result int64
	row.Scan(&result)

	return result > 0
}

func (u *sqliteDB) GetNewerPayload(currentVersion payloadVersion, channel string) (*payload, error) {
	u.mutex.Lock()
	defer u.mutex.Unlock()

	q, err := u.db.Prepare(`SELECT id,size,sha1,sha256,ver_build,ver_branch,ver_patch,ver_timestamp,ifnull(force_downgrade,0) FROM payloads AS P
		JOIN channel_payload_rel AS R ON P.id=R.payload
		LEFT OUTER JOIN channel_settings AS S ON S.channel=R.channel
		WHERE R.channel=?
		ORDER BY ver_build DESC, ver_branch DESC, ver_patch DESC, ver_timestamp DESC LIMIT 1;`)
	if err != nil {
		return nil, err
	}

	result := q.QueryRow(channel)

	var p payload
	var latest payloadVersion
	var forceDowngrade int
	var latestTimestamp int64
	err = result.Scan(&p.ID, &p.Size, &p.SHA1, &p.SHA256, &latest.build, &latest.branch, &latest.patch, &latestTimestamp, &forceDowngrade)
	if err != nil {
		return nil, err
	}
	latest.timestamp = time.Unix(latestTimestamp, 0).UTC()

	if forceDowngrade == 0 {
		if latest.IsGreater(currentVersion) {
			return &p, nil
		}
		return nil, nil
	}

	// forceDowngrade != 0
	if latest.IsEqual(currentVersion) == false {
		return &p, nil
	}

	return nil, nil
}

func (u *sqliteDB) ListChannels() ([]string, error) {
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

func (u *sqliteDB) ListImages(channel string) ([]payload, error) {
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

func (u *sqliteDB) LogEvent(client string, evType, evResult int) error {
	u.mutex.Lock()
	defer u.mutex.Unlock()

	q, err := u.db.Prepare("INSERT INTO events (client,type,result,timestamp) VALUES (?, ?, ?, ?);")
	if err != nil {
		return err
	}

	_, err = q.Exec(client, evType, evResult, time.Now().UTC().Unix())

	return err
}

type Event struct {
	MachineID string
	Type      int
	Result    int
	Timestamp string
}

func (u *sqliteDB) GetEvents() ([]Event, error) {
	u.mutex.Lock()
	defer u.mutex.Unlock()

	q, err := u.db.Prepare("SELECT client,type,result,timestamp FROM events ORDER BY timestamp ASC;")
	if err != nil {
		return nil, err
	}

	result, err := q.Query()
	if err != nil {
		return nil, err
	}

	out := []Event{}

	for result.Next() {
		var ev Event

		var timestamp int64

		err = result.Scan(&ev.MachineID, &ev.Type, &ev.Result, &timestamp)
		if err != nil {
			return nil, err
		}

		ev.Timestamp = time.Unix(timestamp, 0).UTC().String()
		out = append(out, ev)
	}

	return out, nil
}

func (u *sqliteDB) SetChannelForceDowngrade(channel string, value bool) error {
	var intValue int

	if value {
		intValue = 1
	} else {
		intValue = 0
	}

	u.mutex.Lock()
	defer u.mutex.Unlock()

	result, err := u.db.Exec("UPDATE channel_settings SET force_downgrade=? WHERE channel=?", intValue, channel)
	if err != nil {
		return err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if affected == 0 {
		_, err = u.db.Exec("INSERT OR IGNORE INTO channel_settings (channel, force_downgrade) VALUES (?, ?);", channel, intValue)
		if err != nil {
			return err
		}
	}

	return nil
}

func (u *sqliteDB) GetChannelForceDowngrade(channel string) (bool, error) {
	row := u.db.QueryRow("SELECT force_downgrade FROM channel_settings WHERE channel=?;", channel)

	var intValue int
	err := row.Scan(&intValue)
	if err != nil {
		// unset, returning default
		if err == sql.ErrNoRows {
			return false, nil
		}

		return false, err
	}

	if intValue == 0 {
		return false, nil
	}

	return true, nil
}
