package loglevel

import (
	"fmt"
	"io"
	"log"
	"os"
)

type Level uint8

type LevelFilter struct {
	Levels   map[Level]string
	MinLevel Level
	Writer   io.Writer
}

type Logger struct {
	logger *log.Logger
	filter LevelFilter
	prefix string
}

const (
	Debug = iota
	Info
	Warn
	Error
)

func New(prefix string, minLevel Level) *Logger {
	l := log.New(os.Stdout, "", log.Ldate|log.Lmicroseconds|log.LUTC)
	levels := make(map[Level]string)
	levels[Debug] = "debug"
	levels[Info] = "info"
	levels[Warn] = "warn"
	levels[Error] = "error"
	filter := LevelFilter{
		Levels:   levels,
		MinLevel: minLevel,
		Writer:   os.Stdout,
	}

	return &Logger{logger: l, filter: filter, prefix: prefix}
}

func (l *Logger) SetLevel(level Level) {
	l.filter.MinLevel = level
}

func (l *Logger) printf(level Level, format string, v ...interface{}) {
	if l.filter.MinLevel >= level {
		l.logger.Printf(fmt.Sprintf("[%s] %s: %s", l.filter.Levels[level], l.prefix, format), v...)
	}
}

func (l *Logger) Debugf(format string, v ...interface{}) {
	l.printf(Debug, format, v...)
}

func (l *Logger) Infof(format string, v ...interface{}) {
	l.printf(Info, format, v...)
}

func (l *Logger) Warnf(format string, v ...interface{}) {
	l.printf(Warn, format, v...)
}

func (l *Logger) Errorf(format string, v ...interface{}) {
	l.printf(Error, format, v...)
}

func (l *Logger) Fatalf(format string, v ...interface{}) {
	l.logger.Fatalf(format, v...)
}
