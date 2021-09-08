package main

import (
	"bufio"
	"fmt"
	"os"

	"github.com/ddirect/check"

	"github.com/ddirect/protostream"
)

func FullDiffActionsFactory(sdb *Db, ddb *Db, basePath string, ps protostream.ReadWriter) SyncActions {
	var createdDirs, removedDirs, copied, linked, stashed, removed []string
	return &syncActions{
		CreateDir: func(d *Dir) {
			createdDirs = append(createdDirs, d.Path)
		},
		RemoveDir: func(d *Dir) {
			removedDirs = append(removedDirs, d.Path)
		},
		RemoveFile: func(f *File) {
			removed = append(removed, f.Path)
		},
		CopyFile: func(f *File) {
			copied = append(copied, f.Path)
		},
		LinkFile: func(s *File, d *File) {
			linked = append(linked, fmt.Sprintf("%s -> %s", s.Path, d.Path))
		},
		StashFile: func(f *File) {
			stashed = append(stashed, f.Path)
		},
		Epilogue: func() {
			w := bufio.NewWriter(os.Stdout)
			dumpln := func() {
				check.E(w.WriteByte('\n'))
			}
			dump1 := func(x string) {
				check.IE(w.WriteString(x))
			}
			dump1ln := func(x string) {
				dump1(x)
				dumpln()
			}
			dump := func(name string, slice []string) {
				dump1(" -------- ")
				dump1(name)
				dump1ln(" --------")
				for _, x := range slice {
					dump1ln(x)
				}
				dumpln()
			}
			dump("copied", copied)
			dump("linked", linked)
			dump("stashed", stashed)
			dump("removed", removed)
			dump("created dirs", createdDirs)
			dump("removed dirs", removedDirs)
			check.E(w.Flush())
		},
	}
}
