package main

import (
	"bytes"

	"github.com/ddirect/filemeta"
)

type syncActions struct {
	CreateDir  func(*Dir)
	RemoveDir  func(*Dir)
	RemoveFile func(*File)
	CopyFile   func(*File)
	LinkFile   func(*File, *File)
	StashFile  func(*File)
	Epilogue   func()
}

type SyncActions *syncActions

func SyncCore(sdb *Db, ddb *Db, actions SyncActions) {
	createDirs(sdb, ddb, actions)
	syncFiles(sdb, ddb, actions)
	removeFiles(sdb, ddb, actions)
	removeDirs(sdb, ddb, actions)
	actions.Epilogue()
}

func forEachMissingDir(ref []*Dir, test map[string]*Dir, action func(*Dir)) {
	for _, x := range ref {
		if test[x.Path] == nil {
			action(x)
		}
	}
}

func forEachMissingDirReverse(ref []*Dir, test map[string]*Dir, action func(*Dir)) {
	for i := len(ref) - 1; i >= 0; i-- {
		x := ref[i]
		if test[x.Path] == nil {
			action(x)
		}
	}
}

func forEachMissingFile(ref []*File, test map[string]*File, action func(*File)) {
	for _, x := range ref {
		if test[x.Path] == nil {
			action(x)
		}
	}
}

func createDirs(sdb *Db, ddb *Db, actions SyncActions) {
	forEachMissingDir(sdb.Dirs, ddb.DirsByPath, actions.CreateDir)
}

func removeDirs(sdb *Db, ddb *Db, actions SyncActions) {
	forEachMissingDirReverse(ddb.Dirs, sdb.DirsByPath, actions.RemoveDir)
}

func removeFiles(sdb *Db, ddb *Db, actions SyncActions) {
	forEachMissingFile(ddb.Files, sdb.FilesByPath, actions.RemoveFile)
}

func syncFiles(sdb *Db, ddb *Db, actions SyncActions) {
	for _, sfile := range sdb.Files {
		if dfile := ddb.FilesByPath[sfile.Path]; dfile != nil {
			if bytes.Compare(sfile.Hash, dfile.Hash) == 0 {
				continue
			}
			actions.StashFile(dfile)
		}
		if dfile := ddb.FilesByHash[filemeta.ToHashKey(sfile.Hash)]; dfile != nil {
			actions.LinkFile(dfile, sfile)
		} else {
			actions.CopyFile(sfile)
		}
	}
}
