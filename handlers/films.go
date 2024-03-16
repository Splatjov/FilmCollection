package handlers

import (
	"FilmCollection/db"
	"FilmCollection/structs"
	"encoding/json"
	"fmt"
	"github.com/jackc/pgx"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func AddFilm(w http.ResponseWriter, r *http.Request) {
	err := db.CheckForAutorization(r, true)
	if err != nil {
		http.Error(w, "authorization error", http.StatusUnauthorized)
		slog.Error("Authorization error: ", "error", err, "status", http.StatusUnauthorized)
		return
	}
	var film structs.Film
	err = json.NewDecoder(r.Body).Decode(&film)
	if err != nil {
		http.Error(w, "error reading request body", http.StatusBadRequest)
		slog.Error("Error reading request body: ", "error", err, "status", http.StatusBadRequest)
		return
	}

	err = db.Conn.QueryRow("INSERT INTO films (name, description, release_date, rating) VALUES ($1, $2, $3, $4) RETURNING id", film.Name, film.Description, film.ReleaseDate.Format("2006-01-02"), film.Rating).Scan(&film.Id)
	if err != nil {
		http.Error(w, "error adding film", http.StatusInternalServerError)
		slog.Error("Error adding film: ", "error", err, "status", http.StatusInternalServerError)
		return
	}

	for _, actor := range film.Actors {
		_, err = db.Conn.Exec("INSERT INTO moviecast (filmid, actorid) VALUES ($1, $2)", film.Id, actor)
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
	err := db.CheckForAutorization(r, false)
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
	film, err := db.GetFilmByID(id)
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
	err := db.CheckForAutorization(r, false)
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
	rows, err = db.Conn.Query(sqlQuery, "%"+keyword+"%")
	if err != nil {
		http.Error(w, "error reading films", http.StatusInternalServerError)
		slog.Error("Error reading films: ", "error", err, "status", http.StatusInternalServerError)
		return
	}
	var films []structs.Film
	for rows.Next() {
		var film structs.Film
		var releaseDate time.Time
		err = rows.Scan(&film.Id, &film.Name, &film.Description, &film.Rating, &releaseDate)
		if err != nil {
			http.Error(w, "error reading film", http.StatusInternalServerError)
			slog.Error("Error reading film: ", "error", err, "status", http.StatusInternalServerError)
			return
		}
		film.ReleaseDate = structs.Date{releaseDate}
		films = append(films, film)
	}
	rows.Close()
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
	err := db.CheckForAutorization(r, true)
	if err != nil {
		http.Error(w, "authorization error", http.StatusUnauthorized)
		slog.Error("Authorization error: ", "error", err, "status", http.StatusUnauthorized)
		return
	}
	var film structs.Film
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
	oldFilm, err := db.GetFilmByID(film.Id)
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
	err = db.Conn.QueryRow("UPDATE films SET name = ($1), description = ($2), release_date = ($3), rating = ($4) WHERE id = ($5) RETURNING id", film.Name, film.Description, film.ReleaseDate.Format("2006-01-02"), film.Rating, film.Id).Scan(&film.Id)
	if err != nil {
		http.Error(w, "error adding film", http.StatusInternalServerError)
		slog.Error("Error adding film: ", "error", err, "status", http.StatusInternalServerError)
		return
	}
	if len(film.Actors) != 1 || film.Actors[0] == 0 {
		_, err = db.Conn.Exec("DELETE FROM moviecast WHERE filmid = $1", film.Id)
	}
	for _, actor := range film.Actors {
		if actor == 0 {
			continue
		}
		_, err = db.Conn.Exec("INSERT INTO moviecast (filmid, actorid) VALUES ($1, $2)", film.Id, actor)
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
	err := db.CheckForAutorization(r, true)
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
	_, err = db.Conn.Exec("DELETE FROM films WHERE id = $1", id)
	if err != nil {
		http.Error(w, "error deleting film", http.StatusInternalServerError)
		slog.Error("Error deleting film: ", "error", err, "status", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	slog.Info("DeleteFilm Film deleted", "status", http.StatusOK)
}
