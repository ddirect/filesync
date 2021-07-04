package main

import (
	"os"
	"path/filepath"

	"github.com/ddirect/check"
)

func Walk(base string, dirCb func(relPath string) int, fileCb func(relPath string, name string, dirId int)) {
	var walk func(string, int)
	walk = func(relPath string, dirId int) {
		absPath := filepath.Join(base, relPath)
		entries, err := os.ReadDir(absPath)
		check.E(err)
		for _, e := range entries {
			mode := e.Type()
			if mode.IsRegular() {
				name := e.Name()
				fileCb(filepath.Join(relPath, name), name, dirId)
			} else if mode.IsDir() {
				newPath := filepath.Join(relPath, e.Name())
				walk(newPath, dirCb(newPath))
			}
		}
	}
	walk("", dirCb(""))
}
