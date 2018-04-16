package fs

import (
	"github.com/sirupsen/logrus"
)

// Log is the logrus instance, exported to be able to log throughout the packages.
var Log = logrus.New()

func init() {
	Log.SetLevel(logrus.InfoLevel)
}

// Debug will set the logging level based on the argument.
// A true value will set the level to debug, otherwise info.
func Debug(b bool) {
	if b {
		Log.SetLevel(logrus.DebugLevel)
	} else {
		Log.SetLevel(logrus.InfoLevel)
	}
}
