/*
Package config handles configuration of Moenawark. It allows to load
a TOML configuration file and sets the defaults for each configuration
item via struct tags.
*/
package config

import (
	"fmt"
	"github.com/pelletier/go-toml"
	"os"
	"strconv"
	"sync"
)

var defaultConfigStr = `
db_path = "./data/db/moenawark.sqlite"
sql_path = "./sql"
http_listen = ":8080"
api_prefix = "/api"

[auth]
token_length = 32
token_header = "X-Auth-Token"
session_duration = "1h"

[loglevel]
default = "WARN"


[universe]
radius = 1000
min_place_dist = 80
max_way_length = 150
markov_prefix_length = 3

[universe.region]
count = 5
radius = 120
min_place_dist = 20
max_way_length = 40
`

var (
	tree        *toml.Tree
	defaultTree *toml.Tree
	treeLock    = sync.RWMutex{}
)

func init() {
	if err := loadDefaultTree(); err != nil {
		panic(err.Error())
	}
}

func toString(v interface{}) string {
	if str, ok := v.(string); ok {
		return str
	} else if str, ok := v.(fmt.Stringer); ok {
		return str.String()
	} else if i, ok := v.(int64); ok {
		return strconv.FormatInt(i, 10)
	} else if v == nil {
		return ""
	}
	panic("unexpected toml value type")
}

// Get returns a config item value as a string
func Get(key string) string {
	treeLock.RLock()
	defer treeLock.RUnlock()
	if ok := tree.Has(key); !ok {
		return toString(defaultTree.Get(key))
	}
	return toString(tree.Get(key))
}

// GetInt returns a config item value as an int
func GetInt(key string) int {
	i, err := strconv.Atoi(Get(key))
	if err != nil {
		return 0
	}
	return i
}

// LoadFile loads a TOML configuration file
func LoadFile(path string) error {
	if defaultTree == nil {
		err := loadDefaultTree()
		if err != nil {
			return err
		}
	}
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	t, err := toml.LoadFile(path)
	if err != nil {
		return err
	}

	treeLock.Lock()
	defer treeLock.Unlock()
	tree = t

	return nil
}

func loadDefaultTree() error {
	t, err := toml.Load(defaultConfigStr)
	if err != nil {
		return err
	}

	treeLock.Lock()
	defer treeLock.Unlock()
	defaultTree = t

	return nil
}
