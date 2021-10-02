package mgp

import (
	"bufio"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"

	"github.com/voyages-sncf-technologies/mageproj/internal/util"
	"github.com/voyages-sncf-technologies/mageproj/mgl"
)

// MageProjectOption defines an operation which set an option
type MageProjectOption func(*MageProject)

// MageProject provides Mage dependent high level targets to reuse as is
type MageProject struct {
	projectName string
	groupName   string
	buildDir    string
	packageName string
	ldFlags     string
	testFlags   string
	dckRegistry string
	dckImage    string
	dckAppPath  string
	artifactURL string
	gitURL      string
	mglib       *mgl.MageLibrary
}

type target struct {
	goos   string
	goarch string
}

// NewMageProject constructs a new MageProject instance
func NewMageProject(workdir, projectName, packageName string, options ...MageProjectOption) *MageProject {
	// We want to use Go 1.11 modules even if the source lives inside GOPATH.
	// The default is "auto".
	os.Setenv("GO111MODULE", "on")

	proj := &MageProject{}
	proj.projectName = projectName
	proj.packageName = packageName
	proj.buildDir = "build"
	proj.dckAppPath = "/app"

	proj.mglib = mgl.NewMageLibrary(workdir)

	for _, option := range options {
		option(proj)
	}
	return proj
}

// WithGroupName sets groupName to value
func WithGroupName(val string) MageProjectOption {
	return func(ml *MageProject) {
		ml.groupName = val
	}
}

// WithBuildDir sets buildDir to value
func WithBuildDir(val string) MageProjectOption {
	return func(ml *MageProject) {
		ml.buildDir = val
	}
}

// WithCompileFlags sets ldFlags to value
func WithCompileFlags(val string) MageProjectOption {
	return func(ml *MageProject) {
		ml.ldFlags = val
	}
}

// WithTestFlags sets testFlags to value
func WithTestFlags(val string) MageProjectOption {
	return func(ml *MageProject) {
		ml.testFlags = val
	}
}

// WithDockerRegistry sets dckRegistry to value
func WithDockerRegistry(val string) MageProjectOption {
	return func(ml *MageProject) {
		ml.dckRegistry = val
	}
}

// WithDockerImage sets dckImage to value
func WithDockerImage(val string) MageProjectOption {
	return func(ml *MageProject) {
		ml.dckImage = val
	}
}

// WithDockerAppPath sets dckAppPath to value
func WithDockerAppPath(val string) MageProjectOption {
	return func(ml *MageProject) {
		ml.dckAppPath = val
	}
}

// WithArtifactURL sets artifactURL to value
func WithArtifactURL(val string) MageProjectOption {
	return func(ml *MageProject) {
		ml.artifactURL = val
	}
}

// WithGitURL sets gitURL to value
func WithGitURL(val string) MageProjectOption {
	return func(ml *MageProject) {
		ml.gitURL = val
	}
}

// MageLibrary gets Mage library used by this project
func (p *MageProject) MageLibrary() *mgl.MageLibrary {
	return p.mglib
}

func (p *MageProject) linkFlags() string {
	return p.ldFlags
}

func (p *MageProject) envFlags() (map[string]string, error) {
	version := p.mglib.Version()

	return map[string]string{
		"PACKAGE":    p.packageName,
		"VERSION":    version,
		"BUILD_DATE": time.Now().Format("2006-01-02T15:04:05Z0700"),
	}, nil
}

func (p *MageProject) testGoFlags() string {
	return p.testFlags
}

func (p *MageProject) buildTags() string {
	return ""
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

	tags := ""
	if t := p.buildTags(); t != "" {
		tags = "-tags" + t
	}

	return sh.RunWith(env, mg.GoCmd(), "test", "./...", tags)
}

func (p *MageProject) dumpInfoToDisk() error {
	p.checkBuildDir()

	info := p.PrintInfo()
	if info != "" {
		f := filepath.Join(p.MageLibrary().Workdir(), p.buildDir, "build-info.json")
		return ioutil.WriteFile(f, []byte(info), 0644)
	}

	return nil
}

// Build builds binary in build dir
func (p *MageProject) Build() error {
	mg.Deps(p.Validate)
	mg.Deps(p.Test)

	fmt.Println("===== build")

	util.LogIfVerbose("Building for current OS and architecture")
	_, err := p.buildSpecific(target{})

	if err == nil {
		err = p.dumpInfoToDisk()
	}

	return err
}

// Package packages cross platform binaries in build dir
func (p *MageProject) Package() error {
	mg.Deps(p.Validate)
	mg.Deps(p.Test)

	fmt.Println("===== package")

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

		if t.goos == "windows" {
			archiveName := fmt.Sprintf("%s_%s_%s-%s.zip", p.projectName, version, t.goos, t.goarch)
			util.ZipFiles(filepath.Join(p.buildDir, archiveName), files)
		} else {
			archiveName := fmt.Sprintf("%s_%s_%s-%s.tar.gz", p.projectName, version, t.goos, t.goarch)
			util.TarFiles(filepath.Join(p.buildDir, archiveName), files, false)
		}

		os.Remove(exe)
	}

	return p.dumpInfoToDisk()
}

