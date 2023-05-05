package logger

import (
	"fmt"
	"os"
	"strings"

	"github.com/google/uuid"
	"github.com/rifflock/lfshook"
	"github.com/sirupsen/logrus"
)

type Logger struct {
	*logrus.Logger
}

func New() *Logger {

	// Create a new Logrus logger instance
	logger := &Logger{logrus.New()}

	// Set the log level to Info
	if os.Getenv("DKG_LOG_LEVEL") == "true" {
		logger.SetLevel(logrus.DebugLevel)
	} else {
		logger.SetLevel(logrus.InfoLevel)
	}

	logFilePath := fmt.Sprintf("/var/log/dkg_%s.log", generate8digitUUID())

	// Create a new LFS hook to write log messages to a file
	fileHook := lfshook.NewHook(lfshook.PathMap{
		logrus.InfoLevel:  logFilePath,
		logrus.WarnLevel:  logFilePath,
		logrus.ErrorLevel: logFilePath,
		logrus.DebugLevel: logFilePath,
	}, &logrus.JSONFormatter{})

	// Add the LFS hook to the logger
	logger.AddHook(fileHook)
	// Set the logger to not print to the console
	logger.SetOutput(os.Stdout)

	return logger
}

func generate8digitUUID() string {
	return strings.ReplaceAll(uuid.New().String(), "-", "")[:8]
}
