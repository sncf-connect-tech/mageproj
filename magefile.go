// +build mage

package main

import (
	"os"

	"github.com/nocquidant/mageproj/mgp"
)

var proj *mgp.MageProject

func init() {
	proj = &mgp.MageProject{
		ProjectName: "mageproj",
		BuildDir:    "build",
		PackageName: "github.com/nocquidant/magedir",
	}
	proj = mgp.InitMageProject(currentDir(), proj)
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
