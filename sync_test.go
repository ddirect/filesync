package main

import (
	"fmt"
	"io/fs"
	"net"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/ddirect/check"
	"github.com/ddirect/filemeta"
	"github.com/ddirect/filetest"
	ft "github.com/ddirect/filetest"
	"github.com/ddirect/xrand"
)

type checkSyncOptions struct {
	missingMeta int //percent
}

func TestSyncToEmptyFullMeta(t *testing.T) {
	testSyncToEmpty(t, checkSyncOptions{})
}

func TestSyncToEmptyPartialMeta(t *testing.T) {
	testSyncToEmpty(t, checkSyncOptions{25})
}

func testSyncToEmpty(t *testing.T, checkOpt checkSyncOptions) {
	rnd := xrand.New()

	sBase := ft.TempDir(t, "sour")
	dBase := ft.TempDir(t, "dest")

	sTree := ft.NewRandomTree(rnd, treeOptions())
	ft.CommitMixed(rnd, sTree, ft.DefaultMixes(), sBase)

	checkSync(t, sBase, sTree, dBase, checkOpt)
}

func clone(s []*ft.File) []*ft.File {
	d := make([]*ft.File, len(s))
	copy(d, s)
	return d
}

func TestSyncRelatedFullMeta(t *testing.T) {
	testSyncRelated(t, checkSyncOptions{})
}

func TestSyncRelatedPartialMeta(t *testing.T) {
	testSyncRelated(t, checkSyncOptions{25})
}

func testSyncRelated(t *testing.T, checkOpt checkSyncOptions) {
	rnd := xrand.New()

	o := treeOptions()
	o.Depth++

	// the tree instance is used first to build the destination tree
	// then the source tree; it is then kept as reference of the latter
	tree, nameFactory := ft.NewRandomTree2(rnd, o)
	dataRnd1, dataRnd2 := xrand.NewPair()

	sBase := ft.TempDir(t, "sour")
	dBase := ft.TempDir(t, "dest")

	sMixes := ft.DefaultMixes()
	dMixes := sMixes
	dMixes.Created = 100 // create only unique files, or the final unicity test will fail
	zones := ft.DefaultZones()
	files := tree.AllFilesSlice()

	ft.CommitDirs(tree, dBase)
	dDs, dExc := ft.CommitZonedFilesMixed(dataRnd1, rnd, files, zones, dMixes, dBase, false)
	excluded := clone(dExc)
	notChanged := clone(files[:len(files)*zones.NoChange/100])

	tree.EachDirRecursive(func(d *ft.Dir) {
		//rename 1/10 of the directories
		if rnd.Intn(10) == 0 {
			d.Name = nameFactory()
		}
	})
	ft.CommitDirs(tree, sBase)
	sDs, sExc := ft.CommitZonedFilesMixed(dataRnd2, rnd, files, zones, sMixes, sBase, true)
	if !reflect.DeepEqual(notChanged, files[:len(notChanged)]) {
		t.Fatal("equal zone mismatch")
	}
	tab := ft.DirStatsTable()
	tab.SetHeading(".", "sour", "dest")
	sDs.AppendToTable(tab)
	dDs.AppendToTable(tab)
	fmt.Println(tab)
	if reflect.DeepEqual(sExc, excluded) {
		t.Fatal("same files excluded")
	}

	tree.RemoveFiles(sExc)
	checkSync(t, sBase, tree, dBase, checkOpt)
}

func checkSync(t *testing.T, sBase string, sTree *filetest.Dir, dBase string, opt checkSyncOptions) {
	fetch := filemeta.OpRefresh
	onlyWithMeta := false
	if opt.missingMeta > 0 {
		// precalculate the metadata and use it for the dbs; note that this needs to be
		// done in advance (not in readDbCore) or links to files which will get a hash
		// won't be synchronized
		rnd := xrand.New()
		refresh := func(base string) {
			Walk(base,
				func(relPath string) int {
					return 0
				},
				func(relPath string, name string, dirId int) {
					if rnd.Intn(100) >= opt.missingMeta {
						check.E(filemeta.Refresh(filepath.Join(base, relPath)).Error)
					}
				})
		}
		refresh(sBase)
		// can't refresh part of dBase or the test will fail due to the additional
		// files tracked in the destination
		fetch = filemeta.OpGet
		onlyWithMeta = true
	}

	c1, c2 := net.Pipe()
	go func() {
		check.E(serveConnection(c1, readDbCore(sBase, fetch), sBase))
	}()
	syncConnection(readDbCore(dBase, fetch), dBase, c2, RecvActionsFactory)

	var dTree *filetest.Dir
	if onlyWithMeta {
		filterFactory := func(base string) func(e ft.Entry) bool {
			return func(e ft.Entry) bool {
				return len(filemeta.Get(filepath.Join(base, e.Path())).Hash) > 0
			}
		}
		sTree = filetest.NewDirFromStorageFiltered(sBase, filterFactory(sBase))
		dTree = filetest.NewDirFromStorageFiltered(dBase, filterFactory(dBase))
	} else {
		checkMetaValidAndFilesUnique(t, dBase)
		dTree = filetest.NewDirFromStorage(dBase)
	}
	if !sTree.Compare(dTree) {
		t.Fail()
	}
}

func checkMetaValidAndFilesUnique(t *testing.T, base string) {
	async := filemeta.AsyncOperations(filemeta.OpGet, 0, 0)
	go func() {
		defer close(async.FileIn)
		filepath.WalkDir(base, func(path string, d fs.DirEntry, err error) error {
			check.E(err)
			if !d.IsDir() {
				async.FileIn <- path
			}
			return nil
		})
	}()
	defer func() {
		// ensure the tree is walked completely before leaving, which will cause
		// the tree to be deleted and the walk goroutine to panic
		for range async.DataOut {
		}
	}()
	inodes := make(map[filemeta.HashKey]uint64)
	for data := range async.DataOut {
		check.E(data.Error)
		if len(data.Hash) == 0 {
			t.Fatal("missing hash on", data.Path)
		}
		hashKey := filemeta.ToHashKey(data.Hash)
		if inode, ok := inodes[hashKey]; !ok {
			inodes[hashKey] = data.Inode
		} else {
			if inode != data.Inode {
				t.Fatal("found same hash on multiple files")
			}
		}
	}
}
