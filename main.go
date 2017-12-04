/*
See the README for what Moenawark is.

The main package implements simple subcommands to setup and start Moenawark.
*/
package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/morluque/moenawark/config"
	"github.com/morluque/moenawark/loglevel"
	"github.com/morluque/moenawark/markov"
	"github.com/morluque/moenawark/model"
	"github.com/morluque/moenawark/server"
	"github.com/morluque/moenawark/sqlstore"
	"github.com/morluque/moenawark/universe"
	"os"
)

var (
	// Version of Moenawark
	Version = "dev"
	// BuildDate of Moenawark
	BuildDate = "today"
	log       *loglevel.Logger
)

func init() {
	log = loglevel.New("main", loglevel.Debug)
}

func main() {
	if len(os.Args) <= 1 {
		log.Fatal("Missing first argument <action>")
	}
	action := os.Args[1]

	opts := flag.NewFlagSet("moenawark", flag.PanicOnError)
	var configPath = opts.String("cfg", "moenawark.toml", "path to TOML config file")
	opts.Parse(os.Args[2:])
	log.Infof("config path: %s\n", *configPath)

	_, err := config.Parse(*configPath)
	if err != nil {
		log.Fatal(err)
	}

	switch action {
	case "initdb":
		initDB()
	case "inituniverse":
		initUniverse()
	case "server":
		server.ServeHTTP()
		log.Infof("One day, a server will be started here. But not today.")
	case "version":
		fmt.Printf("Moenawark %s build %s\n", Version, BuildDate)
	default:
		log.Fatal("Unknown action %s", action)
	}
}

func readAdminUser() (*model.User, error) {
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

	u := model.NewUser(login, password)
	u.GameMaster = true
	u.Status = "active"
	return u, nil
}

func initUniverse() {
	db, err := sqlstore.Open(config.Cfg.DBPath)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	ucfg := config.Cfg.Universe
	cfg := universe.Config{
		Radius:       float64(ucfg.Radius),
		MinPlaceDist: float64(ucfg.MinPlaceDist),
		MaxWayLength: float64(ucfg.MaxWayLength),
		MarkovGen:    markov.Load(os.Stdin, ucfg.MarkovPrefixLength),
		RegionConfig: universe.RegionConfig{
			Count:        ucfg.Region.Count,
			Radius:       float64(ucfg.Region.Radius),
			MinPlaceDist: float64(ucfg.Region.MinPlaceDist),
			MaxWayLength: float64(ucfg.Region.MaxWayLength),
		},
	}
	tx, err := db.Begin()
	if err != nil {
		log.Fatal(err)
	}
	u := universe.Generate(cfg, tx)
	err = tx.Commit()
	if err != nil {
		log.Fatal(err)
	}
	if err := u.WriteDotFile("tmp.gv"); err != nil {
		log.Fatal(err)
	}
}

func initDB() {
	db, err := sqlstore.Init(config.Cfg.DBPath)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	admin, err := readAdminUser()
	if err != nil {
		log.Fatal(err)
	}
	tx, err := db.Begin()
	if err != nil {
		log.Fatal(err)
	}
	if err = admin.Save(tx); err != nil {
		log.Fatal(err)
	}
	err = tx.Commit()
	if err != nil {
		log.Fatal(err)
	}
	log.Infof("Created admin user %s", admin.Login)

	data, err := json.Marshal(admin)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("u: %s\n", data)
}
