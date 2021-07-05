package main

import (
	"github.com/ddirect/check"
	"github.com/ddirect/filesync/records"
	"github.com/ddirect/protostream"
)

func CommandSender(ps protostream.ReadWriter) func(records.Command_Op) {
	r := new(records.Command)
	return func(op records.Command_Op) {
		r.Op = op
		check.E(ps.WriteMessage(r))
	}
}
