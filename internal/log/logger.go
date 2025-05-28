// Copyright 2025 The Zeus Authors.
// Licensed under the MIT License. See the LICENSE file for details.

package log

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/raphaeldichler/zeus/internal/assert"
	"github.com/sirupsen/logrus"
)

var (
	logger *logrus.Logger
)

func init() {
	var loggerOutput io.Writer = os.Stdout
	if os.Getenv("DISCARD_LOGS") == "1" {
		loggerOutput = io.Discard
	}

	logger = logrus.New()
	logger.SetOutput(loggerOutput)
	logger.SetFormatter(&logrus.TextFormatter{
		DisableTimestamp: true,
		DisableColors:    false,
		DisableQuote:     true,
	})
}

type Logger struct {
	application string
	daemon      string
}

func New(
	application string,
	daemon string,
) *Logger {
	assert.NotNil(logger, "invalid setup of the logger")

	return &Logger{
		application: application,
		daemon:      daemon,
	}
}

func (l *Logger) Debug(format string, args ...any) {
	logger.Debugf(l.prefix()+format, args...)
}

// Info logs an info-level message with the custom format.
func (l *Logger) Info(format string, args ...any) {
	logger.Infof(l.prefix()+format, args...)
}

// Error logs an error-level message with the custom format.
func (l *Logger) Error(format string, args ...any) {
	logger.Errorf(l.prefix()+format, args...)
}

// prefix returns the "[time] [app] [daemon] " prefix.
func (l *Logger) prefix() string {
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	return fmt.Sprintf("[%s] [%s] [%s] ", timestamp, l.application, l.daemon)
}
