package database

import (
	"database/sql"

	models "github.com/ERRORIK404/Distribution_gRPC_Calculator/pkg/db_models"
)

func CreateUser(db *sql.DB, username, hashpassword string) error {
    _, err := db.Exec(
        "INSERT INTO users (username, password_hash) VALUES (?, ?)",
        username, hashpassword,
    )
    return err
}

func GetUser(db *sql.DB, username string) (*models.User, error) {
    var user models.User
    err := db.QueryRow(
        "SELECT id, username, password_hash FROM users WHERE username = ?",
        username,
    ).Scan(&user.ID, &user.Username, &user.PasswordHash)

    if err != nil {
        return nil, err
    }
    return &user, nil
}