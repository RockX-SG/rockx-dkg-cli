package logger

import (
	"fmt"
	"os"
	"regexp"

	"github.com/rifflock/lfshook"
	"github.com/sirupsen/logrus"
)

func New(filename string) *logrus.Logger {
	if !IsLogFilename(filename) {
		return nil
	}

	// Create a new Logrus logger instance
	logger := logrus.New()

	// Set the log level to Info
	logger.SetLevel(logrus.InfoLevel)

	// Create a new LFS hook to write log messages to a file
	logFilePath := fmt.Sprintf("/var/log/%s", filename)
	fileHook := lfshook.NewHook(lfshook.PathMap{
		logrus.InfoLevel:  logFilePath,
		logrus.WarnLevel:  logFilePath,
		logrus.ErrorLevel: logFilePath,
	}, &logrus.JSONFormatter{})

	// Add the LFS hook to the logger
	logger.AddHook(fileHook)

	// Set the logger to not print to the console
	logger.SetOutput(os.Stdout)

	return logger
}

func IsLogFilename(filename string) bool {
	pattern := `^[[:alnum:]_\-]+\.log$`
	match, err := regexp.MatchString(pattern, filename)
	if err != nil {
		return false
	}
	return match
}
