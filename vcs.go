package main

import (
	"fmt"
	"go/build"
	"os"
	"path"
	"path/filepath"
)

type vcs interface {
	print()
}

func existsDir(dir string) bool {
	fi, err := os.Stat(dir)
	return err == nil && fi.IsDir()
}

func retrieveVCS(vcsMap map[string]vcs, pkg string) (err error) {
	for _, srcDir := range build.Default.SrcDirs() {
		if !existsDir(filepath.Join(srcDir, pkg)) {
			continue
		}

		for p := pkg; p != ""; p = path.Dir(p) {
			if _, ok := vcsMap[p]; ok {
				return
			}
			dir := filepath.Join(srcDir, p)
			if isGit(dir) {
				vcsMap[p], err = newGit(srcDir, p)
				return
			} else if isBzr(dir) {
				vcsMap[p], err = newBzr(srcDir, p)
				return
			} else if isHg(dir) {
				vcsMap[p], err = newHg(srcDir, p)
				return
			}
		}
		err = fmt.Errorf("missing VCS at %s", pkg)
		return
	}

	err = fmt.Errorf("missing package: %s", pkg)
	return
}
