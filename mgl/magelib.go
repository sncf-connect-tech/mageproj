package mgl

import (
	"archive/zip"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"
)

// MageLibrary provides Mage independent functions to build its own targets
type MageLibrary struct {
	workdir string
	pkgs    *PackageInfos
	git     *GitInfos
	dck     *DockerInfos
}

// NewMageLibrary constructs new MageLibrary instance
func NewMageLibrary(workdir string) *MageLibrary {
	commons := MageLibrary{}
	commons.workdir = workdir
	commons.pkgs = &PackageInfos{}
	commons.git = &GitInfos{}
	commons.dck = &DockerInfos{}
	return &commons
}

// Workdir returns workdir used
func (c *MageLibrary) Workdir() string {
	return c.workdir
}

// Version extracts version from git tag
func (c *MageLibrary) Version() string {
	timestamp := time.Now().UnixNano() / int64(time.Millisecond)
	version := fmt.Sprintf("dev@%d", timestamp)

	git, err := c.GitDetails()
	if err == nil {
		version = fmt.Sprintf("dev@%s", git.Rev)
		if git.TagAtRev != "" {
			version = git.TagAtRev
		}
	}

	return version
}

func execOutput(cmd string, args ...string) (string, error) {
	out, err := exec.Command(cmd, args...).Output()
	if err != nil {
		return "", err
	}
	if len(out) > 0 {
		return string(out), nil
	}
	return "", nil
}

// PackageInfos holds information regarding the go project
type PackageInfos struct {
	prefixLen int
	Names     []string
	init      sync.Once
}

func goCmd() string {
	if cmd := os.Getenv("MAGEFILE_GOCMD"); cmd != "" {
		return cmd
	}
	return "go"
}

// PackageDetails aggregates the package information regarding the go project
func (c *MageLibrary) PackageDetails() (*PackageInfos, error) {
	var err error
	c.pkgs.init.Do(func() {
		var s string
		s, err = execOutput(goCmd(), "list", "./...")
		if err != nil {
			return
		}
		names := strings.Split(s, "\n")
		for i := range names {
			names[i] = "." + names[i][c.pkgs.prefixLen:]
		}
		c.pkgs.Names = names
	})

	return c.pkgs, err
}

// GitInfos holds information regarding git
type GitInfos struct {
	Rev            string
	TagAtRev       string
	LatestTag      string
	RevAtLatestTag string
	init           sync.Once
}

func trimString(val string) string {
	val = strings.TrimSpace(val)
	val = strings.TrimPrefix(val, "\"")
	val = strings.TrimSuffix(val, "\n")
	val = strings.TrimSuffix(val, "\"")
	return val
}

// GitDetails aggregates the information regarding git
func (c *MageLibrary) GitDetails() (*GitInfos, error) {
	var err error
	c.git.init.Do(func() {
		c.git.Rev, _ = execOutput("git", "rev-parse", "--short", "HEAD")
		c.git.Rev = trimString(c.git.Rev)

		c.git.TagAtRev, _ = execOutput("git", "tag", fmt.Sprintf("--points-at=%s", c.git.Rev))
		c.git.TagAtRev = trimString(c.git.TagAtRev)

		c.git.LatestTag, _ = execOutput("git", "for-each-ref", "--format=\"%(tag)\"", "--sort=-taggerdate", "refs/tags")
		if c.git.LatestTag != "" {
			allTags := strings.Split(c.git.LatestTag, "\n")
			if len(allTags) > 0 {
				c.git.LatestTag = trimString(allTags[0])
			}
		}
		c.git.LatestTag = trimString(c.git.LatestTag)

		c.git.RevAtLatestTag, _ = execOutput("git", "rev-list", "--abbrev-commit", "-n", "1", fmt.Sprintf("%s", c.git.LatestTag))
		c.git.RevAtLatestTag = trimString(c.git.RevAtLatestTag)
	})
	return c.git, err
}

// DockerInfos holds information regarding docker
type DockerInfos struct {
	Registry string
	Image    string
	Usr      string
	Pwd      string
	init     sync.Once
}

// DockerDetails aggregates the information regarding docker
func (c *MageLibrary) DockerDetails(registry, image, user string) *DockerInfos {
	c.dck.init.Do(func() {
		c.dck.Registry = registry
		c.dck.Image = image
		c.dck.Usr = user
		c.dck.Pwd = "to.be.set"
		if usr := os.Getenv("DOCKER_USR"); usr != "" {
			c.dck.Usr = usr
		}
		if pwd := os.Getenv("DOCKER_PWD"); pwd != "" {
			c.dck.Pwd = pwd
		}
	})
	return c.dck
}

