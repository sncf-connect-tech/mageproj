package util

import (
	"log"
	"os"
)

// by default, the log stdlib package will be set to discard output.
// Running with mage -v will set the output to stdout.
var logger = log.New(os.Stdout, "", 0)

// LogIfVerbose logs message if verbose flag is on.
func LogIfVerbose(msg string, v ...interface{}) {
	if Verbose() {
		logger.Println(msg, v)
	}
}

// LogIfDebug logs message if debug flag is on.
func LogIfDebug(msg string, v ...interface{}) {
	if Debug() {
		logger.Println(msg, v)
	}
}
