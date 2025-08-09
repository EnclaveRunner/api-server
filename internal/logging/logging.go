package logging

import (
	"os"

	"github.com/sirupsen/logrus"
)

// NewLogger returns a configured logrus Logger with pretty printing
func NewLogger() *logrus.Logger {
	logger := logrus.New()
	logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
		ForceColors:   true,
	})
	logger.SetOutput(os.Stdout)

	return logger
}
