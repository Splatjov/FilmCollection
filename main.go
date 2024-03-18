package main

import (
	"FilmCollection/handlers"
	"log"
	"log/slog"
	"net/http"
	"os"
)

// @title FilmCollection API
// @version 1.0
// @description This is a simple API for a film collection
func main() {
	// check if the logs directory exists
	if _, err := os.Stat("logs"); os.IsNotExist(err) {
		err := os.Mkdir("logs", 0755)
		if err != nil {
			log.Fatal("Failed to create logs directory:", err)
		}
	}

	file, err := os.OpenFile("logs/app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal("Failed to open log file:", err)
	}
	log.SetOutput(file)

	mux := http.NewServeMux()

	handlers.InitHandlers(mux)

	hostPort := ":" + os.Getenv("SERVER_PORT")
	slog.Info("Server started at " + hostPort)
	panic(http.ListenAndServe(hostPort, mux))
}
