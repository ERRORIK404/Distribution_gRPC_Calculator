package database

import (
    "database/sql"
    "os"
	
    _ "github.com/mattn/go-sqlite3"
)

func InitDB() (*sql.DB, error) {
    db, err := sql.Open("sqlite3", "./calculator.db")
    if err != nil {
        return nil, err
    }

    // Применяем схему
    if _, err := db.Exec(`PRAGMA foreign_keys = ON;`); err != nil {
        return nil, err
    }

    schema, err := os.ReadFile("database/schema.sql")
    if err != nil {
        return nil, err
    }

    if _, err := db.Exec(string(schema)); err != nil {
        return nil, err
    }

    return db, nil
}