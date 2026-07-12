package logger

import (
	"io"
	"os"
	"sync"
	"time"

	"github.com/rs/zerolog"
)

// logBuffer is a thread-safe circular buffer for storing recent log lines
type logBuffer struct {
	mu    sync.RWMutex
	lines []string
	limit int
}

func (b *logBuffer) Write(p []byte) (n int, err error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.lines = append(b.lines, string(p))
	if len(b.lines) > b.limit {
		b.lines = b.lines[1:]
	}
	return len(p), nil
}

func (b *logBuffer) GetLines() []string {
	b.mu.RLock()
	defer b.mu.RUnlock()
	copied := make([]string, len(b.lines))
	copy(copied, b.lines)
	return copied
}

// ActiveLogBuffer holds the latest 500 daemon logs
var ActiveLogBuffer = &logBuffer{limit: 500}

// Log is the global logger instance
var Log zerolog.Logger

// Setup initializes the global logger
func Setup(verbose bool, daemon bool) {
	var output io.Writer = os.Stdout

	if !daemon {
		// Output structured but readable console logger for CLI usage
		output = zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: time.RFC3339,
		}
	}

	level := zerolog.InfoLevel
	if verbose {
		level = zerolog.DebugLevel
	}

	zerolog.SetGlobalLevel(level)

	// Route logs to both standard output writer and in-memory buffer
	multi := zerolog.MultiLevelWriter(output, ActiveLogBuffer)
	Log = zerolog.New(multi).With().Timestamp().Logger()
}

// Info logs an info message
func Info(format string, v ...interface{}) {
	Log.Info().Msgf(format, v...)
}

// Debug logs a debug message
func Debug(format string, v ...interface{}) {
	Log.Debug().Msgf(format, v...)
}

// Warn logs a warning message
func Warn(format string, v ...interface{}) {
	Log.Warn().Msgf(format, v...)
}

// Error logs an error message
func Error(err error, format string, v ...interface{}) {
	if err != nil {
		Log.Error().Err(err).Msgf(format, v...)
	} else {
		Log.Error().Msgf(format, v...)
	}
}

// Fatal logs a fatal error and calls os.Exit(1)
func Fatal(err error, format string, v ...interface{}) {
	if err != nil {
		Log.Fatal().Err(err).Msgf(format, v...)
	} else {
		Log.Fatal().Msgf(format, v...)
	}
}
