package model

import (
	"database/sql"
	sqlite3 "github.com/mattn/go-sqlite3"
	"github.com/morluque/moenawark/mwkerr"
)

// Character is an in-game character, controlled by a player.
type Character struct {
	ID      int64  `json:"id"`
	Name    string `json:"name"`
	Power   uint   `json:"power"`
	Actions uint   `json:"actions"`
}

// NewCharacter creates a new character.
func NewCharacter(name string, power uint, actions uint) *Character {
	return &Character{Name: name, Power: power, Actions: actions}
}

func (c *Character) create(db *sql.Tx) error {
	result, err := db.Exec(
		"INSERT INTO characters (name, power, actions) VALUES ($1, $2, $3)",
		c.Name,
		c.Power,
		c.Actions)
	if err == nil {
		charID, err := result.LastInsertId()
		if err != nil {
			return err
		}
		c.ID = charID
	}
	return err
}

func (c *Character) update(db *sql.Tx) error {
	_, err := db.Exec(
		"UPDATE characters SET name = $1, power = $2, actions = $3",
		c.Name,
		c.Power,
		c.Actions)
	return err
}

// Save stores the character in database.
func (c *Character) Save(db *sql.Tx) error {
	var err error

	if c.ID <= 0 {
		err = c.create(db)
	} else {
		err = c.update(db)
	}
	if err != nil {
		if e, ok := err.(sqlite3.Error); ok {
			switch e.Code {
			case sqlite3.ErrConstraint:
				return mwkerr.New(mwkerr.DuplicateCharacter, "Duplicate character name %s", c.Name)
			default:
				return err
			}
		} else {
			return err
		}
	}
	return nil
}

// LoadCharacter fetches a character from databse by it's name.
func LoadCharacter(db *sql.Tx, name string) (*Character, error) {
	var id int64
	var power, actions uint
	row := db.QueryRow("SELECT id, power, actions FROM characters WHERE name = $1", name)
	err := row.Scan(&id, &power, &actions)
	if err != nil {
		return nil, err
	}
	return &Character{ID: id, Name: name, Power: power, Actions: actions}, nil
}

// LoadCharacterByID fetches a character from database by it's ID.
func LoadCharacterByID(db *sql.Tx, id int64) (*Character, error) {
	var name string
	var power, actions uint
	row := db.QueryRow("SELECT name, power, actions FROM characters WHERE id = $1", id)
	err := row.Scan(&name, &power, &actions)
	if err != nil {
		return nil, err
	}
	return &Character{ID: id, Name: name, Power: power, Actions: actions}, nil
}
