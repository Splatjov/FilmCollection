package main

import (
	"fmt"
	"github.com/jackc/pgx"
	"log"
	"log/slog"
	"net/http"
	"os"
)

var conn *pgx.Conn

func main() {
	file, err := os.OpenFile("app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal("Failed to open log file:", err)
	}
	log.SetOutput(file)
	config, configServer, err := loadConfig("config.toml")
	if err != nil {
		slog.Error("Failed to load config: ", "error", err)
		return
	}
	conn, err = pgx.Connect(config)
	if err != nil {
		slog.Error("Failed to connect to the database: ", "error", err)
		return
	}
	defer func(conn *pgx.Conn) {
		err := conn.Close()
		if err != nil {
			slog.Error("Failed to close the database connection: ", "error", err)
		}
	}(conn)
	mux := http.NewServeMux()
	err = initTables()
	if err != nil {
		slog.Error("Failed to init tables: ", "error", err)
		return
	}
	fmt.Println("Все подключилось!")
	initHandlers(mux)
	err = http.ListenAndServe(fmt.Sprintf("%s:%d", configServer.Host, configServer.Port), mux)
}
