/*
Package loglevel handles level-based logging for Moenawark.

It allows to define one logger per package and dynamically set the log
level of each logger.
*/
package loglevel

import (
	"fmt"
	"github.com/morluque/moenawark/config"
	"io"
	"log"
	"os"
	"strings"
)

// Level represents of log level as an unsigned integer
type Level uint8

type levelFilter struct {
	levels   map[Level]string
	minLevel Level
	writer   io.Writer
}

// Logger is a logger with a minimum level; only messages above that level will be printed.
type Logger struct {
	logger *log.Logger
	filter levelFilter
	prefix string
}

// Available semi-standard log levels
const (
	Unknown = iota
	Debug
	Info
	Warn
	Error
	None
)

// New returns a new Logger with given prefix and minimum level.
func New(prefix string, minLevel Level) *Logger {
	l := log.New(os.Stdout, "", log.Ldate|log.Lmicroseconds|log.LUTC)
	levels := make(map[Level]string)
	levels[Unknown] = "unknown"
	levels[Debug] = "debug"
	levels[Info] = "info"
	levels[Warn] = "warn"
	levels[Error] = "error"
	filter := levelFilter{
		levels:   levels,
		minLevel: minLevel,
		writer:   os.Stdout,
	}

	return &Logger{logger: l, filter: filter, prefix: prefix}
}

// SetLevel dynamically sets the minimum log level
func (l *Logger) SetLevel(level Level) {
	l.filter.minLevel = level
}

func (l *Logger) getLevelByName(levelName string) Level {
	lowerLevelName := strings.ToLower(levelName)
	for level, name := range l.filter.levels {
		if name == lowerLevelName {
			return level
		}
	}
	return Unknown
}

// SetLevelName dynamically sets the minimum log level from a string name.
func (l *Logger) SetLevelName(levelName string) {
	level := l.getLevelByName(levelName)
	if level == Unknown {
		level = l.getLevelByName(config.Get("loglevel.default"))
	}
	l.SetLevel(level)
}

func (l *Logger) printf(level Level, format string, v ...interface{}) {
	if level >= l.filter.minLevel {
		l.logger.Printf(fmt.Sprintf("[%s] %s: %s", l.filter.levels[level], l.prefix, format), v...)
	}
}

// Debugf prints a formatted message if minimum level is less than or equal to Debug.
func (l *Logger) Debugf(format string, v ...interface{}) {
	l.printf(Debug, format, v...)
}

// Infof prints a formatted message if minimum level is less than or equal to Info.
func (l *Logger) Infof(format string, v ...interface{}) {
	l.printf(Info, format, v...)
}

// Warnf prints a formatted message if minimum level is less than or equal to Warn.
func (l *Logger) Warnf(format string, v ...interface{}) {
	l.printf(Warn, format, v...)
}

// Errorf prints a formatted message if minimum level is less than or equal to Error.
func (l *Logger) Errorf(format string, v ...interface{}) {
	l.printf(Error, format, v...)
}

// Fatalf calls the underlying log.Logger's method Fatalf(); this will terminate your program.
func (l *Logger) Fatalf(format string, v ...interface{}) {
	l.logger.Fatalf(format, v...)
}

// Fatal calls the underlying log.Logger's method Fatal(); this will terminate your program.
func (l *Logger) Fatal(v ...interface{}) {
	l.logger.Fatal(v...)
}
