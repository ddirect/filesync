package main

import (
	"net"
	"reflect"
	"testing"

	"github.com/ddirect/check"
	"github.com/ddirect/filemeta"
	"github.com/ddirect/filetest"
	ft "github.com/ddirect/filetest"
	"github.com/ddirect/xrand"
)

func TestSyncToEmpty(t *testing.T) {
	rnd := xrand.New()

	sBase := t.TempDir()
	dBase := t.TempDir()

	sTree := ft.NewRandomTree(rnd, treeOptions())
	ft.CommitMixed(rnd, sTree, ft.DefaultMixes(), sBase)

	checkSync(t, sBase, sTree, dBase)
}

func clone(s []*ft.File) []*ft.File {
	d := make([]*ft.File, len(s))
	copy(d, s)
	return d
}

func TestSyncRelated(t *testing.T) {
	rnd := xrand.New()

	o := treeOptions()
	o.Depth++

	// the tree instance is used first to build the destination tree
	// then the source tree; it is then kept as reference of the latter
	tree, nameFactory := ft.NewRandomTree2(rnd, o)
	dataRnd1, dataRnd2 := xrand.NewPair()

	sBase := t.TempDir()
	dBase := t.TempDir()

	mixes := ft.DefaultMixes()
	zones := ft.DefaultZones()
	files := tree.AllFilesSlice()

	ft.CommitDirs(tree, dBase)
	dDs, dExc := ft.CommitZonedFilesMixed(dataRnd1, rnd, files, zones, mixes, dBase, false)
	excluded := clone(dExc)
	notChanged := clone(files[:len(files)*zones.NoChange/100])

	tree.EachDirRecursive(func(d *ft.Dir) {
		//rename 1/10 of the directories
		if rnd.Intn(10) == 0 {
			d.Name = nameFactory()
		}
	})
	ft.CommitDirs(tree, sBase)
	sDs, sExc := ft.CommitZonedFilesMixed(dataRnd2, rnd, files, zones, mixes, sBase, true)
	if !reflect.DeepEqual(notChanged, files[:len(notChanged)]) {
		t.Fatal("equal zone mismatch")
	}
	if sDs != dDs {
		t.Fatalf("different stats: %s <-> %s", sDs, dDs)
	}
	if reflect.DeepEqual(sExc, excluded) {
		t.Fatal("same files excluded")
	}

	tree.RemoveFiles(sExc)
	checkSync(t, sBase, tree, dBase)
}

func checkSync(t *testing.T, sBase string, sTree *filetest.Dir, dBase string) {
	c1, c2 := net.Pipe()
	go func() {
		check.E(serveConnection(c1, readDbCore(sBase, filemeta.Refresh), sBase))
	}()
	syncConnection(readDbCore(dBase, filemeta.Refresh), dBase, c2, RecvActionsFactory)
	dTree := filetest.NewDirFromStorage(dBase)
	if !sTree.Compare(dTree) {
		t.Fail()
	}
}
