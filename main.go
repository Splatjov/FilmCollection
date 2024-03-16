package main

import (
	"FilmCollection/config"
	"FilmCollection/handlers"
	"fmt"
	"log"
	"net/http"
	"os"
)

func main() {
	file, err := os.OpenFile("app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal("Failed to open log file:", err)
	}
	log.SetOutput(file)

	mux := http.NewServeMux()

	fmt.Println("Все подключилось!")
	handlers.InitHandlers(mux)
	err = http.ListenAndServe(fmt.Sprintf("%s:%d", config.Server.Host, config.Server.Port), mux)
}
