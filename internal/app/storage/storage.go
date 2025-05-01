package storage

import (
	"fmt"

	"github.com/Leo-MathGuy/YandexLMS_Final/internal/app/logging"

	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

// MARK: DB
var db *sql.DB

func ConnectDB() {
	database, err := sql.Open("sqlite3", "./sqlite3.db")
	if err != nil {
		logging.Panic("DB Failed")
	}
	db = database
}

func query(q string) (sql.Result, error) {
	if db == nil {
		logging.Error("Database not connected")
		return nil, fmt.Errorf("database not connected")
	}
	if res, err := db.Exec(q); err != nil {
		return nil, err
	} else {
		return res, nil
	}
}

func createTables() {
	if _, err := query(`
	CREATE TABLE IF NOT EXISTS Users (
	id INT AUTO_INCREMENT PRIMARY KEY
	login VARCHAR(32) NOT NULL UNIQUE
	passHash BLOB(32) NOT NULL
	)
	`); err != nil {
		logging.Panic(err.Error())
	}
}
