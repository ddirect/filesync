package main

import (
	"net"
	"reflect"
	"testing"

	"github.com/ddirect/check"
	"github.com/ddirect/filemeta"
	"github.com/ddirect/filetest"
	"github.com/ddirect/xrand"
)

func TestSyncToEmpty(t *testing.T) {
	rnd := xrand.New()

	sBase := t.TempDir()
	dBase := t.TempDir()

	sTree := createTree(rnd)
	commit1(sTree, rnd, sBase)

	checkSync(t, sBase, sTree, dBase)
}

func checkSync(t *testing.T, sBase string, sTree *filetest.Dir, dBase string) {
	c1, c2 := net.Pipe()
	go func() {
		check.E(serveConnection(c1, readDbCore(sBase, filemeta.Refresh), sBase))
	}()
	syncConnection(readDbCore(dBase, filemeta.Refresh), dBase, c2, RecvActionsFactory)
	if !reflect.DeepEqual(sTree, filetest.NewDirFromStorage(dBase)) {
		t.Fail()
	}
}
