package model

import (
	"database/sql"
	"fmt"
	"github.com/morluque/moenawark/mwkerr"
	"github.com/morluque/moenawark/password"
	"github.com/morluque/moenawark/sqlstore"
	"time"
)

// User represents a user of the game.
type User struct {
	ID         int64      `json:"id"`
	Character  *Character `json:"character,omitempty"`
	Login      string     `json:"login"`
	password   string     `json:""`
	Status     string     `json:"status"`
	GameMaster bool       `json:"game_master"`
}

/*
NewUser creates a new user with default values.
By default, a user is not yet registered and not a game master. The password
will be hashed before storing into the struct.
*/
func NewUser(login string, plaintextPassword string) *User {
	password := password.Encode(plaintextPassword)
	return &User{Login: login, password: password, Status: "new", GameMaster: false}
}

/*
HasCharacter returns true if the user has an associated character.

For example, game masters don't necessarily have associated characters.
*/
func (u *User) HasCharacter() bool {
	return u.Character != nil
}

// SetPassword sets a new (hashed) password for this user.
func (u *User) SetPassword(plaintext string) {
	u.password = password.Encode(plaintext)
}

func (u *User) getHashedPassword() string {
	return u.password
}

// CheckPassword verifies that a plaintext password matches the one stored as a
// hash.
func (u *User) CheckPassword(plaintextPassword string) error {
	return password.Check(u.password, plaintextPassword)
}

func (u *User) getCharacterID() sql.NullInt64 {
	if u.HasCharacter() {
		return sql.NullInt64{Int64: u.Character.ID, Valid: true}
	}
	return sql.NullInt64{Int64: 0, Valid: false}
}

func (u *User) create(db *sql.Tx) error {
	characterID := u.getCharacterID()
	if characterID.Valid && characterID.Int64 <= 0 {
		u.Character.Save(db)
	}
	now := time.Now().Unix()
	result, err := db.Exec(
		`INSERT INTO users (login, password, status, game_master, character_id, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6)`,
		u.Login,
		u.getHashedPassword(),
		u.Status,
		u.GameMaster,
		characterID,
		now)
	if err == nil {
		id, err := result.LastInsertId()
		if err != nil {
			return err
		}
		u.ID = id
	}
	return err
}

func (u *User) update(db *sql.Tx) error {
	characterID := u.getCharacterID()
	if characterID.Valid && characterID.Int64 <= 0 {
		u.Character.Save(db)
	}
	_, err := db.Exec(
		`UPDATE users
		    SET login = $1, password = $2, status = $3, game_master = $4, character_id = $5
		  WHERE id = $6`,
		u.Login,
		u.getHashedPassword(),
		u.Status,
		u.GameMaster,
		characterID,
		u.ID)
	return err
}

// Save stores a user in database or updates it if it was previously stored.
func (u *User) Save(db *sql.Tx) error {
	var err error

	if u.ID <= 0 {
		err = u.create(db)
	} else {
		err = u.update(db)
	}
	if err != nil {
		if sqlstore.IsConstraintError(err) {
			log.Errorf("Constraint error: %s", err.Error())
			return mwkerr.New(mwkerr.DuplicateModel, "Duplicate user with login %s", u.Login)
		}
		return err
	}
	return nil
}

// Delete will remove user from database if it's status is "new".
func (u *User) Delete(db *sql.Tx) error {
	if u.Status != "new" {
		log.Warnf("User %s was active, can't delete now.", u.Login)
		return fmt.Errorf("Can only delete inactive users, but %s is %s", u.Login, u.Status)
	}
	_, err := db.Exec("DELETE FROM users WHERE id = $1", u.ID)
	return err
}

// ListUsers loads a list of users from database with pagination.
func ListUsers(db *sql.Tx, first, count uint) ([]User, error) {
	users := make([]User, count)
	q := fmt.Sprintf(`
	    SELECT id, login, status, game_master, character_id
	      FROM users
	  ORDER BY id
	     LIMIT %d OFFSET %d`, count, first)
	rows, err := db.Query(q)
	if err != nil {
		return users, err
	}
	defer rows.Close()

	i := 0
	for rows.Next() {
		var id int64
		var login, status string
		var gameMaster bool
		var characterID sql.NullInt64
		err = rows.Scan(&id, &login, &status, &gameMaster, &characterID)
		if err != nil {
			return users, err
		}
		var c *Character
		if characterID.Valid {
			char, _ := LoadCharacterByID(db, characterID.Int64)
			c = char
		}
		users[i] = User{ID: id, Login: login, Status: status, GameMaster: gameMaster, Character: c}
		i++
	}
	err = rows.Err()
	if err != nil {
		return users, err
	}
	users = users[0:i]
	return users, nil
}

// LoadUser loads a user from database by its login.
func LoadUser(db *sql.Tx, login string) (*User, error) {
	var id int64
	var password, status string
	var gameMaster bool
	var characterID sql.NullInt64

	row := db.QueryRow("SELECT id, password, status, game_master, character_id FROM users WHERE login = $1", login)
	err := row.Scan(&id, &password, &status, &gameMaster, &characterID)
	if err != nil {
		return nil, err
	}
	var c *Character
	if characterID.Valid {
		char, _ := LoadCharacterByID(db, characterID.Int64)
		c = char
	}
	return &User{ID: id, Login: login, password: password, Status: status, GameMaster: gameMaster, Character: c}, nil
}

// AuthUser loads a user from database if the login/password match.
func AuthUser(db *sql.Tx, login string, plaintextPassword string) (*User, error) {
	authErr := mwkerr.New(mwkerr.AuthError, "Authentication error")
	u, err := LoadUser(db, login)
	if err != nil {
		return nil, authErr
	}

	err = password.Check(u.password, plaintextPassword)
	if err != nil {
		return nil, authErr
	}

	return u, nil
}

// HasAdmin returns true if at least one user in database is game master.
func HasAdmin(db *sql.Tx) bool {
	var adminCount int
	row := db.QueryRow("SELECT count(id) AS nbadmin FROM users WHERE game_master = 1")
	err := row.Scan(&adminCount)
	if err != nil {
		return false
	}

	return adminCount > 0
}
