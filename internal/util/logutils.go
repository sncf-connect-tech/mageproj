package util

import (
	"log"
	"os"
)

// by default, the log stdlib package will be set to discard output.
// Running with mage -v will set the output to stdout.
var logger = log.New(os.Stderr, "", 0)

// LogIfVerbose logs message if verbose flag is on.
func LogIfVerbose(msg string, v ...interface{}) {
	if Verbose() {
		logger.Println(msg, v)
	}
}
