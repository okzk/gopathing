package main

import (
	"fmt"
	"go/build"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const shellHeader = `#!/bin/bash -e

if [ -z "$GOPATH" ]; then
  echo 'Missing $GOPATH' >&2
  exit 1
fi
GOPATH=$(echo $GOPATH | cut -f1 -d:)
`

func exitIfError(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func main() {
	dir, err := os.Getwd()
	exitIfError(err)

	importMap := map[string]bool{}
	err = retrieveImportRecursively(importMap, dir)
	exitIfError(err)
	if len(importMap) == 0 {
		fmt.Fprintln(os.Stderr, "No package dependency!")
		return
	}

	vcsMap := map[string]vcs{}
	for path, _ := range importMap {
		err := retrieveVCS(vcsMap, path)
		exitIfError(err)
	}

	fmt.Print(shellHeader)
	for _, k := range sortedKeys(vcsMap) {
		vcsMap[k].print()
	}
}

func retrieveImportRecursively(importMap map[string]bool, dir string) error {
	err := retrieveImport(importMap, ".", dir)
	if err != nil {
		return err
	}

	list, err := ioutil.ReadDir(dir)
	if err != nil {
		return err
	}

	for _, fi := range list {
		if !fi.IsDir() {
			continue
		}
		n := fi.Name()
		if n == "vendor" || strings.HasPrefix(n, ".") || strings.HasPrefix(n, "_") {
			continue
		}
		retrieveImportRecursively(importMap, filepath.Join(dir, n))
	}

	return nil
}

func retrieveImport(importMap map[string]bool, path string, dir string) error {
	build.Default.SrcDirs()
	pkg, err := build.Import(path, dir, build.AllowBinary)
	if err != nil {
		if _, ok := err.(*build.NoGoError); ok {
			return nil
		} else {
			return err
		}
	}

	for _, path := range pkg.Imports {
		if isStandardImport(path) {
			continue
		}
		if importMap[path] {
			continue
		}

		if !build.IsLocalImport(path) {
			importMap[path] = true
		}

		err := retrieveImport(importMap, path, pkg.Dir)
		if err != nil {
			return err
		}
	}
	return nil
}

func isStandardImport(path string) bool {
	return !strings.Contains(path, ".")
}

func sortedKeys(vcsMap map[string]vcs) []string {
	keys := make([]string, 0, len(vcsMap))
	for k, _ := range vcsMap {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	return keys
}
