package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"

	"github.com/ddirect/check"
	"github.com/ddirect/filemeta"
	"github.com/ddirect/filesync/records"
	"github.com/ddirect/format"
	"github.com/ddirect/protostream"
)

type Db struct {
	Dirs        []*Dir
	Files       []*File
	DirsByPath  map[string]*Dir
	FilesByHash map[filemeta.HashKey]*File
	FilesByPath map[string]*File
}

func newDb() *Db {
	return &Db{
		DirsByPath:  make(map[string]*Dir),
		FilesByHash: make(map[filemeta.HashKey]*File),
		FilesByPath: make(map[string]*File),
	}
}

func ReadDb(basePath string, cache bool) (db *Db) {
	if cache {
		var tim time.Time
		db, tim = ReadCache(basePath)
		if db != nil {
			fmt.Fprintf(os.Stderr, "db cached on %s loaded\n", format.TimeMs(tim))
		}
	}
	if db == nil {
		db = readDbCore(basePath, filemeta.OpGet)
		WriteCache(basePath, db)
	}
	return
}

func readDbCore(basePath string, op filemeta.Op) *Db {
	workers := runtime.NumCPU()
	const queueBuf = 4000
	queue1 := make(chan *File, queueBuf)
	queue2 := make(chan *File, queueBuf)
	var wg1, wg2 sync.WaitGroup
	wg1.Add(workers)
	for i := 0; i < workers; i++ {
		go func() {
			for file := range queue1 {
				data := filemeta.Operation(op, filepath.Join(basePath, file.Path))
				check.E(data.Error)
				if len(data.Hash) > 0 {
					file.Hash = data.Hash
					file.TimeNs = data.ModTimeNs
					file.Size = data.Size
					queue2 <- file
				}
			}
			wg1.Done()
		}()
	}

	db := newDb()

	wg2.Add(1)
	go func() {
		for file := range queue2 {
			db.FilesByHash[filemeta.ToHashKey(file.Hash)] = file
		}
		wg2.Done()
	}()

	Walk(basePath, func(relPath string) int {
		stat, err := os.Stat(filepath.Join(basePath, relPath)) // TODO: or Lstat?
		check.E(err)
		dir := &Dir{relPath, stat.ModTime().UnixNano()}
		db.Dirs = append(db.Dirs, dir)
		db.DirsByPath[relPath] = dir
		return len(db.Dirs) - 1
	}, func(relPath string, name string, dirId int) {
		file := &File{Path: relPath, Name: name, DirIndex: dirId}
		db.Files = append(db.Files, file)
		db.FilesByPath[relPath] = file
		queue1 <- file
	})

	close(queue1)
	wg1.Wait()
	close(queue2)
	wg2.Wait()

	return db
}

func DbHeaderSender(ps protostream.ReadWriter) func(*Db) {
	r := new(records.DbHeader)
	return func(db *Db) {
		r.DirCount = int64(len(db.Dirs))
		r.FileCount = int64(len(db.Files))
		check.E(ps.WriteMessage(r))
	}
}

func DbHeaderReceiver(ps protostream.ReadWriter) func(*Db) {
	r := new(records.DbHeader)
	return func(db *Db) {
		check.E(ps.ReadMessage(r))
		da := make([]Dir, r.DirCount)
		d := make([]*Dir, len(da))
		for i := range d {
			d[i] = &da[i]
		}
		db.Dirs = d
		fa := make([]File, r.FileCount)
		f := make([]*File, len(fa))
		for i := range f {
			f[i] = &fa[i]
		}
		db.Files = f
	}
}

func (db *Db) codec(headerCodec func(*Db), dirCodec func(*Dir), fileCodec func(*File)) {
	headerCodec(db)
	dirs := db.Dirs
	for _, dir := range dirs {
		dirCodec(dir)
	}
	files := db.Files
	for _, file := range files {
		fileCodec(file)
	}
}

func (db *Db) Send(ps protostream.ReadWriter) {
	db.codec(DbHeaderSender(ps), DirRecordSender(ps), FileRecordSender(ps))
}

func RecvDb(ps protostream.ReadWriter) *Db {
	db := newDb()
	db.codec(DbHeaderReceiver(ps), DirRecordReceiver(ps, db), FileRecordReceiver(ps, db))
	return db
}
