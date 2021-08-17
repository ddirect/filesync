package main

import (
	"bytes"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/ddirect/check"
	"github.com/ddirect/filemeta"
	ft "github.com/ddirect/filetest"
	"github.com/ddirect/protostream"
	"github.com/ddirect/xrand"
)

func compareDir(errf errorfunc, td *ft.Dir, db *Db) {
	tdFiles := td.AllFilesMap()
	for fpath, tf := range tdFiles {
		if dbf := db.FilesByPath[fpath]; dbf == nil {
			errf("%s not found in db", fpath)
		} else {
			if bytes.Compare(dbf.Hash, tf.Hash) != 0 {
				errf("hash mismatch for %s", fpath)
			}
		}
	}
	for _, dbf := range db.Files {
		if tf := tdFiles[dbf.Path]; tf == nil {
			errf("extra file %s found in db", dbf.Path)
		}
	}
}

func checkDbConsistency(errf errorfunc, db *Db) {
	if len(db.Dirs) != len(db.DirsByPath) {
		errf("len(db.Dirs) != len(db.DirsByPath)")
	}
	if len(db.Files) != len(db.FilesByPath) {
		errf("len(db.Files) != len(db.FilesByPath)")
	}
	for _, d := range db.Dirs {
		if x := db.DirsByPath[d.Path]; x == nil {
			errf("db.DirsByPath['%s'] not found - found in db.Dirs", d.Path)
		} else if x != d {
			errf("db.DirsByPath['%s'] doesn't match db.Dirs", d.Path)
		}
	}
	byHash := make(map[filemeta.HashKey]*File)
	for _, f := range db.Files {
		parent, name := filepath.Split(f.Path)
		if parent != "" {
			parent = parent[:len(parent)-1]
		}
		if name != f.Name {
			errf("name in path is '%s' instead of '%s'", name, f.Name)
		}
		if x := db.DirsByPath[parent]; x == nil {
			errf("db.DirsByPath['%s'] not found - found in db.Files", parent)
		}
		if x := db.FilesByPath[f.Path]; x == nil {
			errf("db.FilesByPath['%s'] not found - found in db.Files", name)
		} else if x != f {
			errf("db.FilesByPath['%s'] doesn't match db.Files", name)
		}
		hashKey := filemeta.ToHashKey(f.Hash)
		if x := db.FilesByHash[hashKey]; x == nil {
			errf("db.FilesByHash[%02x] not found - found in db.Files - %s", f.Hash, name)
		}
		byHash[hashKey] = f
	}
	if len(db.FilesByHash) != len(byHash) {
		errf("len(db.FilesByHash) != len(hash map extracted from db.Files)")
	}
}

func checkSerDes(errf errorfunc, db *Db) {
	r, w, err := os.Pipe()
	check.E(err)
	go func() {
		s := protostream.New(w)
		db.Send(s)
		check.E(s.Flush())
		check.E(w.Close())
	}()
	rdb := RecvDb(protostream.New(r))
	check.E(r.Close())
	checkDbConsistency(errf, rdb)
	// once consistency is verified it's enough to check that the
	// dir and file arrays are equal
	if !reflect.DeepEqual(db.Dirs, rdb.Dirs) {
		errf("dirs mismatch on serdes")
	}
	if !reflect.DeepEqual(db.Files, rdb.Files) {
		errf("files mismatch on serdes")
	}
}

func TestDb(t *testing.T) {
	rnd := xrand.New()

	base := t.TempDir()

	tree := ft.NewRandomTree(rnd, treeOptions())
	st := ft.CommitMixed(rnd, tree, ft.DefaultMixes(), base)

	db := readDbCore(base, filemeta.Refresh)

	errf := t.Errorf
	if len(db.FilesByHash) != st.UniqueHashes {
		errf("len(db.FilesByHash) != UniqueHashes")
	}
	checkDbConsistency(errf, db)
	compareDir(errf, tree, db)
	checkSerDes(errf, db)
}
