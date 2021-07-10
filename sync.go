package main

import (
	"bytes"
)

type SyncAgent interface {
	CreateDir(string)
	RemoveDir(string)
	CopyFile(*File)
	LinkFile(*File, string)
	StashFile(*File)
}

func Sync(sdb *Db, ddb *Db, agent SyncAgent) {
	createDirs(sdb, ddb, agent)
	syncFiles(sdb, ddb, agent)
	removeDirs(sdb, ddb, agent)
}

func forEachMissingDir(ref []*Dir, test map[string]*Dir, action func(string)) {
	for _, dir := range ref {
		if test[dir.Path] == nil {
			action(dir.Path)
		}
	}
}

func createDirs(sdb *Db, ddb *Db, agent SyncAgent) {
	forEachMissingDir(sdb.Dirs, ddb.DirsByPath, agent.CreateDir)
}

func removeDirs(sdb *Db, ddb *Db, agent SyncAgent) {
	forEachMissingDir(ddb.Dirs, sdb.DirsByPath, agent.RemoveDir)
}

func syncFiles(sdb *Db, ddb *Db, agent SyncAgent) {
	for _, sfile := range sdb.Files {
		if dfile := ddb.FilesByPath[sfile.Path]; dfile != nil {
			if bytes.Compare(sfile.Hash, dfile.Hash) == 0 {
				continue
			}
			agent.StashFile(dfile)
		}
		if dfile := ddb.FilesByHash[toHashKey(sfile.Hash)]; dfile != nil {
			agent.LinkFile(dfile, sfile.Path)
		} else {
			agent.CopyFile(sfile)
		}
	}
}
