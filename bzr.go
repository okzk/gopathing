package main

import (
	"html/template"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

var bzrTemplate = template.Must(template.New("bzr template").Parse(`
if [ ! -d "$GOPATH/src/{{.Package}}" ]; then
  mkdir -p $(dirname "$GOPATH/src/{{.Package}}")
  bzr branch "{{.Repo}}" "$GOPATH/src/{{.Package}}"
fi
cd "$GOPATH/src/{{.Package}}"
bzr pull
bzr revert -r "{{.Rev}}"
`))

type bzr struct {
	Package string
	Repo    string
	Rev     string
}

func isBzr(path string) bool {
	return existsDir(filepath.Join(path, ".bzr"))
}

func newBzr(srcDir, pkg string) (*bzr, error) {
	dir := filepath.Join(srcDir, pkg)

	repoCmd := exec.Command("bzr", "config", "parent_location")
	repoCmd.Dir = dir
	repoCmd.Stderr = os.Stderr
	repo, err := repoCmd.Output()
	if err != nil {
		return nil, err
	}

	revCmd := exec.Command("bzr", "log", "-r-1", "--line")
	revCmd.Dir = dir
	revCmd.Stderr = os.Stderr
	revLine, err := revCmd.Output()
	if err != nil {
		return nil, err
	}
	rev := regexp.MustCompile("^[0-9]+").FindString(strings.TrimSpace(string(revLine)))

	return &bzr{Package: pkg, Repo: strings.TrimSpace(string(repo)), Rev: rev}, nil
}

func (b *bzr) print() {
	bzrTemplate.Execute(os.Stdout, b)
}
