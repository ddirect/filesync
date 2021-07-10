package main

import (
	"github.com/ddirect/check"
	"github.com/ddirect/filesync/records"
	"github.com/ddirect/protostream"
)

type Dir struct {
	Path   string
	TimeNs int64
}

func DirRecordSender(ps protostream.ReadWriter) func(*Dir) {
	r := new(records.Dir)
	return func(d *Dir) {
		r.Path = d.Path
		r.TimeNs = d.TimeNs
		check.E(ps.WriteMessage(r))
	}
}

func DirRecordReceiver(ps protostream.ReadWriter, db *Db) func(*Dir) {
	r := new(records.Dir)
	return func(d *Dir) {
		check.E(ps.ReadMessage(r))
		d.Path = r.Path
		d.TimeNs = r.TimeNs
		db.DirsByPath[d.Path] = d
	}
}
