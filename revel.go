package revel

import (
	"fmt"
	"runtime"
	"strings"

	"github.com/revel/revel"
	"gopkg.in/birkirb/loggers.v1"
	"gopkg.in/birkirb/loggers.v1/mappers"
)

// Logger is a Contextual logger wrapper over Revel's logger.
type Logger struct{}

// NewLogger returns a Contextual Logger for revel's internal logger.
// Note that Revel's loggers must be initialized before any logging can be made.
func NewLogger() loggers.Contextual {
	var l *Logger
	var a = mappers.NewContextualMap(l)
	//a.Info("Now using Revel's logger package (via loggers/mappers/revel).\n")
	return a
}

// LevelPrint is a Mapper method
func (l *Logger) LevelPrint(lev mappers.Level, i ...interface{}) {
	i = append([]interface{}{caller(3) + " "}, i...)
	getRevelLevel(lev)(fmt.Sprint(i...))
}

// LevelPrintf is a Mapper method
func (l *Logger) LevelPrintf(lev mappers.Level, format string, i ...interface{}) {
	getRevelLevel(lev)(fmt.Sprintf(caller(3)+" "+format, i...))
}

// LevelPrintln is a Mapper method
func (l *Logger) LevelPrintln(lev mappers.Level, i ...interface{}) {
	i = append([]interface{}{caller(3)}, i...)
	getRevelLevel(lev)(fmt.Sprintln(i...))
}

// WithField returns an advanced logger with a pre-set field.
func (l *Logger) WithField(key string, value interface{}) loggers.Advanced {
	return l.WithFields(key, value)
}

// WithFields returns an advanced logger with pre-set fields.
func (l *Logger) WithFields(fields ...interface{}) loggers.Advanced {
	s := make([]string, len(fields)/2)
	for i := 0; i+1 < len(fields); i = i + 2 {
		key := fields[i]
		value := fields[i+1]
		s = append(s, fmt.Sprint(key, "=", value))
	}

	r := revelPostfixLogger{strings.Join(s, " ")}
	return mappers.NewAdvancedMap(&r)
}

type revelPostfixLogger struct {
	postfix string
}

func (r *revelPostfixLogger) LevelPrint(lev mappers.Level, i ...interface{}) {
	i = append(i, r.postfix)
	getRevelLevel(lev)(fmt.Sprint(i...))
}

func (r *revelPostfixLogger) LevelPrintf(lev mappers.Level, format string, i ...interface{}) {
	if len(r.postfix) > 0 {
		format = format + " %s"
		i = append(i, r.postfix)
	}
	getRevelLevel(lev)(fmt.Sprintf(format, i...))
}

func (r *revelPostfixLogger) LevelPrintln(lev mappers.Level, i ...interface{}) {
	i = append(i, r.postfix)
	getRevelLevel(lev)(fmt.Sprintln(i...))
}

func getRevelLevel(lev mappers.Level) func(string, ...interface{}) {
	switch lev {
	case mappers.LevelDebug:
		return revel.AppLog.Debug
	case mappers.LevelInfo:
		return revel.AppLog.Info
	case mappers.LevelWarn:
		return revel.AppLog.Warn
	case mappers.LevelError:
		return revel.AppLog.Error
	case mappers.LevelFatal:
		return revel.AppLog.Fatal
	case mappers.LevelPanic:
		return revel.AppLog.Panic
	default:
		panic("unreachable")
	}
}

// shortenFile returns the folder and file name of an absolute file path.
func shortenFile(file string) string {
	short := file
	foundOne := false
	for i := len(file) - 1; i > 0; i-- {
		if file[i] == '/' {
			if !foundOne {
				foundOne = true
				continue
			}
			short = file[i+1:]
			break
		}
	}
	return short
}

// caller returns the funtion call line at the specified depth
// as "dir/file.go:n:
func caller(depth int) string {
	_, file, line, ok := runtime.Caller(depth + 1)
	if !ok {
		file = "???"
		line = 0
	}
	return fmt.Sprintf("%s:%d:", shortenFile(file), line)
}
