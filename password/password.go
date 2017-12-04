/*
Package password handles passwords encryption and decryption for Moenawark.

It currently uses bcrypt.
*/
package password

import (
	"encoding/hex"
	"github.com/morluque/moenawark/config"
	"github.com/morluque/moenawark/loglevel"
	"golang.org/x/crypto/bcrypt"
)

var log *loglevel.Logger

func init() {
	log = loglevel.New("password", loglevel.Debug)
}

// ReloadConfig performs required actions to reload all dynamic config.
func ReloadConfig() {
	log.SetLevelName(config.Get("loglevel.password"))
}

// Encode transforms a plaintext password into an unrecoverable hash.
// Uses bcrypt currently.
func Encode(plaintext string) string {
	b, err := bcrypt.GenerateFromPassword([]byte(plaintext), bcrypt.DefaultCost)
	if err != nil {
		log.Fatal(err)
	}
	return hex.EncodeToString(b)
}

// Check verifies that an encoded password and a plaintext match.
func Check(hashedPassword string, plaintext string) error {
	hash, err := hex.DecodeString(hashedPassword)
	if err != nil {
		return err
	}
	return bcrypt.CompareHashAndPassword(hash, []byte(plaintext))
}
