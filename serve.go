package main

import (
	"fmt"
	"log"
	"net"

	"github.com/ddirect/check"
	"github.com/ddirect/filesync/records"
	"github.com/ddirect/protostream"
)

func Serve(db *Db, basePath string, netAddr NetAddr) {
	l, err := net.Listen(netAddr())
	check.E(err)
	log.Printf("listening on %s", l.Addr())
	for {
		conn, err := l.Accept()
		check.El(err)
		log.Printf("connection from %s", conn.RemoteAddr())
		check.El(serveConnection(conn, db, basePath))
	}
}

func serveConnection(conn net.Conn, db *Db, basePath string) (err error) {
	defer check.Recover(&err)
	defer check.DeferredE(conn.Close)
	ps := protostream.New(conn)
	command := new(records.Command)
	serveFile := FileDataSender(ps, db, basePath)
	for {
		check.E(ps.ReadMessage(command))
		switch command.Op {
		case records.Command_GETDB:
			log.Println("serving db")
			db.Send(ps)
		case records.Command_GETFILE:
			log.Printf("serving %02x", command.Hash)
			serveFile(command.Hash)
		default:
			panic(fmt.Errorf("unknown command %v\n", command.Op))
		}
		check.E(ps.Flush())
	}
}
