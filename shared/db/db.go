package db

import (
	"database/sql"
	"os"
)

// Connect opens a PostgreSQL connection using the DATABASE_URL environment variable.
func Connect() (*sql.DB, error) {
	return sql.Open("postgres", os.Getenv("DATABASE_URL"))
}
