package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/morluque/moenawark/config"
	"github.com/morluque/moenawark/model/user"
	"github.com/morluque/moenawark/sqlstore"
	"log"
	"os"
)

func readAdminUser() (*user.User, error) {
	var login, password string
	scanner := bufio.NewScanner(os.Stdin)

	fmt.Print("Admin login: ")
	if !scanner.Scan() {
		return nil, scanner.Err()
	}
	login = scanner.Text()
	fmt.Print("Admin password: ")
	if !scanner.Scan() {
		return nil, scanner.Err()
	}
	password = scanner.Text()

	u := user.New(login, password)
	u.GameMaster = true
	u.Status = "active"
	return u, nil
}

func main() {
	var configPath = flag.String("cfg", "moenawark.toml", "path to TOML config file")
	flag.Parse()
	fmt.Printf("config path: %s\n", *configPath)

	conf, err := config.Parse(*configPath)
	if err != nil {
		log.Fatal(err)
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

	var adminLogin string
	if len(conf.AdminLogin) > 0 {
		adminLogin = conf.AdminLogin
	} else {
		adminLogin = "admin"
	}
	log.Printf("Admin login is %s\n", adminLogin)
	admin, err := user.Load(db, adminLogin)
	if err != nil {
		log.Print(err)
		admin, err = readAdminUser()
		if err != nil {
			log.Fatal(err)
		}
		if err = admin.Save(db); err != nil {
			log.Fatal(err)
		}
		log.Printf("Created admin user %s", admin.Login)
	}

	data, err := json.Marshal(admin)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("u: %s\n", data)
	db.Close()
}
