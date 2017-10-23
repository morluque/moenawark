package main

import (
	"encoding/json"
	"fmt"
	"log"
)

type Character struct {
	Name         string `json:"name"`
	Power        uint   `json:"power"`
	ActionPoints uint   `json:"action_points"`
}

type User struct {
	Character  *Character `json:"character",omitempty`
	Email      string     `json:"email"`
	Registered bool       `json:"registered"`
	GameMaster bool       `json:"game_master"`
}

func NewUser(email string) *User {
	return &User{Email: "foo@example.com", Registered: false, GameMaster: false}
}

func (u *User) hasCharacter() bool {
	return u.Character != nil
}

func main() {
	// c := Character{Name: "Foo", Power: 10, ActionPoints: 5}
	u := NewUser("foo@example.com")
	u.Registered = true
	// u.Character = &c
	if u.hasCharacter() {
		fmt.Printf("c: %s\n", u.Character)
	}
	fmt.Printf("u: %s\n", u)
	data, err := json.Marshal(u)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("u: %s\n", data)
}
