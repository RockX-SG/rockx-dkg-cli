package logger

import (
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rifflock/lfshook"
	"github.com/sirupsen/logrus"
)

type Logger struct {
	*logrus.Logger
}

func New(logFilePath string) *Logger {

	// Create a new Logrus logger instance
	logger := &Logger{logrus.New()}

	// Set the log level to Info
	if os.Getenv("DKG_LOG_LEVEL") == "release" {
		logger.SetLevel(logrus.InfoLevel)
	} else {
		logger.SetLevel(logrus.DebugLevel)
	}

	// basePath := "/var/log"
	// if os.Getenv("DKG_CLI_LOG_BASE_PATH") != "" {
	// 	basePath = os.Getenv("DKG_CLI_LOG_BASE_PATH")
	// }

	// filename := fmt.Sprintf("dkg_%s.log", generate8digitUUID())
	// logFilePath := fmt.Sprintf("%s/%s", basePath, filename)

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

	logger.Infof("writing logs to: %s", logFilePath)
	return logger
}

func GinLogger(logger *Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		// Process request
		c.Next()

		end := time.Now()
		latency := end.Sub(start)

		clientIP := c.ClientIP()
		method := c.Request.Method
		statusCode := c.Writer.Status()
		errorMessage := c.Errors.ByType(gin.ErrorTypePrivate).String()

		entry := logger.WithFields(logrus.Fields{
			"status_code": statusCode,
			"latency":     latency,
			"client_ip":   clientIP,
			"method":      method,
			"path":        path,
			"query":       raw,
			"error":       errorMessage,
		})

		switch {
		case statusCode >= 500:
			entry.Error()
		case statusCode >= 400:
			entry.Warn()
		default:
			entry.Info()
		}
	}
}

func generate8digitUUID() string {
	return strings.ReplaceAll(uuid.New().String(), "-", "")[:8]
}
