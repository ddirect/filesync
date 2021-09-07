package main

import (
	"errors"
	"fmt"
	"io"
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
		if check.El(serveConnection(conn, db, basePath)) {
			log.Printf("connection closed")
		}
	}
}

func serveConnection(conn net.Conn, db *Db, basePath string) (err error) {
	defer check.Recover(&err)
	defer check.DeferredE(conn.Close)
	ps := protostream.New(conn)
	command := new(records.Command)
	serveFile := FileDataSender(ps, db, basePath)
	for {
		if err = ps.ReadMessage(command); err != nil {
			if errors.Is(err, io.EOF) {
				err = nil
			}
			return
		}
		switch command.Op {
		case records.Command_GETDB:
			log.Println("serving db")
			db.Send(ps)
		case records.Command_GETFILE:
			serveFile(command.Hash)
		default:
			panic(fmt.Errorf("unknown command %v\n", command.Op))
		}
		check.E(ps.Flush())
	}
}
