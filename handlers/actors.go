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

func AddActor(w http.ResponseWriter, r *http.Request) {
	err := db.CheckForAutorization(r, true)
	if err != nil {
		http.Error(w, "authorization error", http.StatusUnauthorized)
		slog.Error("Authorization error: ", "error", err, "status", http.StatusUnauthorized)
		return
	}
	var actor structs.Actor
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
	_, err = db.Conn.Exec("INSERT INTO actors (name, gender, birth_date) VALUES ($1, $2, $3)", actor.Name, actor.Gender, actor.BirthDate.Format("2006-01-02"))
	if err != nil {
		http.Error(w, "error adding actor", http.StatusInternalServerError)
		slog.Error("Error adding actor: ", "error", err, "status", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	slog.Info("AddActor Actor added", "status", http.StatusOK)
}

func GetActor(w http.ResponseWriter, r *http.Request) {
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
	actor, err := db.GetActorByID(id)
	if err != nil {
		http.Error(w, "error reading actor", http.StatusInternalServerError)
		slog.Error("Error reading actor: ", "error", err, "status", http.StatusInternalServerError)
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
	sqlQuery := fmt.Sprintf(`SELECT id FROM actors WHERE name ILIKE $1 ORDER BY %s %s LIMIT %d`, sortParameter, order, limit)
	rows, err = db.Conn.Query(sqlQuery, "%"+keyword+"%")
	if err != nil {
		http.Error(w, "error reading actors", http.StatusInternalServerError)
		slog.Error("Error reading actors: ", "error", err, "status", http.StatusInternalServerError)
		return
	}
	var actors []structs.Actor
	for rows.Next() {
		var actor structs.Actor
		var birthDate time.Time
		err = rows.Scan(&actor.Id)
		actor, err = db.GetActorByID(actor.Id)
		if err != nil {
			http.Error(w, "error reading actor", http.StatusInternalServerError)
			slog.Error("GetActors", err, "error reading actor")
			return
		}
		actor.BirthDate = structs.Date{birthDate}
		actors = append(actors, actor)
	}
	rows.Close()
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
	err := db.CheckForAutorization(r, true)
	if err != nil {
		http.Error(w, "authorization error", http.StatusUnauthorized)
		slog.Error("Authorization error: ", "error", err, "status", http.StatusUnauthorized)
		return
	}
	var actor structs.Actor
	err = json.NewDecoder(r.Body).Decode(&actor)
	if err != nil {
		http.Error(w, "error reading request body", http.StatusBadRequest)
		slog.Error("Error reading request body: ", "error", err, "status", http.StatusBadRequest)
		return
	}
	oldActor, err := db.GetActorByID(actor.Id)
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
	_, err = db.Conn.Exec("UPDATE actors SET name = ($1), gender = ($2), birth_date = ($3) WHERE id = ($4)", actor.Name, actor.Gender, actor.BirthDate.Format("2006-01-02"), actor.Id)
	if err != nil {
		http.Error(w, "error updating actor", http.StatusInternalServerError)
		slog.Error("Error updating actor: ", "error", err, "status", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	slog.Info("UpdateActor Actor updated", "status", http.StatusOK)
}

func DeleteActor(w http.ResponseWriter, r *http.Request) {
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
	_, err = db.Conn.Exec("DELETE FROM actors WHERE id = $1", id)
	if err != nil {
		http.Error(w, "error deleting actor", http.StatusInternalServerError)
		slog.Error("Error deleting actor: ", "error", err, "status", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	slog.Info("DeleteActor Actor deleted", "status", http.StatusOK)
}
