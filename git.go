package main

import (
	"html/template"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

var gitTemplate = template.Must(template.New("git template").Parse(`
if [ ! -d "$GOPATH/src/{{.Package}}" ]; then
  mkdir -p $(dirname "$GOPATH/src/{{.Package}}")
  git clone "{{.Repo}}" "$GOPATH/src/{{.Package}}"
fi
cd $GOPATH/src/{{.Package}}
git fetch
git checkout "{{.Rev}}"{{if .SubModule}}
git submodule update --init --recursive{{end}}
`))

var gitSubmoduleRegex = regexp.MustCompilePOSIX("^submodule.")

type git struct {
	Package   string
	Repo      string
	Rev       string
	SubModule bool
}

func isGit(path string) bool {
	return existsDir(filepath.Join(path, ".git"))
}

func newGit(srcDir, pkg string) (*git, error) {
	dir := filepath.Join(srcDir, pkg)

	repoCmd := exec.Command("git", "config", "remote.origin.url")
	repoCmd.Dir = dir
	repoCmd.Stderr = os.Stderr
	repo, err := repoCmd.Output()
	if err != nil {
		return nil, err
	}

	revCmd := exec.Command("git", "rev-parse", "HEAD")
	revCmd.Dir = dir
	revCmd.Stderr = os.Stderr
	rev, err := revCmd.Output()
	if err != nil {
		return nil, err
	}

	confCmd := exec.Command("git", "config", "-l")
	confCmd.Dir = dir
	confCmd.Stderr = os.Stderr
	conf, err := confCmd.Output()
	if err != nil {
		return nil, err
	}
	subModule := gitSubmoduleRegex.Find(conf) != nil

	return &git{
		Package:   pkg,
		Repo:      strings.TrimSpace(string(repo)),
		Rev:       strings.TrimSpace(string(rev)),
		SubModule: subModule,
	}, nil
}

func (g *git) print() {
	gitTemplate.Execute(os.Stdout, g)
}
