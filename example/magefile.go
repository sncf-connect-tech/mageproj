//go:build mage

package main

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/voyages-sncf-technologies/mageproj/mgp"
)

const (
	projectName = "myapp"
	groupName   = "mygroup"
	buildDir    = "build"
	ldFlags     = `-X "main.Version=$VERSION" -X "main.BuildDate=$BUILD_DATE"`
	artifact    = "artifactory.mycompany.fr"
	gitRepo     = "gitlab.mycompany.fr"
)

var proj *mgp.MageProject

func init() {
	packageName := filepath.Join(gitRepo, groupName, projectName)
	artifactURL := "https://" + filepath.Join(artifact, "repository", projectName)
	gitURL := "https://" + filepath.Join(gitRepo, groupName, projectName)

	withLdFlags := mgp.WithCompileFlags(ldFlags)
	withArtURL := mgp.WithArtifactURL(artifactURL)
	withGitURL := mgp.WithGitURL(gitURL)

	mgp.Logf(">> Using packageName %s\n", packageName)
	mgp.Logf(">> Using artifactURL %s\n", artifactURL)
	mgp.Logf(">> Using gitURL %s\n", gitURL)

	proj = mgp.NewMageProject(currentDir(), projectName, packageName,
		withLdFlags, withArtURL, withGitURL)
}

func currentDir() string {
	workdir, err := os.Getwd()
	if err != nil {
		workdir = "."
	}
	return workdir
}

// Validate runs go format and linters
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

// Deploy deploys x-platform binaries to artifact registry
func Deploy() error {
	_, present := os.LookupEnv("MAGEFILEP_ARTIFACT_USR")
	if !present {
		os.Setenv("MAGEFILEP_ARTIFACT_USR", "myuser")
	}
	_, present = os.LookupEnv("MAGEFILEP_ARTIFACT_PWD")
	if !present {
		return errors.New("missing password for Artifactory (set variable MAGEFILEP_ARTIFACT_PWD)")
	}

	return proj.Deploy()
}

// Clean removes the build directory
func Clean() error {
	return proj.Clean()
}

// ChangeLog generates a ChangeLog based on git history
func ChangeLog() error {
	return proj.ChangeLog()
}
