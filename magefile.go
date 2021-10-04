//go:build mage

package main

import (
	"os"
	"path/filepath"

	"github.com/voyages-sncf-technologies/mageproj/v2/mgp"
)

const (
	projectName  = "mageproj"
	groupName    = "voyages-sncf-technologies"
	artifactRepo = "github.com"
	gitRepo      = "github.com"
)

var proj *mgp.MageProject

func init() {
	packageName := filepath.Join(gitRepo, groupName, projectName)
	artifactURL := "https://" + filepath.Join(artifactRepo, groupName, projectName, "releases")
	gitURL := "https://" + filepath.Join(gitRepo, groupName, projectName)

	withArtURL := mgp.WithArtifactURL(artifactURL)
	withGitURL := mgp.WithGitURL(gitURL)

	proj = mgp.NewMageProject(currentDir(), projectName, packageName,
		withArtURL, withGitURL)
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

// Package packages x-platform binaries in build dir
func Package() error {
	return proj.Package()
}

// Release creates a git tag and push it to remote
func Release() error {
	return proj.Release()
}

// ChangeLog generates a ChangeLog based on git history
func ChangeLog() error {
	return proj.ChangeLog()
}
