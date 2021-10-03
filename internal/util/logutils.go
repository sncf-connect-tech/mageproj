package util

import (
	"log"
	"os"
)

// by default, the log stdlib package will be set to discard output.
// Running with mage -v will set the output to stdout.
var logger = log.New(os.Stderr, "", 0)

// AlwaysLogf logs message either in verbose mode or not.
func AlwaysLogf(msg string, v ...interface{}) {
	logger.Printf(msg, v...)
}

// AlwaysLog logs message either in verbose mode or not.
func AlwaysLog(msg string) {
	logger.Println(msg)
}

// Logf logs message if verbose mode is on.
func Logf(msg string, v ...interface{}) {
	if Verbose() {
		logger.Printf(msg, v...)
	}
}

// Log logs message if verbose mode is on.
func Log(msg string) {
	if Verbose() {
		logger.Println(msg)
	}
}
