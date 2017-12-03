package config

import (
	"github.com/pelletier/go-toml"
	"log"
	"os"
	"reflect"
	"strconv"
)

// Config holds moenawark's configuration
type Config struct {
	DBPath     string       `toml:"db_path" default:"./data/db/moenawark.sqlite"`
	HTTPListen string       `toml:"http_listen" default:"localhost:8080"`
	APIPrefix  string       `toml:"api_prefix" default:"/api"`
	Auth       authInfo     `toml:"auth"`
	Universe   universeInfo `toml:"universe"`
}

type authInfo struct {
	TokenLength     int    `toml:"token_length" default:"32"`
	TokenHeader     string `toml:"token_header" default:"X-Auth-Token"`
	SessionDuration string `toml:"session_duration" default:"1h"`
}

type universeInfo struct {
	Radius             int        `toml:"radius" default:"1000"`
	MinPlaceDist       int        `toml:"min_place_dist" default:"80"`
	MaxWayLength       int        `toml:"max_way_length" default:"150"`
	MarkovPrefixLength int        `toml:"markov_prefix_length" default:"3"`
	Region             regionInfo `toml:"region"`
}

type regionInfo struct {
	Count        int `toml:"count" default:"5"`
	Radius       int `toml:"radius" default:"120"`
	MinPlaceDist int `toml:"min_place_dist" default:"20"`
	MaxWayLength int `toml:"max_way_length" default:"40"`
}

var (
	// Cfg holds the global configuration used throughout Moenawark
	Cfg Config
)

func setDefaults(conf *Config) {
	t := reflect.TypeOf(*conf)

	if len(conf.DBPath) <= 0 {
		conf.DBPath = getDefault(t, "DBPath")
	}
	if len(conf.HTTPListen) <= 0 {
		conf.HTTPListen = getDefault(t, "HTTPListen")
	}
	if len(conf.APIPrefix) <= 0 {
		conf.APIPrefix = getDefault(t, "APIPrefix")
	}

	t = reflect.TypeOf(conf.Auth)
	if conf.Auth.TokenLength == 0 {
		conf.Auth.TokenLength = getDefaultInt(t, "TokenLength")
	}
	if len(conf.Auth.TokenHeader) <= 0 {
		conf.Auth.TokenHeader = getDefault(t, "TokenHeader")
	}
	if len(conf.Auth.SessionDuration) <= 0 {
		conf.Auth.SessionDuration = getDefault(t, "SessionDuration")
	}

	t = reflect.TypeOf(conf.Universe)
	if conf.Universe.Radius == 0 {
		conf.Universe.Radius = getDefaultInt(t, "Radius")
	}
	if conf.Universe.MinPlaceDist == 0 {
		conf.Universe.MinPlaceDist = getDefaultInt(t, "MinPlaceDist")
	}
	if conf.Universe.MaxWayLength == 0 {
		conf.Universe.MaxWayLength = getDefaultInt(t, "MaxWayLength")
	}
	if conf.Universe.MarkovPrefixLength == 0 {
		conf.Universe.MarkovPrefixLength = getDefaultInt(t, "MarkovPrefixLength")
	}

	t = reflect.TypeOf(conf.Universe.Region)
	if conf.Universe.Region.Count == 0 {
		conf.Universe.Region.Count = getDefaultInt(t, "Count")
	}
	if conf.Universe.Region.Radius == 0 {
		conf.Universe.Region.Radius = getDefaultInt(t, "Radius")
	}
	if conf.Universe.Region.MinPlaceDist == 0 {
		conf.Universe.Region.MinPlaceDist = getDefaultInt(t, "MinPlaceDist")
	}
	if conf.Universe.Region.MaxWayLength == 0 {
		conf.Universe.Region.MaxWayLength = getDefaultInt(t, "MaxWayLength")
	}
}

func getDefault(t reflect.Type, fieldName string) string {
	v, found := t.FieldByName(fieldName)
	if !found {
		log.Fatalf("Missing field %s in %v", fieldName, t)
	}
	return v.Tag.Get("default")
}

func getDefaultInt(t reflect.Type, fieldName string) int {
	x := getDefault(t, fieldName)
	i, err := strconv.Atoi(x)
	if err != nil {
		log.Fatalf("default value of field %s is not an int", fieldName)
	}
	return i
}

func setDefaultsAndGlobal(conf *Config) {
	setDefaults(conf)
	Cfg = *conf
}

// Parse loads the TOML configuration file into a Config struct.
func Parse(path string) (*Config, error) {
	conf := new(Config)
	defer setDefaultsAndGlobal(conf)

	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return conf, nil
		}
		return nil, err
	}

	tree, err := toml.LoadFile(path)
	if err != nil {
		return nil, err
	}

	if err = tree.Unmarshal(conf); err != nil {
		return nil, err
	}

	return conf, nil
}
