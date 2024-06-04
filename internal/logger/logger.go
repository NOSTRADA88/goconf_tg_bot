// Package logger provides a simple logging interface and an implementation that writes to stdout.
package logger

import (
	"log"
	"os"
)

// These constants define the logging levels.
const (
	DebugLevel = iota // DebugLevel logs everything.
	InfoLevel         // InfoLevel logs Info, Warnings and Errors.
	WarnLevel         // WarnLevel logs Warnings and Errors.
	ErrorLevel        // ErrorLevel logs only Errors.
)

// Logger is an interface for logging.
type Logger interface {
	Info(args ...interface{})                  // Info logs routine information about program operation.
	Warn(args ...interface{})                  // Warn logs information about potentially harmful situations.
	Error(args ...interface{})                 // Error logs information about error conditions.
	Debug(args ...interface{})                 // Debug logs information useful to developers for debugging the application.
	InfoF(format string, args ...interface{})  // InfoF is like Info but supports formatting.
	WarnF(format string, args ...interface{})  // WarnF is like Warn but supports formatting.
	ErrorF(format string, args ...interface{}) // ErrorF is like Error but supports formatting.
	DebugF(format string, args ...interface{}) // DebugF is like Debug but supports formatting.
}

// Slog is an implementation of Logger that writes to stdout.
type Slog struct {
	sLogger *log.Logger // sLogger is the standard logger used to print the logs.
	level   int         // level is the current logging level.
}

// New creates a new Slog with the given logging level.
func New(level int) *Slog {
	return &Slog{sLogger: log.New(os.Stdout, "", log.LstdFlags), level: level}
}

// SetLevel sets the logging level.
func (sl *Slog) SetLevel(level int) {
	sl.level = level
}

// Info logs routine information about program operation.
func (sl *Slog) Info(args ...interface{}) {
	if sl.level <= InfoLevel {
		sl.sLogger.SetPrefix("INFO: ")
		sl.sLogger.Println(args...)
	}
}

// Warn logs information about potentially harmful situations.
func (sl *Slog) Warn(args ...interface{}) {
	if sl.level <= WarnLevel {
		sl.sLogger.SetPrefix("WARN: ")
		sl.sLogger.Println(args...)
	}
}

// Error logs information about error conditions.
func (sl *Slog) Error(args ...interface{}) {
	if sl.level <= ErrorLevel {
		sl.sLogger.SetPrefix("ERROR: ")
		sl.sLogger.Println(args...)
	}
}

// Debug logs information useful to developers for debugging the application.
func (sl *Slog) Debug(args ...interface{}) {
	if sl.level <= DebugLevel {
		sl.sLogger.SetPrefix("DEBUG: ")
		sl.sLogger.Println(args...)
	}
}

// InfoF is like Info but supports formatting.
func (sl *Slog) InfoF(format string, args ...interface{}) {
	if sl.level <= InfoLevel {
		sl.sLogger.SetPrefix("INFO: ")
		sl.sLogger.Printf(format, args...)
	}
}

// WarnF is like Warn but supports formatting.
func (sl *Slog) WarnF(format string, args ...interface{}) {
	if sl.level <= WarnLevel {
		sl.sLogger.SetPrefix("WARN: ")
		sl.sLogger.Printf(format, args...)
	}
}

// ErrorF is like Error but supports formatting.
func (sl *Slog) ErrorF(format string, args ...interface{}) {
	if sl.level <= ErrorLevel {
		sl.sLogger.SetPrefix("ERROR: ")
		sl.sLogger.Printf(format, args...)
	}
}

// DebugF is like Debug but supports formatting.
func (sl *Slog) DebugF(format string, args ...interface{}) {
	if sl.level <= DebugLevel {
		sl.sLogger.SetPrefix("DEBUG: ")
		sl.sLogger.Printf(format, args...)
	}
}
