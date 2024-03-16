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
		http.Error(w, "ошибка авторизации", http.StatusUnauthorized)
		slog.Error("Ошибка авториазации: ", "error", err, "status", http.StatusUnauthorized)
		return
	}
	var actor Actor
	err = json.NewDecoder(r.Body).Decode(&actor)
	if err != nil {
		http.Error(w, "ошибка чтения тела запроса", http.StatusBadRequest)
		slog.Error("Ошибка чтения тела запроса: ", "error", err, "status", http.StatusBadRequest)
		return
	}

	_, err = conn.Exec("INSERT INTO actors (name, gender, birth_date) VALUES ($1, $2, $3)", actor.Name, actor.Gender, actor.BirthDate.Format("2006-01-02"))
	if err != nil {
		http.Error(w, "ошибка добавления актера", http.StatusInternalServerError)
		slog.Error("Ошибка добавления актера: ", "error", err, "status", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	slog.Info("AddActor Добавление актера", "status", http.StatusOK)
}

func GetActor(w http.ResponseWriter, r *http.Request) {
	err := checkForAutorization(r, false)
	if err != nil {
		http.Error(w, "ошибка авторизации", http.StatusUnauthorized)
		slog.Error("Ошибка авторизации: ", "error", err, "status", http.StatusUnauthorized)
		return
	}
	idString := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idString)
	if err != nil {
		http.Error(w, "неверный формат id", http.StatusBadRequest)
		slog.Error("Ошибка формата id: ", "error", err, "status", http.StatusBadRequest)
		return
	}
	actor, err := getActorByID(id)
	if err != nil {
		http.Error(w, "ошибка чтения актера", http.StatusInternalServerError)
		slog.Error("Ошибка чтения актера: ", "error", err, "status", http.StatusInternalServerError)
		return
	}
	err = actor.getActorsFilms(id)
	if err != nil {
		http.Error(w, "ошибка чтения фильмов актера", http.StatusInternalServerError)
		slog.Error("Ошибка чтения фильмов актера: ", "error", err, "status", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(actor)
	if err != nil {
		http.Error(w, "ошибка записи ответа", http.StatusInternalServerError)
		slog.Error("Ошибка записи ответа: ", "error", err, "status", http.StatusInternalServerError)
		return
	}
	slog.Info("GetActor Получение актера", "status", http.StatusOK)
}

func GetActors(w http.ResponseWriter, r *http.Request) {
	err := checkForAutorization(r, false)
	if err != nil {
		http.Error(w, "ошибка авторизации", http.StatusUnauthorized)
		slog.Error("Ошибка авторизации: ", "error", err, "status", http.StatusUnauthorized)
		return
	}
	limitString := r.URL.Query().Get("limit")
	keyword := r.URL.Query().Get("keyword")
	limit := 10
	if limitString != "" {
		limit, err = strconv.Atoi(limitString)
	}
	if err != nil {
		http.Error(w, "неверный формат limit", http.StatusBadRequest)
		slog.Error("Неверный формат limit: ", "error", err, "status", http.StatusBadRequest)
		return
	}
	sortString := strings.ToLower(r.URL.Query().Get("reverse"))
	if sortString == "" {
		sortString = "false"
	}
	if sortString != "true" && sortString != "false" {
		http.Error(w, "неверный формат reverse", http.StatusBadRequest)
		slog.Error("Неверный формат reverse: ", "error", err, "status", http.StatusBadRequest)
		return
	}
	sortParameter := r.URL.Query().Get("sort_parameter")
	if sortParameter == "" {
		sortParameter = "id"
	}
	if sortParameter != "id" && sortParameter != "name" && sortParameter != "gender" && sortParameter != "birth_date" {
		http.Error(w, "неверный формат sort_parameter", http.StatusBadRequest)
		slog.Error("Неверный формат sort_parameter: ", "error", err, "status", http.StatusBadRequest)
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
		http.Error(w, "ошибка чтения актеров", http.StatusInternalServerError)
		slog.Error("Ошибка чтения актеров: ", "error", err, "status", http.StatusInternalServerError)
		return
	}
	var actors []Actor
	for rows.Next() {
		var actor Actor
		var birthDate time.Time
		err = rows.Scan(&actor.Id, &actor.Name, &actor.Gender, &birthDate)
		if err != nil {
			http.Error(w, "ошибка чтения актера", http.StatusInternalServerError)
			slog.Error("GetActors", err, "ошибка чтения актера")
			return
		}
		actor.BirthDate = Date{birthDate}
		actors = append(actors, actor)
	}
	rows.Close()
	for i := range actors {
		err = actors[i].getActorsFilms(actors[i].Id)
		if err != nil {
			http.Error(w, "ошибка чтения фильмов актера", http.StatusInternalServerError)
			slog.Error("GetActors", err, "ошибка чтения фильмов актера")
			return
		}
	}
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(actors)
	if err != nil {
		http.Error(w, "ошибка записи ответа", http.StatusInternalServerError)
		slog.Error("Ошибка записи ответа: ", "error", err, "status", http.StatusInternalServerError)
		return
	}
	slog.Info("GetActors Получение актеров", "status", http.StatusOK)
}

func UpdateActor(w http.ResponseWriter, r *http.Request) {
	err := checkForAutorization(r, true)
	if err != nil {
		http.Error(w, "ошибка авторизации", http.StatusUnauthorized)
		slog.Error("Ошибка авторизации: ", "error", err, "status", http.StatusUnauthorized)
		return
	}
	var actor Actor
	err = json.NewDecoder(r.Body).Decode(&actor)
	if err != nil {
		http.Error(w, "ошибка чтения тела запроса", http.StatusBadRequest)
		slog.Error("Ошибка чтения тела запроса: ", "error", err, "status", http.StatusBadRequest)
		return
	}
	oldActor, err := getActorByID(actor.Id)
	if err != nil {
		http.Error(w, "ошибка чтения актера", http.StatusInternalServerError)
		slog.Error("Ошибка чтения актера: ", "error", err, "status", http.StatusInternalServerError)
		return
	}
	if actor.Id == 0 {
		http.Error(w, "не указан id актера", http.StatusBadRequest)
		slog.Error("UpdateActor", "status", http.StatusBadRequest, "error", "не указан id актера")
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
		http.Error(w, "ошибка обновления актера", http.StatusInternalServerError)
		slog.Error("Ошибка обновления актера: ", "error", err, "status", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	slog.Info("UpdateActor Обновление актера", "status", http.StatusOK)
}

func DeleteActor(w http.ResponseWriter, r *http.Request) {
	err := checkForAutorization(r, true)
	if err != nil {
		http.Error(w, "ошибка авторизации", http.StatusUnauthorized)
		slog.Error("Ошибка авторизации: ", "error", err, "status", http.StatusUnauthorized)
		return
	}
	idString := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idString)
	if err != nil {
		http.Error(w, "неверный формат id", http.StatusBadRequest)
		slog.Error("Ошибка формата id: ", "error", err, "status", http.StatusBadRequest)
		return
	}
	_, err = conn.Exec("DELETE FROM actors WHERE id = $1", id)
	if err != nil {
		http.Error(w, "ошибка удаления актера", http.StatusInternalServerError)
		slog.Error("Ошибка удаления актера: ", "error", err, "status", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	slog.Info("DeleteActor Удаление актера", "status", http.StatusOK)
}

func AddFilm(w http.ResponseWriter, r *http.Request) {
	err := checkForAutorization(r, true)
	if err != nil {
		http.Error(w, "ошибка авторизации", http.StatusUnauthorized)
		slog.Error("Ошибка авторизации: ", "error", err, "status", http.StatusUnauthorized)
		return
	}
	var film Film
	err = json.NewDecoder(r.Body).Decode(&film)
	if err != nil {
		http.Error(w, "ошибка чтения тела запроса", http.StatusBadRequest)
		slog.Error("Ошибка чтения тела запроса: ", "error", err, "status", http.StatusBadRequest)
		return
	}

	err = conn.QueryRow("INSERT INTO films (name, description, release_date, rating) VALUES ($1, $2, $3, $4) RETURNING id", film.Name, film.Description, film.ReleaseDate.Format("2006-01-02"), film.Rating).Scan(&film.Id)
	if err != nil {
		http.Error(w, "ошибка добавления фильма", http.StatusInternalServerError)
		slog.Error("Ошибка добавления фильма: ", "error", err, "status", http.StatusInternalServerError)
		return
	}

	for _, actor := range film.Actors {
		_, err = conn.Exec("INSERT INTO moviecast (filmid, actorid) VALUES ($1, $2)", film.Id, actor)
		if err != nil {
			http.Error(w, "ошибка добавления актера в movie_cast", http.StatusInternalServerError)
			slog.Error("Ошибка добавления актера в movie_cast: ", "error", err, "status", http.StatusInternalServerError)
			return
		}
	}
	w.WriteHeader(http.StatusOK)
	slog.Info("AddFilm Добавление фильма", "status", http.StatusOK)
}

func GetFilm(w http.ResponseWriter, r *http.Request) {
	err := checkForAutorization(r, false)
	if err != nil {
		http.Error(w, "ошибка авторизации", http.StatusUnauthorized)
		slog.Error("Ошибка авторизации: ", "error", err, "status", http.StatusUnauthorized)
		return
	}
	idString := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idString)
	if err != nil {
		http.Error(w, "неверный формат id", http.StatusBadRequest)
		slog.Error("Ошибка формата id: ", "error", err, "status", http.StatusBadRequest)
		return
	}
	film, err := getFilmByID(id)
	if err != nil {
		http.Error(w, "ошибка чтения фильма", http.StatusInternalServerError)
		slog.Error("Ошибка чтения фильма: ", "error", err, "status", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(film)
	if err != nil {
		http.Error(w, "ошибка записи ответа", http.StatusInternalServerError)
		slog.Error("Ошибка записи ответа: ", "error", err, "status", http.StatusInternalServerError)
		return
	}
	slog.Info("GetFilm Получение фильма", "status", http.StatusOK)
}

func GetFilms(w http.ResponseWriter, r *http.Request) {
	err := checkForAutorization(r, false)
	if err != nil {
		http.Error(w, "ошибка авторизации", http.StatusUnauthorized)
		slog.Error("Ошибка авторизации: ", "error", err, "status", http.StatusUnauthorized)
		return
	}
	limitString := r.URL.Query().Get("limit")
	keyword := r.URL.Query().Get("keyword")
	limit := 10
	if limitString != "" {
		limit, err = strconv.Atoi(limitString)
	}
	if err != nil {
		http.Error(w, "неверный формат limit", http.StatusBadRequest)
		slog.Error("Неверный формат limit: ", "error", err, "status", http.StatusBadRequest)
		return
	}
	sortString := strings.ToLower(r.URL.Query().Get("reverse"))
	if sortString == "" {
		sortString = "true"
	}
	if sortString != "true" && sortString != "false" {
		http.Error(w, "неверный формат reverse", http.StatusBadRequest)
		slog.Error("Неверный формат reverse: ", "error", err, "status", http.StatusBadRequest)
		return
	}
	sortParameter := r.URL.Query().Get("sort_parameter")
	if sortParameter == "" {
		sortParameter = "rating"
	}
	if sortParameter != "id" && sortParameter != "name" && sortParameter != "description" && sortParameter != "rating" && sortParameter != "release_date" {
		http.Error(w, "неверный формат sort_parameter", http.StatusBadRequest)
		slog.Error("Неверный формат sort_parameter: ", "error", err, "status", http.StatusBadRequest)
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
		http.Error(w, "ошибка чтения фильмов", http.StatusInternalServerError)
		slog.Error("Ошибка чтения фильмов: ", "error", err, "status", http.StatusInternalServerError)
		return
	}
	var films []Film
	for rows.Next() {
		var film Film
		var releaseDate time.Time
		err = rows.Scan(&film.Id, &film.Name, &film.Description, &film.Rating, &releaseDate)
		if err != nil {
			http.Error(w, "ошибка чтения фильма", http.StatusInternalServerError)
			slog.Error("Ошибка чтения фильма: ", "error", err, "status", http.StatusInternalServerError)
			return
		}
		film.ReleaseDate = Date{releaseDate}
		films = append(films, film)
	}
	rows.Close()
	for i := range films {
		err = films[i].getFilmsActors(films[i].Id)
		if err != nil {
			http.Error(w, "ошибка чтения актеров фильма", http.StatusInternalServerError)
			slog.Error("Ошибка чтения актеров фильма: ", "error", err, "status", http.StatusInternalServerError)
			return
		}
	}
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(films)
	if err != nil {
		http.Error(w, "ошибка записи ответа", http.StatusInternalServerError)
		slog.Error("Ошибка записи ответа: ", "error", err, "status", http.StatusInternalServerError)
		return
	}
	slog.Info("GetFilms Получение фильмов", "status", http.StatusOK)
}

func UpdateFilm(w http.ResponseWriter, r *http.Request) {
	err := checkForAutorization(r, true)
	if err != nil {
		http.Error(w, "ошибка авторизации", http.StatusUnauthorized)
		slog.Error("Ошибка авторизации: ", "error", err, "status", http.StatusUnauthorized)
		return
	}
	var film Film
	err = json.NewDecoder(r.Body).Decode(&film)
	if err != nil {
		http.Error(w, "ошибка чтения тела запроса", http.StatusBadRequest)
		slog.Error("Ошибка чтения тела запроса: ", "error", err, "status", http.StatusBadRequest)
		return
	}
	if film.Id == 0 {
		http.Error(w, "не указан id фильма", http.StatusBadRequest)
		slog.Error("UpdateFilm", "status", http.StatusBadRequest, "error", "не указан id фильма")
		return
	}
	oldFilm, err := getFilmByID(film.Id)
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
		http.Error(w, "ошибка добавления фильма", http.StatusInternalServerError)
		slog.Error("Ошибка добавления фильма: ", "error", err, "status", http.StatusInternalServerError)
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
			http.Error(w, "ошибка добавления актера в movie_cast", http.StatusInternalServerError)
			slog.Error("Ошибка добавления актера в movie_cast: ", "error", err, "status", http.StatusInternalServerError)
			return
		}
	}
	w.WriteHeader(http.StatusOK)
	slog.Info("UpdateFilm Обновление фильма", "status", http.StatusOK)
}

func DeleteFilm(w http.ResponseWriter, r *http.Request) {
	err := checkForAutorization(r, true)
	if err != nil {
		http.Error(w, "ошибка авторизации", http.StatusUnauthorized)
		slog.Error("Ошибка авторизации: ", "error", err, "status", http.StatusUnauthorized)
		return
	}
	idString := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idString)
	if err != nil {
		http.Error(w, "неверный формат id", http.StatusBadRequest)
		slog.Error("Ошибка формата id: ", "error", err, "status", http.StatusBadRequest)
		return
	}
	_, err = conn.Exec("DELETE FROM films WHERE id = $1", id)
	if err != nil {
		http.Error(w, "ошибка удаления фильма", http.StatusInternalServerError)
		slog.Error("Ошибка удаления фильма: ", "error", err, "status", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	slog.Info("DeleteFilm Удаление фильма", "status", http.StatusOK)
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
