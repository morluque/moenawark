/*
Package model holds the model objects for Moenawark.
*/
package model

import (
	"github.com/morluque/moenawark/loglevel"
)

var log *loglevel.Logger

func init() {
	log = loglevel.New("model", loglevel.Debug)
}

// LogLevel dynamically sets the log level for this package.
func LogLevel(level string) {
	log.SetLevelName(level)
}
