package db

import (
	"FilmCollection/config"
	"github.com/jackc/pgx"
	"log/slog"
)

var Conn *pgx.Conn

func init() {
	var err error
	Conn, err = pgx.Connect(config.Conn)
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
