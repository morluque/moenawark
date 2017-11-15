package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/morluque/moenawark/config"
	"github.com/morluque/moenawark/model/user"
	"github.com/morluque/moenawark/sqlstore"
	"github.com/morluque/moenawark/universe"
	"log"
	"os"
)

func main() {
	if len(os.Args) <= 1 {
		log.Fatal("Missing first argument <action>")
	}
	action := os.Args[1]

	switch action {
	case "initdb":
		initDB()
	case "inituniverse":
		initUniverse()
	case "server":
		log.Print("One day, a server will be started here. But not today.")
	default:
		log.Fatalf("Unknown action %s", action)
	}
}

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

func loadConfig() *config.Config {
	opts := flag.NewFlagSet("moenawark", flag.PanicOnError)
	var configPath = opts.String("cfg", "moenawark.toml", "path to TOML config file")
	opts.Parse(os.Args[2:])
	log.Printf("config path: %s\n", *configPath)

	conf, err := config.Parse(*configPath)
	if err != nil {
		log.Fatal(err)
	}

	return conf
}

func initUniverse() {
	//conf := loadConfig()

	cfg := universe.Config{
		Radius:       300,
		MinPlaceDist: 50,
		MaxWayLength: 100,
		RegionConfig: universe.RegionConfig{
			Count:        3,
			Radius:       80,
			MinPlaceDist: 20,
			MaxWayLength: 30,
		},
	}
	u := universe.Generate(cfg)
	if err := u.WriteDotFile("tmp.gv"); err != nil {
		log.Fatal(err)
	}
}

func initDB() {
	conf := loadConfig()

	var dbPath string
	if len(conf.DBPath) > 0 {
		dbPath = conf.DBPath
	} else {
		dbPath = "data/db/moenawark.sqlite"
	}
	db, err := sqlstore.Init(dbPath)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	admin, err := readAdminUser()
	if err != nil {
		log.Fatal(err)
	}
	if err = admin.Save(db); err != nil {
		log.Fatal(err)
	}
	log.Printf("Created admin user %s", admin.Login)

	data, err := json.Marshal(admin)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("u: %s\n", data)
}
