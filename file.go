package main

import (
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

func FileRecordSender(ps protostream.ReadWriter) func(**File) {
	r := new(records.File)
	return func(fp **File) {
		f := *fp
		r.Name = f.Name
		r.DirIndex = int64(f.DirIndex)
		r.Hash = f.Hash
		r.TimeNs = f.TimeNs
		r.Size = f.Size
		check.E(ps.WriteMessage(r))
	}
}

func FileRecordReceiver(ps protostream.ReadWriter, db *Db) func(**File) {
	r := new(records.File)
	return func(fp **File) {
		check.E(ps.ReadMessage(r))
		f := new(File)
		*fp = f
		f.Name = r.Name
		f.DirIndex = int(r.DirIndex)
		f.Hash = r.Hash
		f.TimeNs = r.TimeNs
		f.Size = r.Size
	}
}
