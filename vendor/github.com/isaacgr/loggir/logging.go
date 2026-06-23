package loggir

import (
	"log/slog"
	"os"
)

// Returns the singleton logging instance
func GetLogger(module string, debug bool) *slog.Logger {
	loglevel := slog.LevelInfo
	if debug == true {
		loglevel = slog.LevelDebug
	}
	opts := options{Level: loglevel}
	logger := slog.New(newModuleHandler(os.Stdout, &opts))
	logger = logger.With("module", module)
	return logger
}
