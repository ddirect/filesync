package main

import (
	"net"

	"github.com/ddirect/check"
	"github.com/ddirect/filesync/records"
	"github.com/ddirect/protostream"
)

func recv(db *Db, basePath string, netAddr NetAddr) {
	return
	conn, err := net.Dial(netAddr())
	check.E(err)
	defer func() {
		if conn != nil {
			conn.Close()
		}
	}()

	ps := protostream.New(conn)
	sendCommand := CommandSender(ps)
	sendCommand(records.Command_GETDB)
	rdb := RecvDb(ps)
	Sync(rdb, db, nil)
}
