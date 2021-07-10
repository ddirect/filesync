package main

import (
	"path/filepath"

	"github.com/ddirect/check"
	"github.com/ddirect/filesync/records"
	"github.com/ddirect/protostream"
)

type File struct {
	Path     string
	Name     string
	DirIndex int
	Hash     []byte
	TimeNs   int64
	Size     int64
}

func FileRecordSender(ps protostream.ReadWriter) func(*File) {
	r := new(records.File)
	return func(f *File) {
		r.Name = f.Name
		r.DirIndex = int64(f.DirIndex)
		r.Hash = f.Hash
		r.TimeNs = f.TimeNs
		r.Size = f.Size
		check.E(ps.WriteMessage(r))
	}
}

func FileRecordReceiver(ps protostream.ReadWriter, db *Db) func(*File) {
	r := new(records.File)
	return func(f *File) {
		check.E(ps.ReadMessage(r))
		f.Name = r.Name
		di := int(r.DirIndex)
		f.DirIndex = di
		f.Hash = r.Hash
		f.TimeNs = r.TimeNs
		f.Size = r.Size

		f.Path = filepath.Join(db.Dirs[di].Path, f.Name)
		db.FilesByPath[f.Path] = f
		db.FilesByHash[toHashKey(f.Hash)] = f
	}
}
