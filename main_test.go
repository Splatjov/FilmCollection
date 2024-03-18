package main

import (
	"FilmCollection/db"
	"FilmCollection/handlers"
	"FilmCollection/structs"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
)

func TestActor(t *testing.T) {
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.SetBasicAuth("compileboy", "1234")
	rr := httptest.NewRecorder()
	handlers.Wrap(handlers.AddActor)(rr, req)
	if rr.Code != http.StatusUnauthorized {
		t.Errorf("AddActor returned wrong status code: got %v want %v", rr.Code, http.StatusUnauthorized)
	}
	bodyString := `{
		"name":"Test",
		"gender": "idk",
		"birth_date": "01.01.2000"
	}`
	body := strings.NewReader(bodyString)
	req, err = http.NewRequest("GET", "/", body)
	if err != nil {
		t.Fatal(err)
	}
	req.SetBasicAuth("splatjov", "1234")
	rr = httptest.NewRecorder()
	handlers.Wrap(handlers.AddActor)(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("AddActor returned wrong status code: got %v want %v", rr.Code, http.StatusOK)
	}
	rows, err := db.Conn.Query("SELECT id FROM actors WHERE name = 'Test' AND gender = 'idk' AND birth_date = '2000-01-01'")
	if err != nil {
		t.Fatal(err)
	}
	if !rows.Next() {
		t.Errorf("AddActor failed to add actor to database")
	}
	rows.Close()
	req, err = http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.SetBasicAuth("splatjov", "1234")
	rr = httptest.NewRecorder()
	//handlers.Wrap(handlers.GetActors)(rr, req)
	handlers.GetActors(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("GetActors returned wrong status code: got %v want %v", rr.Code, http.StatusOK)
	}
	var actors []structs.Actor
	err = json.NewDecoder(rr.Body).Decode(&actors)
	if err != nil {
		t.Fatal(err)
	}
	if len(actors) == 0 {
		t.Errorf("GetActors failed to return actors")
	}
	id := actors[0].Id
	stringId := strconv.Itoa(id)
	req, err = http.NewRequest("GET", "/get_actor?id="+stringId, nil)
	if err != nil {
		t.Fatal(err)
	}
	req.SetBasicAuth("compileboy", "1234")
	rr = httptest.NewRecorder()
	handlers.Wrap(handlers.GetActor)(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("GetActor returned wrong status code: got %v want %v", rr.Code, http.StatusOK)
	}
	bodyString = `{
		"id":` + stringId + `,
		"name":"Test2",
		"gender": "idk",
		"birth_date": "01.01.2000"
	}`
	body = strings.NewReader(bodyString)
	req, err = http.NewRequest("GET", "/update_actor", body)
	if err != nil {
		t.Fatal(err)
	}
	rr = httptest.NewRecorder()
	req.SetBasicAuth("splatjov", "1234")
	handlers.Wrap(handlers.UpdateActor)(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("UpdateActor returned wrong status code: got %v want %v", rr.Code, http.StatusOK)
	}
	req, err = http.NewRequest("GET", "/get_actor?id="+stringId, nil)
	if err != nil {
		t.Fatal(err)
	}
	req.SetBasicAuth("compileboy", "1234")
	rr = httptest.NewRecorder()
	handlers.Wrap(handlers.GetActor)(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("GetActor returned wrong status code: got %v want %v", rr.Code, http.StatusOK)
	}
	var actor structs.Actor
	err = json.NewDecoder(rr.Body).Decode(&actor)
	if err != nil {
		t.Fatal(err)
	}
	if actor.Name != "Test2" {
		t.Errorf("UpdateActor failed to update actor")
	}
	req, err = http.NewRequest("GET", "/delete_actor?id="+stringId, nil)
	if err != nil {
		t.Fatal(err)
	}
	req.SetBasicAuth("splatjov", "1234")
	rr = httptest.NewRecorder()
	handlers.Wrap(handlers.DeleteActor)(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("DeleteActor returned wrong status code: got %v want %v", rr.Code, http.StatusOK)
	}
	req, err = http.NewRequest("GET", "/get_actor?id="+stringId, nil)
	if err != nil {
		t.Fatal(err)
	}
	req.SetBasicAuth("compileboy", "1234")
	rr = httptest.NewRecorder()
	handlers.Wrap(handlers.GetActor)(rr, req)
	if rr.Code != http.StatusInternalServerError {
		t.Errorf("GetActor returned wrong status code: got %v want %v", rr.Code, http.StatusInternalServerError)
	}
}

