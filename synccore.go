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
	stashFiles(sdb, ddb, actions.StashFile)
	copyFiles(sdb, ddb, actions.CopyFile)
	linkFiles(sdb, ddb, actions.LinkFile)
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

func stashFiles(sdb *Db, ddb *Db, stashFile func(*File)) {
	var w int
	files := sdb.Files
	for _, sfile := range files {
		if len(sfile.Hash) == 0 {
			continue
		}
		if dfile := ddb.FilesByPath[sfile.Path]; dfile != nil {
			if bytes.Compare(sfile.Hash, dfile.Hash) == 0 {
				continue
			}
			stashFile(dfile)
		}
		files[w] = sfile
		w++
	}
	sdb.Files = files[:w]
}

func copyFiles(sdb *Db, ddb *Db, copyFile func(*File)) {
	var w int
	files := sdb.Files
	for _, sfile := range files {
		if dfile := ddb.FilesByHash[filemeta.ToHashKey(sfile.Hash)]; dfile == nil {
			copyFile(sfile)
		} else {
			files[w] = sfile
			w++
		}
	}
	sdb.Files = files[:w]
}

func linkFiles(sdb *Db, ddb *Db, linkFile func(*File, *File)) {
	for _, sfile := range sdb.Files {
		linkFile(ddb.FilesByHash[filemeta.ToHashKey(sfile.Hash)], sfile)
	}
}
