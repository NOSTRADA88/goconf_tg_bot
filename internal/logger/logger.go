package logger

import (
	"log"
	"os"
)

const (
	DebugLevel = iota
	InfoLevel
	WarnLevel
	ErrorLevel
)

type Logger interface {
	Info(args ...interface{})
	Warn(args ...interface{})
	Error(args ...interface{})
	Debug(args ...interface{})
	InfoF(format string, args ...interface{})
	WarnF(format string, args ...interface{})
	ErrorF(format string, args ...interface{})
	DebugF(format string, args ...interface{})
}

type Slog struct {
	sLogger *log.Logger
	level   int
}

func New(level int) *Slog {
	return &Slog{sLogger: log.New(os.Stdout, "", log.LstdFlags), level: level}
}

func (sl *Slog) SetLevel(level int) {
	sl.level = level
}

func (sl *Slog) Info(args ...interface{}) {
	if sl.level <= InfoLevel {
		sl.sLogger.SetPrefix("INFO: ")
		sl.sLogger.Println(args...)
	}
}

func (sl *Slog) Warn(args ...interface{}) {
	if sl.level <= WarnLevel {
		sl.sLogger.SetPrefix("WARN: ")
		sl.sLogger.Println(args...)
	}
}

func (sl *Slog) Error(args ...interface{}) {
	if sl.level <= ErrorLevel {
		sl.sLogger.SetPrefix("ERROR: ")
		sl.sLogger.Println(args...)
	}
}

func (sl *Slog) Debug(args ...interface{}) {
	if sl.level <= DebugLevel {
		sl.sLogger.SetPrefix("DEBUG: ")
		sl.sLogger.Println(args...)
	}
}

func (sl *Slog) InfoF(format string, args ...interface{}) {
	if sl.level <= InfoLevel {
		sl.sLogger.SetPrefix("INFO: ")
		sl.sLogger.Printf(format, args...)
	}
}

func (sl *Slog) WarnF(format string, args ...interface{}) {
	if sl.level <= WarnLevel {
		sl.sLogger.SetPrefix("WARN: ")
		sl.sLogger.Printf(format, args...)
	}
}

func (sl *Slog) ErrorF(format string, args ...interface{}) {
	if sl.level <= ErrorLevel {
		sl.sLogger.SetPrefix("ERROR: ")
		sl.sLogger.Printf(format, args...)
	}
}

func (sl *Slog) DebugF(format string, args ...interface{}) {
	if sl.level <= DebugLevel {
		sl.sLogger.SetPrefix("DEBUG: ")
		sl.sLogger.Printf(format, args...)
	}
}
