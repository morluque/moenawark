package password

import (
	"encoding/hex"
	"golang.org/x/crypto/bcrypt"
	"log"
)

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