package db_models

type User struct {
    ID           int
    Username     string
    PasswordHash string
}

type HistoryEntry struct {
    ID         int
    UserLogin  string
    Expression string
    Result     string
}