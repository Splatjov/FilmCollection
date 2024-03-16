package main

import (
	"time"
)

type Date struct {
	time.Time
}

func (d *Date) UnmarshalJSON(data []byte) error {
	var err error
	data = data[1 : len(data)-1]
	d.Time, err = time.Parse("02.01.2006", string(data))
	return err
}

type Actor struct {
	Id        int    `json:"id"`
	Name      string `json:"name"`
	Gender    string `json:"gender"`
	BirthDate Date   `json:"birth_date"`
	Films     []int  `json:"films"`
}

type Film struct {
	Id          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Rating      int    `json:"rating"`
	ReleaseDate Date   `json:"release_date"`
	Actors      []int  `json:"actors"`
}

func (a *Actor) getActorsFilms(id int) error {
	rows, err := conn.Query("SELECT filmid FROM moviecast WHERE actorid = $1", id)
	if err != nil {
		return err
	}
	for rows.Next() {
		var filmID int
		err = rows.Scan(&filmID)
		if err != nil {
			return err
		}
		a.Films = append(a.Films, filmID)
	}
	return nil
}

func (a *Film) getFilmsActors(id int) error {
	rows, err := conn.Query("SELECT actorid FROM moviecast WHERE filmid = $1", id)
	if err != nil {
		return err
	}
	for rows.Next() {
		var actorID int
		err = rows.Scan(&actorID)
		if err != nil {
			return err
		}
		a.Actors = append(a.Actors, actorID)
	}
	return nil
}

func getActorByID(id int) (Actor, error) {
	q := conn.QueryRow("SELECT * FROM actors WHERE id = $1", id)
	var actor Actor
	var birthDate time.Time
	err := q.Scan(&actor.Id, &actor.Name, &actor.Gender, &birthDate)
	actor.BirthDate = Date{birthDate}
	if err != nil {
		return Actor{}, err
	}
	return actor, nil
}

func getFilmByID(id int) (Film, error) {
	q := conn.QueryRow("SELECT * FROM films WHERE id = $1", id)
	var film Film
	var releaseDate time.Time
	err := q.Scan(&film.Id, &film.Name, &film.Description, &film.Rating, &releaseDate)
	film.ReleaseDate = Date{releaseDate}
	if err != nil {
		return Film{}, err
	}
	err = film.getFilmsActors(id)
	if err != nil {
		return Film{}, err
	}
	return film, nil
}
