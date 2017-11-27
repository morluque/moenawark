package model

import (
	"database/sql"
	"github.com/morluque/moenawark/mwkerr"
	"github.com/morluque/moenawark/sqlstore"
)

// Place represents a place in the universe
// As a simplification, our universe only has two dimensions. That makes map
// drawing a lot easier :-) .
type Place struct {
	ID               int64  `json:"id"`
	Name             string `json:"name"`
	X                int    `json:"x"`
	Y                int    `json:"y"`
	EnergyProduction int    `json:"energy_production"`
}

// Wormhole links two places.
type Wormhole struct {
	ID          int64 `json:"id"`
	Source      Place `json:"source"`
	Destination Place `json:"destination"`
	Distance    int   `json:"distance"`
}

// NewPlace initializes a new place
func NewPlace(name string, x, y int) *Place {
	return &Place{Name: name, X: x, Y: y}
}

func (p *Place) create(db *sql.Tx) error {
	result, err := db.Exec(
		"INSERT INTO places (name, x, y, energy_production) VALUES ($1, $2, $3, $4)",
		p.Name,
		p.X,
		p.Y,
		p.EnergyProduction)
	if err == nil {
		id, err := result.LastInsertId()
		if err != nil {
			return err
		}
		p.ID = id
	}
	return err
}

func (p *Place) update(db *sql.Tx) error {
	_, err := db.Exec(
		"UPDATE places SET name = $1, x = $2, y = $3, energy_production = $4 WHERE id = $5",
		p.Name,
		p.X,
		p.Y,
		p.EnergyProduction,
		p.ID)
	return err
}

// Save stores a place in the database, either creating a row or updating an
// existing one.
// It uses the ID field to know wether it was created or not; ID from new
// places is always zero (Go zero value), and existing places have a positive
// non-null value, from SQLite auto-increment facility (which can't be zero).
func (p *Place) Save(db *sql.Tx) error {
	var err error

	if p.ID <= 0 {
		err = p.create(db)
	} else {
		err = p.update(db)
	}
	if err != nil {
		if sqlstore.IsConstraintError(err) {
			return mwkerr.New(mwkerr.DuplicateModel, "Duplicate place %s at (%d, %d)", p.Name, p.X, p.Y)
		}
		return err
	}
	return nil
}

// LoadPlace loads the place at (x, y) coordinate, if it exists.
func LoadPlace(db *sql.DB, x, y int) (*Place, error) {
	var id int64
	var energyProduction int
	var name string

	row := db.QueryRow("SELECT id, name, energy_production FROM places WHERE x = $1, y = $2", x, y)
	err := row.Scan(&id, &name, &energyProduction)
	if err != nil {
		return nil, err
	}
	return &Place{ID: id, X: x, Y: y, EnergyProduction: energyProduction}, nil
}

// NewWormhole initializes a new wormhole linking two places.
func NewWormhole(source, destination *Place, distance int) *Wormhole {
	return &Wormhole{
		Source:      *source,
		Destination: *destination,
		Distance:    distance,
	}
}

func (w *Wormhole) create(db *sql.Tx) error {
	result, err := db.Exec(
		"INSERT INTO wormholes (source_id, destination_id, distance) VALUES ($1, $2, $3)",
		w.Source.ID,
		w.Destination.ID,
		w.Distance)
	if err == nil {
		id, err := result.LastInsertId()
		if err != nil {
			return err
		}
		w.ID = id
	}
	return err
}

func (w *Wormhole) update(db *sql.Tx) error {
	_, err := db.Exec(
		"UPDATE wormholes SET source_id = $1, destination_id = $2, distance = $3 WHERE id = $5",
		w.Source.ID,
		w.Destination.ID,
		w.Distance,
		w.ID)
	return err
}

// Save stores a wormhole to database (create or update row).
func (w *Wormhole) Save(db *sql.Tx) error {
	var err error

	if w.ID <= 0 {
		err = w.create(db)
	} else {
		err = w.update(db)
	}
	if err != nil {
		if sqlstore.IsConstraintError(err) {
			return mwkerr.New(
				mwkerr.DuplicateModel, "Duplicate wormhole from (%d, %d) to (%d, %d)",
				w.Source.X, w.Source.Y,
				w.Destination.X, w.Destination.Y)
		}
		return err
	}
	return nil
}

// LoadWormholes loads wormholes that start at the given place.
func LoadWormholes(db *sql.DB, source *Place) ([]*Wormhole, error) {
	wormholes := make([]*Wormhole, 0)
	sql := `
     SELECT w.id AS w_id,
            w.distance AS w_distance,
				p.id AS p_id,
            p.name AS p_name,
            p.x AS p_x,
            p.y AS p_y,
            p.energy_production AS p_energy_production
       FROM wormholes w,
            places p
      WHERE p.id = $1
        AND w.source_id = p.id
	ORDER BY p_id, w.id`

	rows, err := db.Query(sql, source.ID)
	if err != nil {
		return wormholes, err
	}
	defer rows.Close()
	for rows.Next() {
		var placeID, wormholeID int64
		var distance, x, y, energyProduction int
		var placeName string
		if err := rows.Scan(&wormholeID, &distance, &placeName, &x, &y, &energyProduction); err != nil {
			return wormholes, err
		}
		dest := Place{ID: placeID, Name: placeName, X: x, Y: y, EnergyProduction: energyProduction}
		w := &Wormhole{Source: *source, Destination: dest, Distance: distance}
		wormholes = append(wormholes, w)
	}

	return wormholes, nil
}
