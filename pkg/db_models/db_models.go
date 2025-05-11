package db_models

type User struct {
    ID           int
    Login     string
    PasswordHash string
}

type HistoryEntry struct {
    ID         int
    UserLogin  string
    Expression string
    Status     string
    Result     float64
}