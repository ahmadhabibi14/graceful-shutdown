package configs

import (
	"database/sql"
	"fmt"
	"os"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

func ConnectPostgresSQL() *sqlx.DB {
	postgresDbName := os.Getenv("POSTGRES_DB")
	postgresUser := os.Getenv("POSTGRES_USER")
	postgresPassword := os.Getenv("POSTGRES_PASSWORD")
	postgresPort := os.Getenv("POSTGRES_PORT")

	postgresURL := fmt.Sprintf(
		"user=%v password=%v dbname=%v port=%v sslmode=disable",
		postgresUser, postgresPassword, postgresDbName, postgresPort,
	)

	driverName := "postgres"
	db, err := sql.Open(driverName, postgresURL)
	if err != nil {
		panic(err)
	}

	db.SetMaxIdleConns(10)
	db.SetMaxOpenConns(100)
	db.SetConnMaxIdleTime(5 * time.Minute)
	db.SetConnMaxLifetime(60 * time.Minute)

	sqlxDb := sqlx.MustConnect(driverName, postgresURL)

	return sqlxDb
}