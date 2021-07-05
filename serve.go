package main

import (
	"log"
	"net"

	"github.com/ddirect/check"
	"github.com/ddirect/filesync/records"
	"github.com/ddirect/protostream"
)

func serve(db *Db, basePath string, netAddr NetAddr) {
	l, err := net.Listen(netAddr())
	check.E(err)
	log.Printf("listening on %s", l.Addr())
	for {
		conn, err := l.Accept()
		check.El(err)
		log.Printf("connection from %s", conn.RemoteAddr())
		check.El(serveConn(conn, db, basePath))
		check.El(conn.Close())
	}
}

func serveConn(conn net.Conn, db *Db, basePath string) (err error) {
	defer check.Recover(&err)
	ps := protostream.New(conn)
	command := new(records.Command)
	for {
		check.E(ps.ReadMessage(command))
		switch command.Op {
		case records.Command_GETDB:
			db.Send(ps)
		case records.Command_GETFILE:
		}
	}
	return
}
