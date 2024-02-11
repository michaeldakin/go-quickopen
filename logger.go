package main

import (
	"io"
	"log"
	"log/slog"
	"os"
)

// Call NewLogger to initialise slog
// No usage of structs yet, not sure if required
func NewLogger() *slog.Logger {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}

	logFileHandler, err := os.OpenFile(homeDir+"/personal/go/go-quickopen/quickopen.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil { // returns *PathError
		log.Fatal(err)
	}
	defer logFileHandler.Close()

	// If debug flag is set, write logs to both stdout and log file
	// Otherwise, write logs to log file only
	var writer io.Writer
	if *Debug {
		writer = io.MultiWriter(os.Stdout, logFileHandler)
	} else {
		writer = logFileHandler
	}

	// Set LevelDebug to default so we always log Debug/Info/Warn/Error
	opts := &slog.HandlerOptions{
		Level:     slog.LevelDebug,
		AddSource: true,
	}

	return slog.New(slog.NewJSONHandler(writer, opts))
}
