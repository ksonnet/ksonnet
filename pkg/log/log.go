package log

import (
	"io"

	"github.com/sirupsen/logrus"
)

var (
	// VerbosityLevel is the current verbosity level.
	VerbosityLevel = 0
)

// Init initializes ksonnet's logger.
func Init(verbosity int, w io.Writer) {
	logrus.SetOutput(w)
	logrus.SetFormatter(defaultLogFmt())
	logrus.SetLevel(logLevel(verbosity))
	VerbosityLevel = verbosity

	logrus.WithField("verbosity-level", verbosity).Debug("setting log verbosity")
}

func defaultLogFmt() logrus.Formatter {
	return &logrus.TextFormatter{
		DisableTimestamp:       true,
		DisableLevelTruncation: true,
		QuoteEmptyFields:       true,
	}
}

func logLevel(verbosity int) logrus.Level {
	switch verbosity {
	case 0:
		return logrus.InfoLevel
	default:
		return logrus.DebugLevel
	}
}