func TestFilm(t *testing.T) {
	bodyString := `{
		"name":"Test",
		"description": "idk",
		"rating": 10,
		"release_date": "01.01.2000"
	}`
	body := strings.NewReader(bodyString)
	req, err := http.NewRequest("GET", "/", body)
	if err != nil {
		t.Fatal(err)
	}
	req.SetBasicAuth("splatjov", "1234")
	rr := httptest.NewRecorder()
	handlers.Wrap(handlers.AddFilm)(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("AddFilm returned wrong status code: got %v want %v", rr.Code, http.StatusOK)
	}
	rows, err := db.Conn.Query("SELECT id FROM films WHERE name = 'Test' AND description = 'idk' AND release_date = '2000-01-01'")
	if err != nil {
		t.Fatal(err)
	}
	if !rows.Next() {
		t.Errorf("AddActor failed to add actor to database")
	}
	rows.Close()
	req, err = http.NewRequest("GET", "/get_films?keyword=tEST", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.SetBasicAuth("splatjov", "1234")
	rr = httptest.NewRecorder()
	handlers.Wrap(handlers.GetFilms)(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("GetFilms returned wrong status code: got %v want %v", rr.Code, http.StatusOK)
	}
	var films []structs.Film
	err = json.NewDecoder(rr.Body).Decode(&films)
	if err != nil {
		t.Fatal(err)
	}
	if len(films) == 0 {
		t.Errorf("GetFilms failed to return films")
	}
	id := films[0].Id
	stringId := strconv.Itoa(id)
	req, err = http.NewRequest("GET", "/get_film?id="+stringId, nil)
	if err != nil {
		t.Fatal(err)
	}
	req.SetBasicAuth("compileboy", "1234")
	rr = httptest.NewRecorder()
	handlers.Wrap(handlers.GetFilm)(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("GetFilm returned wrong status code: got %v want %v", rr.Code, http.StatusOK)
	}
	var film structs.Film
	err = json.NewDecoder(rr.Body).Decode(&film)
	if err != nil {
		t.Fatal(err)
	}
	if film.Name != "Test" {
		t.Errorf("GetFilm failed to return correct film")
	}
	bodyString = `{
		"id":` + stringId + `,
		"name":"Test2",
		"description": "idk",	
		"rating": 10,
		"release_date": "01.01.2000"
	}`
	body = strings.NewReader(bodyString)
	req, err = http.NewRequest("GET", "/update_film", body)
	if err != nil {
		t.Fatal(err)
	}
	rr = httptest.NewRecorder()
	req.SetBasicAuth("splatjov", "1234")
	handlers.Wrap(handlers.UpdateFilm)(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("UpdateFilm returned wrong status code: got %v want %v", rr.Code, http.StatusOK)
	}
	req, err = http.NewRequest("GET", "/get_film?id="+stringId, nil)
	if err != nil {
		t.Fatal(err)
	}
	req.SetBasicAuth(
		"compileboy", "1234")
	rr = httptest.NewRecorder()
	handlers.Wrap(handlers.GetFilm)(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("GetFilm returned wrong status code: got %v want %v", rr.Code, http.StatusOK)
	}
	err = json.NewDecoder(rr.Body).Decode(&film)
	if err != nil {
		t.Fatal(err)
	}
	if film.Name != "Test2" {
		t.Errorf("UpdateFilm failed to update film")
	}
	req, err = http.NewRequest("GET", "/delete_film?id="+stringId, nil)
	if err != nil {
		t.Fatal(err)
	}
	req.SetBasicAuth("splatjov", "1234")
	rr = httptest.NewRecorder()
	handlers.Wrap(handlers.DeleteFilm)(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("DeleteFilm returned wrong status code: got %v want %v", rr.Code, http.StatusOK)
	}
	req, err = http.NewRequest("GET", "/get_film?id="+stringId, nil)
	if err != nil {
		t.Fatal(err)
	}
	req.SetBasicAuth("compileboy", "1234")
	rr = httptest.NewRecorder()
	handlers.Wrap(handlers.GetFilm)(rr, req)
	if rr.Code != http.StatusInternalServerError {
		t.Errorf("GetFilm returned wrong status code: got %v want %v", rr.Code, http.StatusInternalServerError)
	}
}
