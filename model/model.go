/*
Package model holds the model objects for Moenawark.
*/
package model

import (
	"github.com/morluque/moenawark/config"
	"github.com/morluque/moenawark/loglevel"
)

var log *loglevel.Logger

func init() {
	log = loglevel.New("model", loglevel.Debug)
}

// ReloadConfig performs required actions to reload all dynamic config.
func ReloadConfig() {
	log.SetLevelName(config.Get("loglevel.model"))
}
