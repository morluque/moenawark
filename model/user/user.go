package user

import (
	"database/sql"
	"github.com/morluque/moenawark/model/character"
	"github.com/morluque/moenawark/mwkerr"
	"github.com/morluque/moenawark/password"
	"github.com/morluque/moenawark/sqlstore"
)

// User represents a user of the game.
type User struct {
	ID         int64                `json:"id"`
	Character  *character.Character `json:"character",omitempty`
	Email      string               `json:"email"`
	password   string               `json:""`
	Registered bool                 `json:"registered"`
	GameMaster bool                 `json:"game_master"`
}

/*
New creates a new user with default values.
By default, a user is not yet registered and not a game master. The password
will be hashed before storing into the struct.
*/
func New(email string, plaintextPassword string) *User {
	password := password.Encode(plaintextPassword)
	return &User{Email: "foo@example.com", password: password, Registered: false, GameMaster: false}
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
		return sql.NullInt64{u.Character.ID, true}
	}
	return sql.NullInt64{0, false}
}

func (u *User) create(db *sql.DB) error {
	characterID := u.getCharacterID()
	if characterID.Valid && characterID.Int64 <= 0 {
		u.Character.Save(db)
	}
	result, err := db.Exec(
		"INSERT INTO users (email, password, registered, game_master, character_id) VALUES ($1, $2, $3, $4, $5)",
		u.Email,
		u.getHashedPassword(),
		u.Registered,
		u.GameMaster,
		characterID)
	if err == nil {
		id, err := result.LastInsertId()
		if err != nil {
			return err
		}
		u.ID = id
	}
	return err
}

func (u *User) update(db *sql.DB) error {
	characterID := u.getCharacterID()
	if characterID.Valid && characterID.Int64 <= 0 {
		u.Character.Save(db)
	}
	_, err := db.Exec(
		"UPDATE users SET email = $1, password = $2, registered = $3, game_master = $4, character_id = $5",
		u.Email,
		u.getHashedPassword(),
		u.Registered,
		u.GameMaster,
		characterID)
	return err
}

// Save stores a user in database or updates it if it was previously stored.
func (u *User) Save(db *sql.DB) error {
	var err error

	if u.ID <= 0 {
		err = u.create(db)
	} else {
		err = u.update(db)
	}
	if err != nil {
		if sqlstore.IsConstraintError(err) {
			return mwkerr.New(mwkerr.DuplicateCharacter, "Duplicate user with email %s", u.Email)
		}
		return err
	}
	return nil
}

// Load loads a user from database by its email.
func Load(db *sql.DB, email string) (*User, error) {
	var id int64
	var password string
	var registered, gameMaster bool
	var characterID sql.NullInt64

	row := db.QueryRow("SELECT id, password, registered, game_master, character_id FROM users WHERE email = $1", email)
	err := row.Scan(&id, &password, &registered, &gameMaster, &characterID)
	if err != nil {
		return nil, err
	}
	var c *character.Character
	if characterID.Valid {
		char, _ := character.LoadByID(db, characterID.Int64)
		c = char
	}
	return &User{ID: id, Email: email, password: password, Registered: registered, GameMaster: gameMaster, Character: c}, nil
}

// Auth loads a user from database if the email/password match.
func Auth(db *sql.DB, email string, plaintextPassword string) (*User, error) {
	authErr := mwkerr.New(mwkerr.AuthError, "Authentication error")
	u, err := Load(db, email)
	if err != nil {
		return nil, authErr
	}

	err = password.Check(u.password, plaintextPassword)
	if err != nil {
		return nil, authErr
	}

	return u, nil
}
