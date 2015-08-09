package main

import (
	"html/template"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

var hgTemplate = template.Must(template.New("hg template").Parse(`
if [ ! -d "$GOPATH/src/{{.Package}}" ]; then
  mkdir -p $(dirname "$GOPATH/src/{{.Package}}")
  hg clone "{{.Repo}}" "$GOPATH/src/{{.Package}}"
fi
cd "$GOPATH/src/{{.Package}}"
hg pull
hg update "{{.Rev}}"
`))

type hg struct {
	Package string
	Repo    string
	Rev     string
}

func isHg(path string) bool {
	return existsDir(filepath.Join(path, ".hg"))
}

func newHg(srcDir, pkg string) (*hg, error) {
	dir := filepath.Join(srcDir, pkg)

	repoCmd := exec.Command("hg", "showconfig", "paths.default")
	repoCmd.Dir = dir
	repoCmd.Stderr = os.Stderr
	repo, err := repoCmd.Output()
	if err != nil {
		return nil, err
	}

	revCmd := exec.Command("hg", "id", "-i")
	revCmd.Dir = dir
	revCmd.Stderr = os.Stderr
	rev, err := revCmd.Output()
	if err != nil {
		return nil, err
	}

	return &hg{Package: pkg, Repo: strings.TrimSpace(string(repo)), Rev: strings.TrimSpace(string(rev))}, nil
}

func (h *hg) print() {
	hgTemplate.Execute(os.Stdout, h)
}
