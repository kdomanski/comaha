package main

import (
	"database/sql"
	//"fmt"
)

type userDB struct {
	db *sql.DB
}

func newUserDB(filename string) (*userDB, error) {
	database, err := sql.Open("sqlite3", filename)
	if err != nil {
		return nil, err
	}

	_, err = database.Exec("CREATE TABLE payloads(filename, size, sha1, sha256)")
	if err != nil {
		return nil, err
	}

	return &userDB{db: database}, nil
}

func (u *userDB) Close() error {
	return u.db.Close()
}
