package handlers

import (
	"net/http"
)

func InitHandlers(mux *http.ServeMux) {
	mux.HandleFunc("POST /add_actor", AddActor)
	mux.HandleFunc("POST /update_actor", UpdateActor)
	mux.HandleFunc("GET /get_actor", GetActor)
	mux.HandleFunc("GET /get_actors", GetActors)
	mux.HandleFunc("POST /add_film", AddFilm)
	mux.HandleFunc("POST /update_film", UpdateFilm)
	mux.HandleFunc("GET /get_film", GetFilm)
	mux.HandleFunc("GET /get_films", GetFilms)
	mux.HandleFunc("POST /delete_actor", DeleteActor)
	mux.HandleFunc("POST /delete_film", DeleteFilm)
}
