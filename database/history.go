package database

import (
	models "github.com/ERRORIK404/Distribution_gRPC_Calculator/pkg/db_models"
)

func (h *DB) AddHistoryEntry(id int32, username string,  result float64, expr, status string) error {
    _, err := h.DB.Exec(
        "INSERT INTO history (id, user_login, expression, result, status) VALUES (?, ?, ?, ?, ?)",
        id, username, expr, result, status,
    )
    return err
}

func (h *DB) UpdateHistoryEntry(id int32, username string, expr string, result float64, status string) error {
    _, err := h.DB.Exec(
        "UPDATE history SET user_login = ?, expression = ?, result = ?, status = ? WHERE id = ?",
        username, expr, result, status, id,
    )
    return err
}

func (h *DB) GetUserHistoryByLogin(login string) ([]models.HistoryEntry, error) {
    rows, err := h.DB.Query(
        "SELECT id, expression, result, status FROM history WHERE user_login = ?",
        login,
    )
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var history []models.HistoryEntry
    for rows.Next() {
        var entry models.HistoryEntry
        if err := rows.Scan(&entry.ID, &entry.Expression, &entry.Result, &entry.Status); err != nil {
            return nil, err
        }
        history = append(history, entry)
    }
    return history, nil
}

func (h *DB) GetHistoryEntryByID(id int32, login string) (*models.HistoryEntry, error) {
    var entry models.HistoryEntry
    err := h.DB.QueryRow(
        "SELECT id, expression, result, status FROM history WHERE id = ? AND user_login = ?",
        id, login,
    ).Scan(&entry.ID, &entry.Expression, &entry.Result, &entry.Status)

    if err != nil {
        return nil, err
    }
    return &entry, nil
}