package main

import (
	"encoding/json"
	"fmt"
	"github.com/jackc/pgx"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func AddActor(w http.ResponseWriter, r *http.Request) {
	err := checkForAutorization(r, true)
	if err != nil {
		http.Error(w, "authorization error", http.StatusUnauthorized)
		slog.Error("Authorization error: ", "error", err, "status", http.StatusUnauthorized)
		return
	}
	var actor Actor
	err = json.NewDecoder(r.Body).Decode(&actor)
	if err != nil {
		http.Error(w, "error reading request body", http.StatusBadRequest)
		slog.Error("Error reading request body: ", "error", err, "status", http.StatusBadRequest)
		return
	}
	if actor.Id != 0 {
		http.Error(w, "id field must be empty", http.StatusBadRequest)
		slog.Error("AddActor", "status", http.StatusBadRequest, "error", "id field must be empty")
		return
	}
	_, err = conn.Exec("INSERT INTO actors (name, gender, birth_date) VALUES ($1, $2, $3)", actor.Name, actor.Gender, actor.BirthDate.Format("2006-01-02"))
	if err != nil {
		http.Error(w, "error adding actor", http.StatusInternalServerError)
		slog.Error("Error adding actor: ", "error", err, "status", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	slog.Info("AddActor Actor added", "status", http.StatusOK)
}

func GetActor(w http.ResponseWriter, r *http.Request) {
	err := checkForAutorization(r, false)
	if err != nil {
		http.Error(w, "authorization error", http.StatusUnauthorized)
		slog.Error("Authorization error: ", "error", err, "status", http.StatusUnauthorized)
		return
	}
	idString := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idString)
	if err != nil {
		http.Error(w, "invalid id format", http.StatusBadRequest)
		slog.Error("ID format error: ", "error", err, "status", http.StatusBadRequest)
		return
	}
	actor, err := getActorByID(id)
	if err != nil {
		http.Error(w, "error reading actor", http.StatusInternalServerError)
		slog.Error("Error reading actor: ", "error", err, "status", http.StatusInternalServerError)
		return
	}
	err = actor.getActorsFilms(id)
	if err != nil {
		http.Error(w, "error reading actor's films", http.StatusInternalServerError)
		slog.Error("Error reading actor's films: ", "error", err, "status", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(actor)
	if err != nil {
		http.Error(w, "error writing response", http.StatusInternalServerError)
		slog.Error("Error writing response: ", "error", err, "status", http.StatusInternalServerError)
		return
	}
	slog.Info("GetActor Actor retrieved", "status", http.StatusOK)
}

func GetActors(w http.ResponseWriter, r *http.Request) {
	err := checkForAutorization(r, false)
	if err != nil {
		http.Error(w, "authorization error", http.StatusUnauthorized)
		slog.Error("Authorization error: ", "error", err, "status", http.StatusUnauthorized)
		return
	}
	limitString := r.URL.Query().Get("limit")
	keyword := r.URL.Query().Get("keyword")
	limit := 10
	if limitString != "" {
		limit, err = strconv.Atoi(limitString)
	}
	if err != nil {
		http.Error(w, "invalid limit format", http.StatusBadRequest)
		slog.Error("Invalid limit format: ", "error", err, "status", http.StatusBadRequest)
		return
	}
	sortString := strings.ToLower(r.URL.Query().Get("reverse"))
	if sortString == "" {
		sortString = "true"
	}
	if sortString != "true" && sortString != "false" {
		http.Error(w, "invalid reverse format", http.StatusBadRequest)
		slog.Error("Invalid reverse format: ", "error", err, "status", http.StatusBadRequest)
		return
	}
	sortParameter := r.URL.Query().Get("sort_parameter")
	if sortParameter == "" {
		sortParameter = "id"
	}
	if sortParameter != "id" && sortParameter != "name" && sortParameter != "gender" && sortParameter != "birth_date" {
		http.Error(w, "invalid sort_parameter format", http.StatusBadRequest)
		slog.Error("Invalid sort_parameter format: ", "error", err, "status", http.StatusBadRequest)
		return
	}
	var rows *pgx.Rows
	order := "ASC"
	if sortString == "true" {
		order = "DESC"
	}
	sqlQuery := fmt.Sprintf(`SELECT * FROM actors WHERE name ILIKE $1 ORDER BY %s %s LIMIT %d`, sortParameter, order, limit)
	rows, err = conn.Query(sqlQuery, "%"+keyword+"%")
	if err != nil {
		http.Error(w, "error reading actors", http.StatusInternalServerError)
		slog.Error("Error reading actors: ", "error", err, "status", http.StatusInternalServerError)
		return
	}
	var actors []Actor
	for rows.Next() {
		var actor Actor
		var birthDate time.Time
		err = rows.Scan(&actor.Id, &actor.Name, &actor.Gender, &birthDate)
		if err != nil {
			http.Error(w, "error reading actor", http.StatusInternalServerError)
			slog.Error("GetActors", err, "error reading actor")
			return
		}
		actor.BirthDate = Date{birthDate}
		actors = append(actors, actor)
	}
	rows.Close()
	for i := range actors {
		err = actors[i].getActorsFilms(actors[i].Id)
		if err != nil {
			http.Error(w, "error reading actor's films", http.StatusInternalServerError)
			slog.Error("GetActors", err, "error reading actor's films")
			return
		}
	}
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(actors)
	if err != nil {
		http.Error(w, "error writing response", http.StatusInternalServerError)
		slog.Error("Error writing response: ", "error", err, "status", http.StatusInternalServerError)
		return
	}
	slog.Info("GetActors Actors retrieved", "status", http.StatusOK)
}

func UpdateActor(w http.ResponseWriter, r *http.Request) {
	err := checkForAutorization(r, true)
	if err != nil {
		http.Error(w, "authorization error", http.StatusUnauthorized)
		slog.Error("Authorization error: ", "error", err, "status", http.StatusUnauthorized)
		return
	}
	var actor Actor
	err = json.NewDecoder(r.Body).Decode(&actor)
	if err != nil {
		http.Error(w, "error reading request body", http.StatusBadRequest)
		slog.Error("Error reading request body: ", "error", err, "status", http.StatusBadRequest)
		return
	}
	oldActor, err := getActorByID(actor.Id)
	if err != nil {
		http.Error(w, "error reading actor", http.StatusInternalServerError)
		slog.Error("Error reading actor: ", "error", err, "status", http.StatusInternalServerError)
		return
	}
	if actor.Id == 0 {
		http.Error(w, "actor id not specified", http.StatusBadRequest)
		slog.Error("UpdateActor", "status", http.StatusBadRequest, "error", "actor id not specified")
		return
	}
	if actor.Name == "" {
		actor.Name = oldActor.Name
	}
	if actor.Gender == "" {
		actor.Gender = oldActor.Gender
	}
	if actor.BirthDate.IsZero() {
		actor.BirthDate = oldActor.BirthDate
	}
	_, err = conn.Exec("UPDATE actors SET name = ($1), gender = ($2), birth_date = ($3) WHERE id = ($4)", actor.Name, actor.Gender, actor.BirthDate.Format("2006-01-02"), actor.Id)
	if err != nil {
		http.Error(w, "error updating actor", http.StatusInternalServerError)
		slog.Error("Error updating actor: ", "error", err, "status", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	slog.Info("UpdateActor Actor updated", "status", http.StatusOK)
}

func DeleteActor(w http.ResponseWriter, r *http.Request) {
	err := checkForAutorization(r, true)
	if err != nil {
		http.Error(w, "authorization error", http.StatusUnauthorized)
		slog.Error("Authorization error: ", "error", err, "status", http.StatusUnauthorized)
		return
	}
	idString := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idString)
	if err != nil {
		http.Error(w, "invalid id format", http.StatusBadRequest)
		slog.Error("ID format error: ", "error", err, "status", http.StatusBadRequest)
		return
	}
	_, err = conn.Exec("DELETE FROM actors WHERE id = $1", id)
	if err != nil {
		http.Error(w, "error deleting actor", http.StatusInternalServerError)
		slog.Error("Error deleting actor: ", "error", err, "status", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	slog.Info("DeleteActor Actor deleted", "status", http.StatusOK)
}

func AddFilm(w http.ResponseWriter, r *http.Request) {
	err := checkForAutorization(r, true)
	if err != nil {
		http.Error(w, "authorization error", http.StatusUnauthorized)
		slog.Error("Authorization error: ", "error", err, "status", http.StatusUnauthorized)
		return
	}
	var film Film
	err = json.NewDecoder(r.Body).Decode(&film)
	if err != nil {
		http.Error(w, "error reading request body", http.StatusBadRequest)
		slog.Error("Error reading request body: ", "error", err, "status", http.StatusBadRequest)
		return
	}

	err = conn.QueryRow("INSERT INTO films (name, description, release_date, rating) VALUES ($1, $2, $3, $4) RETURNING id", film.Name, film.Description, film.ReleaseDate.Format("2006-01-02"), film.Rating).Scan(&film.Id)
	if err != nil {
		http.Error(w, "error adding film", http.StatusInternalServerError)
		slog.Error("Error adding film: ", "error", err, "status", http.StatusInternalServerError)
		return
	}

	for _, actor := range film.Actors {
		_, err = conn.Exec("INSERT INTO moviecast (filmid, actorid) VALUES ($1, $2)", film.Id, actor)
		if err != nil {
			http.Error(w, "error adding actor to movie_cast", http.StatusInternalServerError)
			slog.Error("Error adding actor to movie_cast: ", "error", err, "status", http.StatusInternalServerError)
			return
		}
	}
	w.WriteHeader(http.StatusOK)
	slog.Info("AddFilm Film added", "status", http.StatusOK)
}

func GetFilm(w http.ResponseWriter, r *http.Request) {
	err := checkForAutorization(r, false)
	if err != nil {
		http.Error(w, "authorization error", http.StatusUnauthorized)
		slog.Error("Authorization error: ", "error", err, "status", http.StatusUnauthorized)
		return
	}
	idString := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idString)
	if err != nil {
		http.Error(w, "invalid id format", http.StatusBadRequest)
		slog.Error("ID format error: ", "error", err, "status", http.StatusBadRequest)
		return
	}
	film, err := getFilmByID(id)
	if err != nil {
		http.Error(w, "error reading film", http.StatusInternalServerError)
		slog.Error("Error reading film: ", "error", err, "status", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(film)
	if err != nil {
		http.Error(w, "error writing response", http.StatusInternalServerError)
		slog.Error("Error writing response: ", "error", err, "status", http.StatusInternalServerError)
		return
	}
	slog.Info("GetFilm Film retrieved", "status", http.StatusOK)
}

func GetFilms(w http.ResponseWriter, r *http.Request) {
	err := checkForAutorization(r, false)
	if err != nil {
		http.Error(w, "authorization error", http.StatusUnauthorized)
		slog.Error("Authorization error: ", "error", err, "status", http.StatusUnauthorized)
		return
	}
	limitString := r.URL.Query().Get("limit")
	keyword := r.URL.Query().Get("keyword")
	limit := 10
	if limitString != "" {
		limit, err = strconv.Atoi(limitString)
	}
	if err != nil {
		http.Error(w, "invalid limit format", http.StatusBadRequest)
		slog.Error("Invalid limit format: ", "error", err, "status", http.StatusBadRequest)
		return
	}
	sortString := strings.ToLower(r.URL.Query().Get("reverse"))
	if sortString == "" {
		sortString = "true"
	}
	if sortString != "true" && sortString != "false" {
		http.Error(w, "invalid reverse format", http.StatusBadRequest)
		slog.Error("Invalid reverse format: ", "error", err, "status", http.StatusBadRequest)
		return
	}
	sortParameter := r.URL.Query().Get("sort_parameter")
	if sortParameter == "" {
		sortParameter = "rating"
	}
	if sortParameter != "id" && sortParameter != "name" && sortParameter != "description" && sortParameter != "rating" && sortParameter != "release_date" {
		http.Error(w, "invalid sort_parameter format", http.StatusBadRequest)
		slog.Error("Invalid sort_parameter format: ", "error", err, "status", http.StatusBadRequest)
		return
	}
	var rows *pgx.Rows
	order := "ASC"
	if sortString == "true" {
		order = "DESC"
	}
	sqlQuery := fmt.Sprintf(`SELECT * FROM films WHERE name ILIKE $1 ORDER BY %s %s LIMIT %d`, sortParameter, order, limit)
	rows, err = conn.Query(sqlQuery, "%"+keyword+"%")
	if err != nil {
		http.Error(w, "error reading films", http.StatusInternalServerError)
		slog.Error("Error reading films: ", "error", err, "status", http.StatusInternalServerError)
		return
	}
	var films []Film
	for rows.Next() {
		var film Film
		var releaseDate time.Time
		err = rows.Scan(&film.Id, &film.Name, &film.Description, &film.Rating, &releaseDate)
		if err != nil {
			http.Error(w, "error reading film", http.StatusInternalServerError)
			slog.Error("Error reading film: ", "error", err, "status", http.StatusInternalServerError)
			return
		}
		film.ReleaseDate = Date{releaseDate}
		films = append(films, film)
	}
	rows.Close()
	for i := range films {
		err = films[i].getFilmsActors(films[i].Id)
		if err != nil {
			http.Error(w, "error reading film actors", http.StatusInternalServerError)
			slog.Error("Error reading film actors: ", "error", err, "status", http.StatusInternalServerError)
			return
		}
	}
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(films)
	if err != nil {
		http.Error(w, "error writing response", http.StatusInternalServerError)
		slog.Error("Error writing response: ", "error", err, "status", http.StatusInternalServerError)
		return
	}
	slog.Info("GetFilms Films retrieved", "status", http.StatusOK)
}

func UpdateFilm(w http.ResponseWriter, r *http.Request) {
	err := checkForAutorization(r, true)
	if err != nil {
		http.Error(w, "authorization error", http.StatusUnauthorized)
		slog.Error("Authorization error: ", "error", err, "status", http.StatusUnauthorized)
		return
	}
	var film Film
	err = json.NewDecoder(r.Body).Decode(&film)
	if err != nil {
		http.Error(w, "error reading request body", http.StatusBadRequest)
		slog.Error("Error reading request body: ", "error", err, "status", http.StatusBadRequest)
		return
	}
	if film.Id == 0 {
		http.Error(w, "film id not specified", http.StatusBadRequest)
		slog.Error("UpdateFilm", "status", http.StatusBadRequest, "error", "film id not specified")
		return
	}
	oldFilm, err := getFilmByID(film.Id)
	if err != nil {
		http.Error(w, "error reading film", http.StatusInternalServerError)
		slog.Error("Error reading film: ", "error", err, "status", http.StatusInternalServerError)
		return
	}
	if film.Name == "" {
		film.Name = oldFilm.Name
	}
	if film.Description == "" {
		film.Description = oldFilm.Description
	}
	if film.Rating == 0 { //think about it
		film.Rating = oldFilm.Rating
	}
	if film.ReleaseDate.IsZero() {
		film.ReleaseDate = oldFilm.ReleaseDate
	}
	err = conn.QueryRow("UPDATE films SET name = ($1), description = ($2), release_date = ($3), rating = ($4) WHERE id = ($5) RETURNING id", film.Name, film.Description, film.ReleaseDate.Format("2006-01-02"), film.Rating, film.Id).Scan(&film.Id)
	if err != nil {
		http.Error(w, "error adding film", http.StatusInternalServerError)
		slog.Error("Error adding film: ", "error", err, "status", http.StatusInternalServerError)
		return
	}
	if len(film.Actors) != 1 || film.Actors[0] == 0 {
		_, err = conn.Exec("DELETE FROM moviecast WHERE filmid = $1", film.Id)
	}
	for _, actor := range film.Actors {
		if actor == 0 {
			continue
		}
		_, err = conn.Exec("INSERT INTO moviecast (filmid, actorid) VALUES ($1, $2)", film.Id, actor)
		if err != nil {
			http.Error(w, "error adding actor to movie_cast", http.StatusInternalServerError)
			slog.Error("Error adding actor to movie_cast: ", "error", err, "status", http.StatusInternalServerError)
			return
		}
	}
	w.WriteHeader(http.StatusOK)
	slog.Info("UpdateFilm Film updated", "status", http.StatusOK)
}

func DeleteFilm(w http.ResponseWriter, r *http.Request) {
	err := checkForAutorization(r, true)
	if err != nil {
		http.Error(w, "authorization error", http.StatusUnauthorized)
		slog.Error("Authorization error: ", "error", err, "status", http.StatusUnauthorized)
		return
	}
	idString := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idString)
	if err != nil {
		http.Error(w, "invalid id format", http.StatusBadRequest)
		slog.Error("ID format error: ", "error", err, "status", http.StatusBadRequest)
		return
	}
	_, err = conn.Exec("DELETE FROM films WHERE id = $1", id)
	if err != nil {
		http.Error(w, "error deleting film", http.StatusInternalServerError)
		slog.Error("Error deleting film: ", "error", err, "status", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	slog.Info("DeleteFilm Film deleted", "status", http.StatusOK)
}

func initHandlers(mux *http.ServeMux) {
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
