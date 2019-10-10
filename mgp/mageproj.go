package mgp

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"

	"github.com/nocquidant/mageproj/mgl"
)

// MageProject provides Mage dependent high level targets to reuse as is
type MageProject struct {
	ProjectName string
	GroupName   string
	BuildDir    string
	PackageName string
	LdFlags     string
	DckRegistry string
	DckImage    string
	ArtifactURL string
	GitURL      string
	mglib       *mgl.MageLibrary
}

// InitMageProject initializes MageProject
func InitMageProject(workdir string, proj *MageProject) *MageProject {
	// We want to use Go 1.11 modules even if the source lives inside GOPATH.
	// The default is "auto".
	os.Setenv("GO111MODULE", "on")

	proj.mglib = mgl.NewMageLibrary(workdir)

	return proj
}

// MageLibrary gets Mage library used by this project
func (p *MageProject) MageLibrary() *mgl.MageLibrary {
	return p.mglib
}

func (p *MageProject) linkFlags() string {
	return p.LdFlags
}

func (p *MageProject) envFlags() (map[string]string, error) {
	version := p.mglib.Version()

	return map[string]string{
		"PACKAGE":    p.PackageName,
		"VERSION":    version,
		"BUILD_DATE": time.Now().Format("2006-01-02T15:04:05Z0700"),
	}, nil
}

func (p *MageProject) testGoFlags() string {
	return ""
}

func (p *MageProject) buildTags() string {
	return "none"
}

// Validate runs go format and linters
func (p *MageProject) Validate() error {
	mg.Deps(p.mglib.InstallDeps)
	mg.Deps(p.mglib.Format, p.mglib.Vet)

	fmt.Println("===== validate")
	return nil
}

// Test runs tests with go test
func (p *MageProject) Test() error {
	fmt.Println("===== test")

	env := map[string]string{"GOFLAGS": p.testGoFlags()}
	return sh.RunWith(env, mg.GoCmd(), "test", "./...", "-tags", p.buildTags())
}

// Build builds binary in build dir
func (p *MageProject) Build() error {
	mg.Deps(p.Validate)
	mg.Deps(p.Test)

	fmt.Println("===== build")

	fmt.Printf("Building for current OS and architecture \n")
	_, err := p.buildSpecific(target{})

	return err
}

// Package packages cross platform binaries in build dir
func (p *MageProject) Package() error {
	version := p.mglib.Version()
	targets := []target{
		{"windows", "amd64"},
		{"darwin", "amd64"},
		{"linux", "amd64"},
	}
	for _, t := range targets {
		fmt.Printf("Building for OS %s and architecture %s\n", t.goos, t.goarch)

		exe, err := p.buildSpecific(t)
		if err != nil {
			return err
		}

		files := []string{
			exe,
			"README.md",
		}

		archiveName := fmt.Sprintf("%s_%s_%s-%s.zip", p.ProjectName, version, t.goos, t.goarch)
		p.mglib.ZipFiles(filepath.Join(p.BuildDir, archiveName), files)

		os.Remove(exe)
	}
	return nil
}

type target struct {
	goos   string
	goarch string
}

func (p *MageProject) buildSpecific(t target) (string, error) {
	envFlags, err := p.envFlags()
	if err != nil {
		return "", err
	}

	if t.goos != "" && t.goarch != "" {
		envFlags["GOOS"] = t.goos
		envFlags["GOARCH"] = t.goarch
	}

	exe := filepath.Join(p.BuildDir, p.ProjectName)
	if t.goos == "windows" {
		exe += ".exe"
	}

	err = sh.RunWith(envFlags, mg.GoCmd(), "build", "-o", exe, "-ldflags="+p.linkFlags())
	return exe, err
}

// Clean removes the build directory
func (p *MageProject) Clean() error {
	fmt.Println("===== clean")

	return sh.Rm(p.BuildDir)
}

// DockerBuild builds Docker image
func (p *MageProject) DockerBuild() error {
	fmt.Println("===== docker image")

	dck := p.mglib.DockerDetails(p.DckRegistry, p.DckImage, "")

	docker := sh.RunCmd("docker")
	if err := docker("build", "-t", dck.Image, "."); err != nil {
		return err
	}
	return nil
}

// DockerPush pushes Docker image to artifacts repository
func (p *MageProject) DockerPush() error {
	fmt.Println("===== docker release")

	git, _ := p.mglib.GitDetails()
	if git.TagAtRev == "" {
		return errors.New("A git tag is needed to push Docker image")
	}

	dck := p.mglib.DockerDetails(p.DckRegistry, p.DckImage, "jenkins_nexus")

	docker := sh.RunCmd("docker")
	if err := docker("login", dck.Registry, "-u", dck.Usr, "-p", dck.Pwd); err != nil {
		return err
	}
	if err := docker("tag", dck.Image+":latest", dck.Image+":"+git.TagAtRev); err != nil {
		return err
	}
	if err := docker("push", dck.Image+":"+git.TagAtRev); err != nil {
		return err
	}

	if git.Rev == git.RevAtLatestTag {
		if err := docker("push", dck.Image+":latest"); err != nil {
			return err
		}
	}

	return nil
}

// PrintInfo prints information used internally
func (p *MageProject) PrintInfo() {
	var sb strings.Builder
	sb.WriteString("{")
	sb.WriteString(`"workdir": "` + p.mglib.Workdir() + `", `)

	git, _ := p.mglib.GitDetails()
	sb.WriteString(`"git": {`)
	sb.WriteString(`"rev": "` + git.Rev + `", `)
	sb.WriteString(`"tagAtRev": "` + git.TagAtRev + `", `)
	sb.WriteString(`"latestTag": "` + git.LatestTag + `", `)
	sb.WriteString(`"revAtLatestTag": "` + git.RevAtLatestTag + `"`)
	sb.WriteString("}, ")

	dck := p.mglib.DockerDetails(p.DckRegistry, p.DckImage, "")
	sb.WriteString(`"docker": {`)
	sb.WriteString(`"registry": "` + dck.Registry + `", `)
	sb.WriteString(`"image": "` + dck.Image + `"`)
	sb.WriteString("}")

	sb.WriteString("}")

	fmt.Println(sb.String())
}
