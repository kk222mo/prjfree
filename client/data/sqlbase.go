package data

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

var DB *sql.DB

func LoadDB(path string) {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return
	}
	DB = db
	_, err = DB.Exec(`CREATE TABLE IF NOT EXISTS blocks (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    hash VARCHAR(256),
    date DATETIME,
    num INTEGER,
    topic VARCHAR(128),
    data TEXT, UNIQUE(hash, num)
    )`)
	if err != nil {
		return
	}
}
