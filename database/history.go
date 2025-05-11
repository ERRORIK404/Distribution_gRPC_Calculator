package database

import (
	"database/sql"

	models "github.com/ERRORIK404/Distribution_gRPC_Calculator/pkg/db_models"
)

func AddHistoryEntry(db *sql.DB, username string, expr, result string) error {
    _, err := db.Exec(
        "INSERT INTO history (user_login, expression, result) VALUES (?, ?, ?)",
        username, expr, result,
    )
    return err
}

func GetUserHistoryByLogin(db *sql.DB, login string) ([]models.HistoryEntry, error) {
    rows, err := db.Query(
        "SELECT id, expression, result, timestamp FROM history WHERE user_login = ?",
        login,
    )
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var history []models.HistoryEntry
    for rows.Next() {
        var entry models.HistoryEntry
        if err := rows.Scan(&entry.ID, &entry.Expression, &entry.Result); err != nil {
            return nil, err
        }
        history = append(history, entry)
    }
    return history, nil
}

func GetHistoryEntryByLogin(db *sql.DB, login string) (*models.HistoryEntry, error) {
    var entry models.HistoryEntry
    err := db.QueryRow(
        "SELECT id, user_login, expression, result FROM history WHERE user_login = ?",
        login,
    ).Scan(&entry.ID, &entry.UserLogin, &entry.Expression, &entry.Result)

    if err != nil {
        return nil, err
    }
    return &entry, nil
}