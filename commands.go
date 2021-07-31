package main

import (
	"github.com/ddirect/check"
	"github.com/ddirect/filesync/records"
	"github.com/ddirect/protostream"
)

func SimpleCommandSender(ps protostream.ReadWriter) func(records.Command_Op) {
	r := new(records.Command)
	return func(op records.Command_Op) {
		r.Op = op
		check.E(ps.WriteMessage(r))
	}
}

func GetFileCommandSender(ps protostream.ReadWriter) func(hash []byte) {
	r := &records.Command{Op: records.Command_GETFILE}
	return func(hash []byte) {
		r.Hash = hash
		check.E(ps.WriteMessage(r))
		check.E(ps.Flush())
	}
}
