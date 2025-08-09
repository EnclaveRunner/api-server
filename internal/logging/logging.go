package logging

import (
	"os"

	"github.com/sirupsen/logrus"
)

// NewLogger returns a configured logrus Logger with pretty printing
func NewLogger() *logrus.Logger {
	logger := logrus.New()
	logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:             true,
		ForceColors:               true,
		DisableColors:             false,
		ForceQuote:                false,
		DisableQuote:              false,
		EnvironmentOverrideColors: false,
		DisableTimestamp:          false,
		TimestampFormat:           "",
		DisableSorting:            false,
		SortingFunc:               nil,
		DisableLevelTruncation:    false,
		PadLevelText:              false,
		QuoteEmptyFields:          false,
		FieldMap:                  nil,
		CallerPrettyfier:          nil,
	})
	logger.SetOutput(os.Stdout)

	return logger
}
