package mgl

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/voyages-sncf-technologies/mageproj/v2/internal/util"
)

// MageLibrary provides Mage independent functions to build its own targets
type MageLibrary struct {
	workdir string
	pkgs    *PackageInfos
	git     *GitInfos
	art     *ArtifactInfos
	dck     *DockerInfos
}

// MageLibraryOption defines an operation on MageLibrary (to set a param)
type MageLibraryOption func(*MageLibrary)

// GitInfos holds information regarding git
type GitInfos struct {
	Rev            string
	TagAtRev       string
	LatestTag      string
	RevAtLatestTag string
	init           sync.Once
}

// PackageInfos holds information regarding the go project
type PackageInfos struct {
	prefixLen int
	Names     []string
	init      sync.Once
}

// ArtifactInfos holds information regarding artifacts registry
type ArtifactInfos struct {
	URL  string
	Usr  string
	Pwd  string
	init sync.Once
}

// DockerInfos holds information regarding docker
type DockerInfos struct {
	Registry string
	Image    string
	Usr      string
	Pwd      string
	init     sync.Once
}

// historyTokens converts git log history to tokens
type historyTokens struct {
	refNames      string
	committerDate string
	commitHash    string
	committerName string
	subject       string
}

// NewMageLibrary constructs a new MageLibrary instance
func NewMageLibrary(workdir string, options ...MageLibraryOption) *MageLibrary {
	commons := &MageLibrary{}
	commons.workdir = workdir
	commons.pkgs = &PackageInfos{}
	commons.git = &GitInfos{}
	commons.art = &ArtifactInfos{}
	commons.dck = &DockerInfos{}
	for _, option := range options {
		option(commons)
	}
	return commons
}

// Workdir returns the workdir used
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

