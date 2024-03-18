package db

import (
	"github.com/jackc/pgx"
	"log/slog"
	"os"
	"strconv"
)

var Conn *pgx.Conn

func init() {
	var err error
	port, err := strconv.Atoi(os.Getenv("POSTGRES_INSIDE_PORT"))
	if err != nil {
		panic("Failed to get port: " + err.Error())
	}

	config := pgx.ConnConfig{
		User:     os.Getenv("POSTGRES_USER"),
		Password: os.Getenv("POSTGRES_PASSWORD"),
		Database: os.Getenv("POSTGRES_DB"),
		Port:     uint16(port),
		Host:     os.Getenv("POSTGRES_HOST"),
	}

	Conn, err = pgx.Connect(config)
	if err != nil {
		slog.Error("Failed to connect to the database: ", "error", err)
		return
	}

	err = initTables()
	if err != nil {
		slog.Error("Failed to init tables: ", "error", err)
		return
	}
}
