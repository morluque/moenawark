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
	"github.com/morluque/moenawark/mwkerr"
	"github.com/morluque/moenawark/password"
	"github.com/morluque/moenawark/server"
	"github.com/morluque/moenawark/sqlstore"
	"github.com/morluque/moenawark/universe"
	"os"
	"os/signal"
	"syscall"
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

func handleSignals(reloadFunc func()) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGUSR1)
	for {
		s := <-c
		if s == syscall.SIGUSR1 {
			reloadFunc()
		}
	}
}

func reloadConfig(path string) {
	log.Infof("reloading configuration")
	err := config.LoadFile(path)
	if err != nil {
		log.Errorf(err.Error())
	}
	log.SetLevelName(config.Get("loglevel.main"))
	server.ReloadConfig()
	markov.ReloadConfig()
	model.ReloadConfig()
	mwkerr.ReloadConfig()
	password.ReloadConfig()
	sqlstore.ReloadConfig()
	universe.ReloadConfig()
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

	reloadConfig(*configPath)
	go handleSignals(func() {
		reloadConfig(*configPath)
	})

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
	db, err := sqlstore.Open(config.Get("db_path"))
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	cfg := universe.Config{
		Radius:       float64(config.GetInt("universe.radius")),
		MinPlaceDist: float64(config.GetInt("universe.min_place_dist")),
		MaxWayLength: float64(config.GetInt("universe.max_way_length")),
		MarkovGen:    markov.Load(os.Stdin, config.GetInt("universe.markov_prefix_length")),
		RegionConfig: universe.RegionConfig{
			Count:        config.GetInt("universe.region.count"),
			Radius:       float64(config.GetInt("universe.region.radius")),
			MinPlaceDist: float64(config.GetInt("universe.region.min_place_dist")),
			MaxWayLength: float64(config.GetInt("universe.region.max_way_length")),
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
	db, err := sqlstore.Init(config.Get("db_path"))
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
