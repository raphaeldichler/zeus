// Copyright 2025 The Zeus Authors.
// Licensed under the Apache License 2.0. See the LICENSE file for details.

package log

import (
	"fmt"
	"io"
	stdlog "log"
	"os"
	"time"

	"github.com/raphaeldichler/zeus/internal/util/assert"
)

var (
	stdLogger *stdlog.Logger
)

func init() {
	var loggerOutput io.Writer = os.Stdout
	if os.Getenv("DISCARD_LOGS") == "1" {
		loggerOutput = io.Discard
	}

	stdLogger = stdlog.New(loggerOutput, "", 0)
}

type Logger struct {
	application string
	daemon      string
}

func New(application string, daemon string) *Logger {
	assert.NotNil(stdLogger, "invalid setup of the logger")

	return &Logger{
		application: application,
		daemon:      daemon,
	}
}

func (l *Logger) Debug(format string, args ...any) {
	stdLogger.Print(l.format("DEBUG", format, args...))
}

func (l *Logger) Info(format string, args ...any) {
	stdLogger.Print(l.format("INFO", format, args...))
}

func (l *Logger) Error(format string, args ...any) {
	stdLogger.Print(l.format("ERROR", format, args...))
}

func (l *Logger) format(level string, format string, args ...any) string {
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	prefix := fmt.Sprintf("[%s] [%s] [%s] [%s] ", timestamp, l.application, l.daemon, level)
	return prefix + fmt.Sprintf(format, args...)
}
