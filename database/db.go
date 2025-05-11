package database

import (
	"database/sql"
	"log"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

type DB struct {
	DB *sql.DB
}

func InitDB() (*DB, error) {
    dbPath := "./calculator.db"
    
    // Проверяем существование файла базы данных
    _, err := os.Stat(dbPath)
    dbExists := err == nil

    db, err := sql.Open("sqlite3", dbPath)
    if err != nil {
        return nil, err
    }

    // Применяем схему только если база данных не существовала
    if !dbExists {
        log.Println("Create db")
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
    } else {
        log.Println("Db already exists")
    }

    return &DB{DB: db}, nil
}