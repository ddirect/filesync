package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/ddirect/check"
	"github.com/ddirect/filemeta"
	"github.com/ddirect/filesync/records"
	"github.com/ddirect/protostream"
	"github.com/ddirect/sys"
)

const cacheAttr = "filesync.cache"

func ensureCacheDir() string {
	d := filepath.Join(os.TempDir(), "ddirect.filesync")
	check.E(os.MkdirAll(d, 0777))
	return d
}

type CacheInfo struct {
	records.CacheMeta
	File string
}

func ReadCache(basePath string) (db *Db, cand *CacheInfo) {
	cacheDir := ensureCacheDir()
	path := check.SE(filepath.Abs(basePath))
	fi, err := sys.Stat(path)
	check.E(err)
	entries, err := os.ReadDir(cacheDir)
	check.E(err)

	var toRemove []string
	for _, e := range entries {
		ci := &CacheInfo{File: filepath.Join(cacheDir, e.Name())}
		if filemeta.ReadCustom(ci.File, cacheAttr, ci) == nil {
			if ci.Path == path && ci.Device == fi.Device {
				if cand != nil {
					toRemove = append(toRemove, cand.File)
				}
				cand = ci
			}
		} else {
			toRemove = append(toRemove, ci.File)
		}
	}
	for _, x := range toRemove {
		check.E(os.Remove(x))
	}
	if cand != nil {
		db = readDbFromFile(cand.File)
	}
	return
}

func WriteCache(basePath string, db *Db) {
	defer func() {
		if e := recover(); e != nil {
			fmt.Fprintf(os.Stderr, "cannot write cache: %v\n", e)
		}
	}()
	path := check.SE(filepath.Abs(basePath))
	fi, err := sys.Stat(path)
	check.E(err)
	meta := records.NewCacheMeta(path, fi.Device)
	f, err := os.Create(filepath.Join(ensureCacheDir(), fmt.Sprintf("%0.19d", meta.TimeNs)))
	check.E(err)
	defer f.Close() // safe
	stream := protostream.New(f)
	db.Send(stream)
	check.E(stream.Writer.Flush())
	f1 := f
	f = nil
	check.E(f1.Close())
	check.E(filemeta.WriteCustom(f1.Name(), cacheAttr, meta))
}

func readDbFromFile(fileName string) *Db {
	f, err := os.Open(fileName)
	check.E(err)
	defer f.Close()
	return RecvDb(protostream.New(f))
}
