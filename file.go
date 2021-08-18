package main

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/ddirect/filemeta"

	"github.com/ddirect/check"
	"github.com/ddirect/filesync/records"
	"github.com/ddirect/protostream"
)

type File struct {
	Path     string
	Name     string
	DirIndex int
	Hash     []byte
	TimeNs   int64
	Size     int64
}

func FileRecordSender(ps protostream.ReadWriter) func(*File) {
	r := new(records.File)
	return func(f *File) {
		r.Name = f.Name
		r.DirIndex = int64(f.DirIndex)
		r.Hash = f.Hash
		r.TimeNs = f.TimeNs
		r.Size = f.Size
		check.E(ps.WriteMessage(r))
	}
}

func FileRecordReceiver(ps protostream.ReadWriter, db *Db) func(*File) {
	r := new(records.File)
	return func(f *File) {
		check.E(ps.ReadMessage(r))
		f.Name = r.Name
		di := int(r.DirIndex)
		f.DirIndex = di
		f.Hash = r.Hash
		f.TimeNs = r.TimeNs
		f.Size = r.Size

		f.Path = filepath.Join(db.Dirs[di].Path, f.Name)
		db.FilesByPath[f.Path] = f
		db.FilesByHash[filemeta.ToHashKey(f.Hash)] = f
	}
}

func FileDataSender(ps protostream.ReadWriter, db *Db, basePath string) func([]byte) {
	return func(hash []byte) {
		f := db.FilesByHash[filemeta.ToHashKey(hash)]
		if f == nil {
			panic(fmt.Errorf("file hash '%s' not found in db", hex.EncodeToString(hash)))
		}
		file, err := os.Open(filepath.Join(basePath, f.Path))
		check.E(err)
		defer check.DeferredE(file.Close)
		check.E(ps.WriteStream(file.Read))
	}
}

func FileDataReceiver(ps protostream.ReadWriter, basePath string) func(*File) {
	fs := filemeta.NewFileWriter()
	return func(f *File) {
		check.E(fs.Open(filepath.Join(basePath, f.Path), os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0664))
		defer check.DeferredE(func() error {
			attr, err := fs.Close(f.TimeNs)
			if err != nil {
				return err
			}
			if f.Size != attr.Size {
				return fmt.Errorf("received %d instead of %d", attr.Size, f.Size)
			}
			if bytes.Compare(f.Hash, attr.Hash) != 0 {
				return errors.New("hash mismatch on received file")
			}
			return nil
		})
		check.E(ps.ReadStream(fs.Write))
	}
}