// PackageDetails aggregates the package information regarding the go project
func (c *MageLibrary) PackageDetails() (*PackageInfos, error) {
	var err error
	c.pkgs.init.Do(func() {
		var s string
		s, err = util.ExecOutput(util.GoCmd(), "list", "./...")
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

// GitDetails aggregates the information regarding git
func (c *MageLibrary) GitDetails() (*GitInfos, error) {
	var err error
	c.git.init.Do(func() {
		gitDir := filepath.Join(c.Workdir(), ".git")

		c.git.Rev, _ = util.ExecOutput(util.GitCmd(), "--git-dir", gitDir, "rev-parse", "--short", "HEAD")
		c.git.Rev = util.TrimString(c.git.Rev)

		c.git.TagAtRev, _ = util.ExecOutput(util.GitCmd(), "--git-dir", gitDir, "tag", fmt.Sprintf("--points-at=%s", c.git.Rev))
		c.git.TagAtRev = util.TrimString(c.git.TagAtRev)

		c.git.LatestTag, _ = util.ExecOutput(util.GitCmd(), "--git-dir", gitDir, "for-each-ref", "--format=\"%(tag)\"", "--sort=-taggerdate", "refs/tags")
		if c.git.LatestTag != "" {
			allTags := strings.Split(c.git.LatestTag, "\n")
			if len(allTags) > 0 {
				c.git.LatestTag = util.TrimString(allTags[0])
			}
		}
		c.git.LatestTag = util.TrimString(c.git.LatestTag)

		c.git.RevAtLatestTag, _ = util.ExecOutput(util.GitCmd(), "--git-dir", gitDir, "rev-list", "--abbrev-commit", "-n", "1", c.git.LatestTag)
		c.git.RevAtLatestTag = util.TrimString(c.git.RevAtLatestTag)
	})
	return c.git, err
}

// ArtifactDetails aggregates the information regarding artifacts registry
func (c *MageLibrary) ArtifactDetails(url, user string) *ArtifactInfos {
	c.art.init.Do(func() {
		c.art.URL = url
		c.art.Usr = user
		c.art.Pwd = "to.be.set"
	})
	if usr := os.Getenv("MAGEFILEP_ARTIFACT_USR"); usr != "" {
		c.art.Usr = usr
	}
	if pwd := os.Getenv("MAGEFILEP_ARTIFACT_PWD"); pwd != "" {
		c.art.Pwd = pwd
	}
	return c.art
}

// DockerDetails aggregates the information regarding docker
func (c *MageLibrary) DockerDetails(registry, image, user string) *DockerInfos {
	c.dck.init.Do(func() {
		c.dck.Registry = registry
		c.dck.Image = image
		c.dck.Usr = user
		c.dck.Pwd = "to.be.set"
	})
	if usr := os.Getenv("MAGEFILEP_DOCKER_USR"); usr != "" {
		c.dck.Usr = usr
	}
	if pwd := os.Getenv("MAGEFILEP_DOCKER_PWD"); pwd != "" {
		c.dck.Pwd = pwd
	}
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
			s, err := util.ExecOutput("gofmt", "-l", f)
			if err != nil {
				fmt.Printf("ERROR: running gofmt on %q: %v\n", f, err)
				failed = true
			}
			if s != "" {
				if first {
					util.AlwaysLog("The following files are not gofmt'ed:")
					first = false
				}
				failed = true
				util.AlwaysLog(s)
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
		return errors.New("errors running golint")
	}
	return nil
}

// Vet runs go vet linter
func (c *MageLibrary) Vet() error {
	if err := util.RunCmd(util.GoCmd(), "vet", "./..."); err != nil {
		return fmt.Errorf("error running go vet: %v", err)
	}
	return nil
}

// InstallDeps installs the additional dependencies: goimports & golint
func (c *MageLibrary) InstallDeps() error {
	err := util.RunCmd(util.GoCmd(), "install", "golang.org/x/lint/golint@latest")
	if err == nil {
		return err
	}
	err = util.RunCmd(util.GoCmd(), "install", "golang.org/x/tools/cmd/goimports@latest")
	return err
}

// ChangeLog generates a ChangeLog based on git history
func (c *MageLibrary) ChangeLog(version, filename string, artifactURL, gitURL string) error {
	newfile, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer newfile.Close()

	t := time.Now()
	ts := t.Format("2006-01-02 15:04:05 -0700")
	fl := fmt.Sprintf("\n## Release [%s](%s/%s) (%s)\n\n", version, artifactURL, version, ts)
	if _, err = newfile.WriteString(fl); err != nil {
		return err
	}

	gitDir := filepath.Join(c.Workdir(), ".git")
	out, err := util.ExecOutput(util.GitCmd(), "--git-dir", gitDir, "log", "--pretty=\"tformat:%d|%ci|%h|%cn|%s\"")
	if err != nil {
		return nil
	}

	tag := regexp.MustCompile(`^.*tag: (v[0-9]+\.[0-9]+\.[0-9]+).*$`)

	lines := strings.Split(out, "\n")
	for _, line := range lines {
		tokens := convertToTokens(line)
		if tokens == nil {
			continue
		}
		var outs string
		if tag.MatchString(tokens.refNames) {
			match := tag.FindStringSubmatch(tokens.refNames)
			outs = fmt.Sprintf("\n## Release [%s](%s/%s) (%s)\n\n", match[1], artifactURL, match[1], tokens.committerDate)

			if _, err = newfile.WriteString(outs); err != nil {
				return err
			}
		}

		if keep := os.Getenv("MAGEFILEP_MERGE_COMMIT"); keep != "yes" {
			if strings.Contains(tokens.subject, "Merge branch") {
				if strings.Contains(tokens.subject, "into") {
					continue
				}
			}
		}

		outs = fmt.Sprintf("* [%s](%s/commit/%s) - %s (%s)\n", tokens.commitHash, gitURL, tokens.commitHash, tokens.subject, tokens.committerName)

		if _, err = newfile.WriteString(outs); err != nil {
			return err
		}
	}

	util.AlwaysLogf("File %s generated", newfile.Name())
	return nil
}

func convertToTokens(line string) *historyTokens {
	tokens := strings.Split(line, "|")
	if len(tokens) < 5 {
		return nil
	}
	history := historyTokens{}
	history.refNames = util.TrimString(tokens[0])
	history.committerDate = util.TrimString(tokens[1])
	history.commitHash = util.TrimString(tokens[2])
	history.committerName = util.TrimString(tokens[3])
	history.subject = util.TrimString(tokens[4])
	return &history
}
