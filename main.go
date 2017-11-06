package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/morluque/moenawark/config"
	"github.com/morluque/moenawark/model/character"
	"github.com/morluque/moenawark/model/user"
	"github.com/morluque/moenawark/sqlstore"
	"log"
)

func main() {
	var configPath = flag.String("cfg", "moenawark.toml", "path to TOML config file")
	flag.Parse()
	fmt.Printf("config path: %s\n", *configPath)

	conf, err := config.Parse(*configPath)
	if err != nil {
		log.Fatal(err)
	}

	c := character.New("Foo", 10, 5)
	u := user.New("foo@example.com", "secret")
	u.Status = "active"
	u.Character = c
	if u.HasCharacter() {
		fmt.Printf("c: %v\n", u.Character)
	}

	var dbPath string
	if len(conf.DBPath) > 0 {
		dbPath = conf.DBPath
	} else {
		dbPath = "data/db/moenawark.sqlite"
	}
	db, err := sqlstore.Open(dbPath)
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
