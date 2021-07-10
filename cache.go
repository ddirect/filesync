package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

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

func ReadCache(basePath string) (*Db, time.Time) {
	cacheDir := ensureCacheDir()
	path := check.SE(filepath.Abs(basePath))
	fi, err := sys.Stat(path)
	check.E(err)
	entries, err := os.ReadDir(cacheDir)
	check.E(err)

	meta := new(records.CacheMeta)
	var toRemove []string
	var cand string
	var candNs int64
	for _, e := range entries {
		name := filepath.Join(cacheDir, e.Name())
		if filemeta.ReadCustom(name, cacheAttr, meta) == nil {
			if meta.Path == path && meta.Device == fi.Device {
				if cand != "" {
					toRemove = append(toRemove, cand)
				}
				cand = name
				candNs = meta.TimeNs
			}
		} else {
			toRemove = append(toRemove, name)
		}
	}
	for _, x := range toRemove {
		check.E(os.Remove(x))
	}
	if cand == "" {
		return nil, time.Time{}
	}
	return readDbFromFile(cand), time.Unix(0, candNs)
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
	f, err := os.Create(filepath.Join(ensureCacheDir(), strconv.FormatInt(meta.TimeNs, 10)))
	check.E(err)
	defer f.Close() // safe
	stream := protostream.New(f)
	db.Send(stream)
	check.E(stream.Flush())
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
