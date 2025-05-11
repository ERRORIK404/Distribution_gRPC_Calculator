-- Пользователи
CREATE TABLE IF NOT EXISTS users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    login TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL
);

-- История операций
CREATE TABLE IF NOT EXISTS history (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_login TEXT NOT NULL,
    expression TEXT NOT NULL,
    result REAL NOT NULL,
    status TEXT NOT NULL,
    FOREIGN KEY (user_login) REFERENCES users (login)
);