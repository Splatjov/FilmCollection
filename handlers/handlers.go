package handlers

import (
	"net/http"
)

func InitHandlers(mux *http.ServeMux) {
	mux.HandleFunc("POST /add_actor", Wrap(AddActor))
	mux.HandleFunc("POST /update_actor", Wrap(UpdateActor))
	mux.HandleFunc("GET /get_actor", Wrap(GetActor))
	mux.HandleFunc("GET /get_actors", Wrap(GetActors))
	mux.HandleFunc("POST /add_film", Wrap(AddFilm))
	mux.HandleFunc("POST /update_film", Wrap(UpdateFilm))
	mux.HandleFunc("GET /get_film", Wrap(GetFilm))
	mux.HandleFunc("GET /get_films", Wrap(GetFilms))
	mux.HandleFunc("POST /delete_actor", Wrap(DeleteActor))
	mux.HandleFunc("POST /delete_film", Wrap(DeleteFilm))
}
