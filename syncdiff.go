package main

import (
	"fmt"

	"github.com/ddirect/format"
	"github.com/ddirect/protostream"
)

type stats struct {
	count int
	size  int64
}

func (s *stats) update(file *File) {
	s.count++
	s.size += file.Size
}

func (s *stats) appendRow(t *format.Table, name string) {
	if s.count > 0 {
		t.AppendRow(name, s.count, format.Size(s.size))
	}
}

func DiffActionsFactory(sdb *Db, ddb *Db, basePath string, ps protostream.ReadWriter) SyncActions {
	sdirCount := len(sdb.Dirs)
	sfileCount := len(sdb.Files)
	ddirCount := len(ddb.Dirs)
	dfileCount := len(ddb.Files)
	var createdDirs, removedDirs int
	var copied, linked, stashed, removed stats
	return &syncActions{
		CreateDir: func(*Dir) {
			createdDirs++
		},
		RemoveDir: func(*Dir) {
			removedDirs++
		},
		RemoveFile: func(f *File) {
			removed.update(f)
		},
		CopyFile: func(f *File) {
			copied.update(f)
		},
		LinkFile: func(f *File, _ *File) {
			linked.update(f)
		},
		StashFile: func(f *File) {
			stashed.update(f)
		},
		Epilogue: func() {
			var s, d, f format.Table
			s.AppendRow(".", "remote", "local")
			s.AppendRow("total dirs", sdirCount, ddirCount)
			s.AppendRow("total files", sfileCount, dfileCount)
			d.AppendRow("dirs", "count")
			d.AppendRow("created", createdDirs)
			d.AppendRow("removed", removedDirs)
			f.AppendRow("files", "count", "size")
			copied.appendRow(&f, "copied")
			linked.appendRow(&f, "linked")
			removed.appendRow(&f, "removed")
			stashed.appendRow(&f, "stashed")
			fmt.Printf("%s\n%s\n%s", &s, &d, &f)
		},
	}
}
