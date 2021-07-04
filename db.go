package main

import (
	"os"
	"path/filepath"
	"runtime"
	"sync"

	"github.com/ddirect/filesync/records"
	"google.golang.org/protobuf/proto"

	//"google.golang.org/protobuf/proto"

	"github.com/ddirect/check"
	"github.com/ddirect/filemeta"
)

type HashKey [filemeta.HashSize]byte

func toHashKey(x []byte) (k HashKey) {
	copy(k[:], x)
	return
}

type Dir struct {
	Path   string
	TimeNs int64
}

func DirRecordBuilder() func(*Dir) proto.Message {
	r := new(records.Dir)
	return func(d *Dir) proto.Message {
		r.Path = d.Path
		r.TimeNs = d.TimeNs
		return r
	}
}

type File struct {
	Path     string
	Name     string
	DirIndex int
	Hash     []byte
	TimeNs   int64
	Size     int64
}

func FileRecordBuilder() func(*File) proto.Message {
	r := new(records.File)
	return func(d *File) proto.Message {
		r.Name = d.Name
		r.DirIndex = int64(d.DirIndex)
		r.Hash = d.Hash
		r.TimeNs = d.TimeNs
		r.Size = d.Size
		return r
	}
}

type Db struct {
	Dirs   []Dir
	Files  []*File
	ByHash map[HashKey]*File
	ByPath map[string]*File
}

func DbHeaderRecordBuilder() func(*Db) proto.Message {
	r := new(records.DbHeader)
	return func(db *Db) proto.Message {
		r.DirCount = int64(len(db.Dirs))
		r.FileCount = int64(len(db.Files))
		return r
	}
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

	db := &Db{ByHash: make(map[HashKey]*File), ByPath: make(map[string]*File)}

	wg2.Add(1)
	go func() {
		for file := range queue2 {
			db.ByHash[toHashKey(file.Hash)] = file
		}
		wg2.Done()
	}()

	Walk(basePath, func(relPath string) int {
		stat, err := os.Stat(filepath.Join(basePath, relPath)) // TODO: or Lstat?
		check.E(err)
		db.Dirs = append(db.Dirs, Dir{relPath, stat.ModTime().UnixNano()})
		return len(db.Dirs) - 1
	}, func(relPath string, name string, dirId int) {
		file := &File{Path: relPath, Name: name, DirIndex: dirId}
		db.Files = append(db.Files, file)
		db.ByPath[relPath] = file
		queue1 <- file
	})

	close(queue1)
	wg1.Wait()
	close(queue2)
	wg2.Wait()

	return db
}
