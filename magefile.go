// +build mage

package main

import (
	"os"

	"github.com/voyages-sncf-technologies/mageproj/mgp"
)

const (
	projectName = "mageproj"
	packageName = "github.com/nocquidant/magedir"
)

var proj *mgp.MageProject

func init() {
	proj = mgp.NewMageProject(currentDir(), projectName, packageName)
}

func currentDir() string {
	workdir, err := os.Getwd()
	if err != nil {
		workdir = "."
	}
	return workdir
}

// Clean removes the build directory
func Clean() error {
	return proj.Clean()
}

// Validate runs format and linters
func Validate() error {
	return proj.Validate()
}

// Test runs tests with go test
func Test() error {
	return proj.Test()
}

// Build builds binary in build dir
func Build() error {
	return proj.Build()
}

// ChangeLog generates a ChangeLog based on git history
func ChangeLog() error {
	return proj.ChangeLog()
}
