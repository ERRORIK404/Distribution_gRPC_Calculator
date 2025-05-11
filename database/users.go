package database

import (
	models "github.com/ERRORIK404/Distribution_gRPC_Calculator/pkg/db_models"
)

func (db *DB) CreateUser(login, hashpassword string) error {
    _, err := db.DB.Exec(
        "INSERT INTO users (login, password_hash) VALUES (?, ?)",
        login, hashpassword,
    )
    return err
}

func (db *DB) GetUser(login string) (*models.User, error) {
    var user models.User
    err := db.DB.QueryRow(
        "SELECT id, login, password_hash FROM users WHERE login = ?",
        login,
    ).Scan(&user.ID, &user.Login, &user.PasswordHash)

    if err != nil {
        return nil, err
    }
    return &user, nil
}