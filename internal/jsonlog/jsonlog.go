package jsonlog

import (
	"encoding/json"
	"io"
	"os"
	"runtime/debug"
	"sync"
	"time"

	"github.com/michaelcjefferson/kamar-listener/internal/data"
)

type Level int8

// iota works similar to auto-increment - starting at 0, each successive const is assigned a sequential int. In other words, LevelInfo is a Level set as 0, LevelOff is a Level set to 3 etc.
const (
	LevelInfo Level = iota
	LevelDebug
	LevelError
	LevelFatal
	LevelOff
)

func (l Level) String() string {
	switch l {
	case LevelInfo:
		return "INFO"
	case LevelDebug:
		return "DEBUG"
	case LevelError:
		return "ERROR"
	case LevelFatal:
		return "FATAL"
	default:
		return ""
	}
}

// The mutex (mutual exclusion lock) prevents two log triggers from trying to write at the same time (which would lead to jumbled log messages). logModel is optional, and provides the ability to simultaneously write logs to the database if logModel exists.
type Logger struct {
	out      io.Writer
	minLevel Level
	mu       sync.Mutex
	logModel *data.LogModel
}

func New(out io.Writer, minLevel Level, logModel *data.LogModel) *Logger {
	return &Logger{
		out:      out,
		minLevel: minLevel,
		logModel: logModel,
	}
}

func (l *Logger) PrintInfo(message string, properties map[string]any) {
	l.print(LevelInfo, message, properties)
}

func (l *Logger) PrintDebug(message string, properties map[string]any) {
	l.print(LevelDebug, message, properties)
}

func (l *Logger) PrintError(err error, properties map[string]any) {
	l.print(LevelError, err.Error(), properties)
}

func (l *Logger) PrintFatal(err error, properties map[string]any) {
	l.print(LevelFatal, err.Error(), properties)
	// As it is a fatal error, terminate the application
	os.Exit(1)
}

// As print is an internal function only (PrintInfo/Debug/Error/Fatal are the only ones that will be called from outside this package), it is not capitalised.
func (l *Logger) print(level Level, message string, properties map[string]any) (int, error) {
	if level < l.minLevel {
		return 0, nil
	}

	aux := data.Log{
		Level: level.String(),
		Time:  time.Now().UTC(),
		// Time:       time.Now().UTC().Format(time.RFC3339),
		Message:    message,
		Properties: properties,
	}

	// If log is at least debug level, include a stacktrace in the log
	if level >= LevelDebug {
		aux.Trace = string(debug.Stack())
	}

	// Write log to database in secondary thread, if logModel exists
	if l.logModel != nil {
		l.logModel.Insert(&aux)
	}
	// if l.logModel != nil {
	// 	go func() {
	// 		err := l.logModel.Insert(&aux)
	// 		if err != nil {
	// 			mes := []byte(LevelError.String() + "failed to write log to database:" + err.Error())
	// 			l.out.Write(mes)
	// 		}
	// 	}()
	// }

	// Create the line constituting the log and populate it with all info in aux marshalled to JSON. If that fails, create a log line recording that error instead.
	var line []byte

	line, err := json.Marshal(aux)
	if err != nil {
		line = []byte(LevelError.String() + ": unable to marshal log message:" + err.Error())
	}

	// Lock mutex, and defer unlock until function returns with the result of the log write operation.
	l.mu.Lock()
	defer l.mu.Unlock()

	return l.out.Write(append(line, '\n'))
}

// TODO: What is the reason for LevelError specifically here? Check Let's Go
func (l *Logger) Write(message []byte) (n int, err error) {
	return l.print(LevelError, string(message), nil)
}
