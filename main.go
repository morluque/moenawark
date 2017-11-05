package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/morluque/moenawark/model/character"
	"github.com/morluque/moenawark/model/user"
	"github.com/morluque/moenawark/sqlstore"
	"log"
)

func main() {
	var dbpath = flag.String("dbpath", "data/db/moenawark.sqlite", "path to DB file")
	flag.Parse()
	dataSource := fmt.Sprintf("file:%s", *dbpath)
	fmt.Printf("DB path: %s\n", dataSource)

	c := character.New("Foo", 10, 5)
	u := user.New("foo@example.com", "secret")
	u.Registered = true
	u.Character = c
	if u.HasCharacter() {
		fmt.Printf("c: %v\n", u.Character)
	}

	db, err := sqlstore.Open(dataSource)
	if err != nil {
		log.Fatal(err)
	}

	if err := c.Save(db); err != nil {
		log.Fatal(err)
	}
	c.Power = 20
	if err := c.Save(db); err != nil {
		log.Fatal(err)
	}

	fmt.Printf("u: %v\n", u)
	if err := u.Save(db); err != nil {
		log.Fatal(err)
	}

	uu, err := user.Auth(db, "foo@example.com", "secret")
	if err != nil {
		log.Fatal("Can't load user foo@example.com")
	}

	data, err := json.Marshal(uu)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("u: %s\n", data)
}
