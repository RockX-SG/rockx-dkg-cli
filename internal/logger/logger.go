/*
 * ==================================================================
 *Copyright (C) 2022-2023 Altstake Technology Pte. Ltd. (RockX)
 *This file is part of rockx-dkg-cli <https://github.com/RockX-SG/rockx-dkg-cli>
 *CAUTION: THESE CODES HAVE NOT BEEN AUDITED
 *
 *rockx-dkg-cli is free software: you can redistribute it and/or modify
 *it under the terms of the GNU General Public License as published by
 *the Free Software Foundation, either version 3 of the License, or
 *(at your option) any later version.
 *
 *rockx-dkg-cli is distributed in the hope that it will be useful,
 *but WITHOUT ANY WARRANTY; without even the implied warranty of
 *MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
 *GNU General Public License for more details.
 *
 *You should have received a copy of the GNU General Public License
 *along with rockx-dkg-cli. If not, see <http://www.gnu.org/licenses/>.
 *==================================================================
 */

package logger

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rifflock/lfshook"
	"github.com/sirupsen/logrus"
)

func New(serviceName string) *logrus.Logger {

	basePath := "."
	if os.Getenv("DKG_LOG_PATH") != "" {
		basePath = os.Getenv("DKG_LOG_PATH")
	}
	logFilePath := fmt.Sprintf("%s/rockx_dkg_%s.log", basePath, serviceName)

	logger := logrus.New()

	if os.Getenv("DKG_LOG_LEVEL") == "release" {
		logger.SetLevel(logrus.InfoLevel)
	} else {
		logger.SetLevel(logrus.DebugLevel)
	}

	fileHook := lfshook.NewHook(lfshook.PathMap{
		logrus.InfoLevel:  logFilePath,
		logrus.WarnLevel:  logFilePath,
		logrus.ErrorLevel: logFilePath,
		logrus.DebugLevel: logFilePath,
	}, &logrus.JSONFormatter{})

	logger.AddHook(fileHook)

	if serviceName == "cli" && os.Getenv("DKG_LOG_LEVEL") != "debug" {
		logger.SetOutput(io.Discard)
	} else {
		logger.AddHook(fileHook)
		logger.SetOutput(os.Stdout)
	}

	logger.Infof("writing logs to: %s", logFilePath)
	return logger
}

func GinLogger(logger *logrus.Logger) gin.HandlerFunc {
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
