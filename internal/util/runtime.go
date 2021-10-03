package util

import (
	"os"
	"os/exec"
	"strconv"
)

// RunCmd runs the given command displaying its standard output if in verbose mode
func RunCmd(name string, arg ...string) error {
	c := exec.Command(name, arg...)

	out, err := c.CombinedOutput()
	if Verbose() && out != nil && len(out) > 0 {
		AlwaysLog(string(out))
	}
	return err
}

// Verbose reports whether a magefile was run with the verbose flag.
func Verbose() bool {
	val, present := os.LookupEnv("MAGEFILE_VERBOSE")
	if present {
		b, _ := strconv.ParseBool(val)
		return b
	}
	val, present = os.LookupEnv("MAGEP_VERBOSE")
	if present {
		b, _ := strconv.ParseBool(val)
		return b
	}
	return false
}

// GoCmd reports the command to use to build go code. By default it is
// the "go" binary in the PATH.
func GoCmd() string {
	val, present := os.LookupEnv("MAGEFILE_GOCMD")
	if present {
		return val
	}
	val, present = os.LookupEnv("MAGEP_GOCMD")
	if present {
		return val
	}
	return "go"
}

// GitCmd reports the command to use to extract git info. By default it is
// the "go" binary in the PATH.
func GitCmd() string {
	val, present := os.LookupEnv("MAGEP_GITCMD")
	if present {
		return val
	}
	return "git"
}
