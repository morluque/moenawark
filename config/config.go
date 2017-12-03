package config

import (
	"github.com/pelletier/go-toml"
	"log"
	"os"
	"time"
)

type duration struct {
	time.Duration
}

func (d *duration) UnmarshalText(text []byte) error {
	var err error
	d.Duration, err = time.ParseDuration(string(text))
	return err
}

// Config holds moenawark's configuration
type Config struct {
	DBPath     string       `toml:"db_path"`
	HTTPListen string       `toml:"http_listen"`
	APIPrefix  string       `toml:"api_prefix"`
	Auth       authInfo     `toml:"auth"`
	Universe   universeInfo `toml:"universe"`
}

type authInfo struct {
	TokenLength     int      `toml:"token_length"`
	TokenHeader     string   `toml:"token_header"`
	SessionDuration duration `toml:"session_duration"`
}

type universeInfo struct {
	Radius             int        `toml:"radius"`
	MinPlaceDist       int        `toml:"min_place_dist"`
	MaxWayLength       int        `toml:"max_way_length"`
	MarkovPrefixLength int        `toml:"markov_prefix_length"`
	Region             regionInfo `toml:"region"`
}

type regionInfo struct {
	Count        int `toml:"count"`
	Radius       int `toml:"radius"`
	MinPlaceDist int `toml:"min_place_dist"`
	MaxWayLength int `toml:"max_way_length"`
}

var (
	// Cfg holds the global configuration used throughout Moenawark
	Cfg Config
)

// Parse loads the TOML configuration file into a Config struct.
func Parse(path string) (*Config, error) {
	sessionDuration, err := time.ParseDuration("1h")
	if err != nil {
		log.Fatal("WTFBBQ!")
	}
	conf := new(Config)
	conf.DBPath = "./data/db/moenawark.sqlite"
	conf.HTTPListen = "localhost:8080"
	conf.APIPrefix = "/api"
	conf.Auth = authInfo{
		TokenLength:     32,
		TokenHeader:     "X-Auth-Token",
		SessionDuration: duration{Duration: sessionDuration},
	}
	conf.Universe = universeInfo{
		Radius:             1000,
		MinPlaceDist:       80,
		MaxWayLength:       150,
		MarkovPrefixLength: 3,
		Region: regionInfo{
			Count:        5,
			Radius:       120,
			MinPlaceDist: 20,
			MaxWayLength: 40,
		},
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return conf, nil
	}
	tree, err := toml.LoadFile(path)
	if err != nil {
		return nil, err
	}
	if err = tree.Unmarshal(&conf); err != nil {
		return nil, err
	}

	Cfg = *conf

	return conf, nil
}
