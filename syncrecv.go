package main

import (
	"encoding/hex"
	"os"
	"path/filepath"

	"github.com/ddirect/xrand"

	"github.com/ddirect/check"
	"github.com/ddirect/protostream"
)

func randomNameFactory() func() string {
	rnd := xrand.New()
	rawName := make([]byte, 32)
	encodedName := make([]byte, hex.EncodedLen(len(rawName)))
	return func() string {
		rnd.Fill(rawName)
		hex.Encode(encodedName, rawName)
		return string(encodedName)
	}
}

func RecvActionsFactory(sdb *Db, ddb *Db, basePath string, ps protostream.ReadWriter) SyncActions {
	da := DiffActionsFactory(sdb, ddb, basePath, ps)
	sendGetFile := GetFileCommandSender(ps)
	recvFileData := FileDataReceiver(ps, basePath)
	randomName := randomNameFactory()
	fullPath := func(name string) string {
		return filepath.Join(basePath, name)
	}
	return &syncActions{
		CreateDir: func(d *Dir) {
			check.E(os.Mkdir(fullPath(d.Path), 0775))
			da.CreateDir(d)
		},
		RemoveDir: func(d *Dir) {
			check.E(os.Remove(fullPath(d.Path)))
			da.RemoveDir(d)
		},
		RemoveFile: func(f *File) {
			check.E(os.Remove(fullPath(f.Path)))
			da.RemoveFile(f)
		},
		CopyFile: func(f *File) {
			sendGetFile(f.Hash)
			recvFileData(f)
			ddb.FilesByPath[f.Path] = f
			ddb.FilesByHash[toHashKey(f.Hash)] = f
			da.CopyFile(f)
		},
		LinkFile: func(sf *File, df *File) {
			check.E(os.Link(fullPath(sf.Path), fullPath(df.Path)))
			ddb.FilesByPath[df.Path] = df
			da.LinkFile(sf, df)
		},
		StashFile: func(f *File) {
			path := fullPath(randomName())
			check.E(os.Rename(fullPath(f.Path), path))
			delete(ddb.FilesByPath, f.Path)
			f.Path = path
			ddb.FilesByPath[path] = f
			da.StashFile(f)
		},
		Epilogue: func() {
			da.Epilogue()
		},
	}
}