// Format formats code via gofmt
func (c *MageLibrary) Format() error {
	pkgs, err := c.PackageDetails()
	if err != nil {
		return err
	}

	failed := false
	first := true
	for _, pkg := range pkgs.Names {
		files, err := filepath.Glob(filepath.Join(pkg, "*.go"))
		if err != nil {
			return nil
		}
		for _, f := range files {
			// gofmt doesn't exit with non-zero when it finds unformatted code
			// so we have to explicitly look for output, and if we find any, we
			// should fail this target.
			s, err := execOutput("gofmt", "-l", f)
			if err != nil {
				fmt.Printf("ERROR: running gofmt on %q: %v\n", f, err)
				failed = true
			}
			if s != "" {
				if first {
					fmt.Println("The following files are not gofmt'ed:")
					first = false
				}
				failed = true
				fmt.Println(s)
			}
		}
	}
	if failed {
		return errors.New("improperly formatted go files")
	}
	return nil
}

// Lint runs golint linter
func (c *MageLibrary) Lint() error {
	pkgs, err := c.PackageDetails()
	if err != nil {
		return err
	}

	failed := false
	for _, pkg := range pkgs.Names {
		// We don't actually want to fail this target if we find golint errors,
		// so we don't pass -set_exit_status, but we still print out any failures.
		if err := exec.Command("golint", pkg).Run(); err != nil {
			fmt.Printf("ERROR: running go lint on %q: %v\n", pkg, err)
			failed = true
		}
	}
	if failed {
		return errors.New("Errors running golint")
	}
	return nil
}

// Vet runs go vet linter
func (c *MageLibrary) Vet() error {
	if err := exec.Command(goCmd(), "vet", "./...").Run(); err != nil {
		return fmt.Errorf("error running go vet: %v", err)
	}
	return nil
}

// InstallDeps installs the additional dependencies: goimports & golint
func (c *MageLibrary) InstallDeps() error {
	err := exec.Command(goCmd(), "get", "golang.org/x/lint/golint").Run()
	if err == nil {
		return err
	}
	return exec.Command(goCmd(), "get", "golang.org/x/tools/cmd/goimports").Run()
}

// ZipFiles compresses one or many files into a single zip archive file.
// The original code was published under MIT licence under https://golangcode.com/create-zip-files-in-go/
func (c *MageLibrary) ZipFiles(filename string, files []string) error {
	newfile, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer newfile.Close()

	zipWriter := zip.NewWriter(newfile)
	defer zipWriter.Close()

	// Add files to zip
	for _, file := range files {

		zipfile, err := os.Open(file)
		if err != nil {
			return err
		}
		defer zipfile.Close()

		// Get the file information
		info, err := zipfile.Stat()
		if err != nil {
			return err
		}

		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}

		// Change to deflate to gain better compression
		// see http://golang.org/pkg/archive/zip/#pkg-constants
		header.Method = zip.Deflate

		writer, err := zipWriter.CreateHeader(header)
		if err != nil {
			return err
		}
		_, err = io.Copy(writer, zipfile)
		if err != nil {
			return err
		}
	}

	return nil
}

// ChangeLog generates a ChangeLog based on git history
func (c *MageLibrary) ChangeLog(filename string, artifactURL, gitURL string) error {
	newfile, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer newfile.Close()

	out, err := execOutput("git", "log", "--pretty=\"tformat:%d|%ci|%h|%cn|%s\"")
	if err != nil {
		return nil
	}

	tag := regexp.MustCompile(`^.*tag: v([0-9]+\.[0-9]+\.[0-9]+).*$`)

	lines := strings.Split(out, "\n")
	for _, line := range lines {
		tokens := convertToTokens(line)
		if tokens == nil {
			continue
		}
		var outs string
		if tag.MatchString(tokens.refNames) {
			match := tag.FindStringSubmatch(tokens.refNames)
			outs = fmt.Sprintf("\n## Release v[%s](%s/%s) (%s)\n\n", match[1], artifactURL, match[1], tokens.committerDate)
		} else {
			outs = fmt.Sprintf("* [%s](%s/commit/%s) - %s (%s)\n", tokens.commitHash, gitURL, tokens.commitHash, tokens.subject, tokens.committerName)
		}
		if _, err = newfile.WriteString(outs); err != nil {
			return err
		}
	}

	return nil
}

type historyTokens struct {
	refNames      string
	committerDate string
	commitHash    string
	committerName string
	subject       string
}

func convertToTokens(line string) *historyTokens {
	tokens := strings.Split(line, "|")
	if len(tokens) < 5 {
		return nil
	}
	history := historyTokens{}
	history.refNames = trimString(tokens[0])
	history.committerDate = trimString(tokens[1])
	history.commitHash = trimString(tokens[2])
	history.committerName = trimString(tokens[3])
	history.subject = trimString(tokens[4])
	return &history
}
