package structs

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
	if err != nil {
		d.Time, err = time.Parse("2006-01-02T00:00:00Z", string(data))
		if err != nil {
			return err
		}
		err = nil
	}
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