// Deploy deploys cross platform binaries to artifacts registry
func (p *MageProject) Deploy() error {
	fmt.Println("===== deploy")

	var files []string

	dir := filepath.Join(p.mglib.Workdir(), p.buildDir)
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if filepath.Ext(path) == ".zip" || strings.HasSuffix(path, ".tar.gz") {
			files = append(files, path)
		}
		return nil
	})

	if err != nil {
		return err
	}

	git, _ := p.mglib.GitDetails()
	if git.TagAtRev == "" {
		return errors.New("a git tag is needed to deploy the binaries to a registry")
	}

	art := p.mglib.ArtifactDetails(p.artifactURL, "")

	httpClient := &http.Client{}
	for _, file := range files {
		url := art.URL + "/" + filepath.Join(git.TagAtRev, filepath.Base(file))
		util.LogIfVerbose("Artifact url: ", url)

		details, err := util.GetFileDetails(file)
		if err != nil {
			return err
		}

		util.LogIfVerbose("Sum SHA256: ", details.Checksum.Sha256)
		util.LogIfVerbose("Sum SHA1: ", details.Checksum.Sha1)
		util.LogIfVerbose("Sum MD5: ", details.Checksum.Md5)

		f, err := os.Open(file)
		if err != nil {
			return err
		}
		defer f.Close()

		req, err := http.NewRequest("PUT", url, bufio.NewReader(f))
		if err != nil {
			return err
		}

		req.SetBasicAuth(art.Usr, art.Pwd)

		req.Header.Set("X-Checksum-SHA256", details.Checksum.Sha256)
		req.Header.Set("X-Checksum-SHA1", details.Checksum.Sha1)
		req.Header.Set("X-Checksum-MD5", details.Checksum.Md5)

		req.ContentLength = details.Size
		req.Close = true

		util.LogIfVerbose("Uploading file: ", file)
		resp, err := httpClient.Do(req)
		if err != nil {
			return err
		}

		util.LogIfVerbose("Received HTTP status code: ", resp.StatusCode)

		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		util.LogIfDebug("Received HTTP body: ", string(body))

		if !(resp.StatusCode >= 200 && resp.StatusCode <= 299) {
			return fmt.Errorf("unsuccessful request code %d", resp.StatusCode)
		}
	}

	return nil
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

	exe := filepath.Join(p.buildDir, p.projectName)
	if t.goos == "windows" {
		exe += ".exe"
	}

	ldflags := ""
	if f := p.linkFlags(); f != "" {
		ldflags = "-ldflags=" + f
	}

	err = sh.RunWith(envFlags, mg.GoCmd(), "build", "-o", exe, ldflags)
	return exe, err
}

// Clean removes the build directory
func (p *MageProject) Clean() error {
	fmt.Println("===== clean")

	if err := sh.Rm(p.buildDir); err != nil {
		fmt.Println(err)
	}

	return nil
}

// CleanAll removes the build directory and the docker image used for build
func (p *MageProject) CleanAll() error {
	fmt.Println("===== clean all")

	if err := sh.Rm(p.buildDir); err != nil {
		fmt.Println(err)
	}

	docker := sh.RunCmd("docker")
	dockerImage := "tmp/" + p.projectName + ".build"

	imgs, err := util.ExecOutput("docker", "images", "-q", dockerImage)
	if err != nil {
		fmt.Println(err)
	}

	imgs = util.TrimString(imgs)
	if len(imgs) > 0 {
		if err := docker("rmi", "--force", imgs); err != nil {
			fmt.Println(err)
		}
	}

	return nil
}

func (p *MageProject) checkBuildDir() {
	path := filepath.Join(p.MageLibrary().Workdir(), p.buildDir)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		os.Mkdir(path, 0755)
	}
}

// BuildWithDocker builds binary in build dir using a Dockerfile.build file
func (p *MageProject) BuildWithDocker() error {
	dockerImage := "tmp/" + p.projectName + ".build"

	p.checkBuildDir()

	docker := sh.RunCmd("docker")
	if err := docker("build", "-t", dockerImage, "-f", "Dockerfile.build", "."); err != nil {
		return err
	}

	vol := filepath.Join(p.mglib.Workdir(), p.buildDir) + ":" + filepath.Join(p.dckAppPath, p.buildDir)
	if err := docker("run", "-v", vol, dockerImage); err != nil {
		return err
	}

	return nil
}

// DockerBuildImage builds Docker image
func (p *MageProject) DockerBuildImage() error {
	fmt.Println("===== docker image")

	dck := p.mglib.DockerDetails(p.dckRegistry, p.dckImage, "")

	docker := sh.RunCmd("docker")
	if err := docker("build", "-t", dck.Image, "."); err != nil {
		return err
	}
	return nil
}

// DockerPushImage pushes Docker image to a repository
func (p *MageProject) DockerPushImage() error {
	fmt.Println("===== docker release")

	git, _ := p.mglib.GitDetails()
	if git.TagAtRev == "" {
		return errors.New("a git tag is needed to push Docker image")
	}

	dck := p.mglib.DockerDetails(p.dckRegistry, p.dckImage, "")

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
func (p *MageProject) PrintInfo() string {
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

	art := p.mglib.ArtifactDetails(p.artifactURL, "")
	sb.WriteString(`"artifact": {`)
	sb.WriteString(`"url": "` + art.URL + `"`)
	sb.WriteString("}, ")

	dck := p.mglib.DockerDetails(p.dckRegistry, p.dckImage, "")
	sb.WriteString(`"docker": {`)
	sb.WriteString(`"registry": "` + dck.Registry + `", `)
	sb.WriteString(`"image": "` + dck.Image + `"`)
	sb.WriteString("}")

	sb.WriteString("}")

	return sb.String()
}

// ChangeLog generates a ChangeLog based on git history
func (p *MageProject) ChangeLog() error {
	return p.mglib.ChangeLog("ChangeLog.md", p.artifactURL, p.gitURL)
}
