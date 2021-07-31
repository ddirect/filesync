package main

import (
	"net"

	"github.com/ddirect/check"
	"github.com/ddirect/filesync/records"
	"github.com/ddirect/protostream"
)

type SyncActionsFactory func(sdb *Db, ddb *Db, basePath string, ps protostream.ReadWriter) SyncActions

func Sync(db *Db, basePath string, netAddr NetAddr, syncActionsFactory SyncActionsFactory) {
	conn, err := net.Dial(netAddr())
	check.E(err)
	defer func() {
		if conn != nil {
			conn.Close()
		}
	}()
	syncConnection(db, basePath, conn, syncActionsFactory)
}

func syncConnection(db *Db, basePath string, conn net.Conn, syncActionsFactory SyncActionsFactory) {
	ps := protostream.New(conn)
	SimpleCommandSender(ps)(records.Command_GETDB)
	ps.Flush()
	rdb := RecvDb(ps)
	SyncCore(rdb, db, syncActionsFactory(rdb, db, basePath, ps))
}
