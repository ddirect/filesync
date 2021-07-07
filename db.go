package main

import (
	"os"
	"path/filepath"
	"runtime"
	"sync"

	"github.com/ddirect/check"
	"github.com/ddirect/filemeta"
	"github.com/ddirect/filesync/records"
	"github.com/ddirect/protostream"
)

type HashKey [filemeta.HashSize]byte

func toHashKey(x []byte) (k HashKey) {
	copy(k[:], x)
	return
}

type Db struct {
	Dirs        []*Dir
	Files       []*File
	DirsByPath  map[string]*Dir
	FilesByHash map[HashKey]*File
	FilesByPath map[string]*File
}

func ReadDb(basePath string) *Db {
	workers := runtime.NumCPU()
	const queueBuf = 4000
	queue1 := make(chan *File, queueBuf)
	queue2 := make(chan *File, queueBuf)
	var wg1, wg2 sync.WaitGroup
	wg1.Add(workers)
	for i := 0; i < workers; i++ {
		go func() {
			for file := range queue1 {
				data, err := filemeta.Get(filepath.Join(basePath, file.Path))
				check.E(err)
				if attr := data.Attr; attr != nil {
					file.Hash = attr.Hash
					file.TimeNs = attr.TimeNs
					file.Size = attr.Size
					queue2 <- file
				}
			}
			wg1.Done()
		}()
	}

	db := &Db{FilesByHash: make(map[HashKey]*File), FilesByPath: make(map[string]*File)}

	wg2.Add(1)
	go func() {
		for file := range queue2 {
			db.FilesByHash[toHashKey(file.Hash)] = file
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
		db.Dirs = make([]Dir, r.DirCount)
		db.Files = make([]*File, r.FileCount)
	}
}

func (db *Db) codec(headerCodec func(*Db), dirCodec func(*Dir), fileCodec func(**File)) {
	headerCodec(db)
	dirs := db.Dirs
	for dirI := range dirs {
		dirCodec(&dirs[dirI])
	}
	files := db.Files
	for fileI := range files {
		fileCodec(&files[fileI])
	}
}

func (db *Db) Send(ps protostream.ReadWriter) {
	db.codec(DbHeaderSender(ps), DirRecordSender(ps), FileRecordSender(ps))
}

func RecvDb(ps protostream.ReadWriter) *Db {
	db := new(Db)
	db.codec(DbHeaderReceiver(ps), DirRecordReceiver(ps), FileRecordReceiver(ps, db))
	return db
}
