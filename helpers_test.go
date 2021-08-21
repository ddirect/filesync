package main

import (
	"fmt"

	ft "github.com/ddirect/filetest"
)

type errorfunc func(string, ...interface{})

func treeOptions() ft.TreeOptions {
	o := ft.DefaultTreeOptions()
	o.FileCount = ft.MinMax{0, 10}
	o.DirCount = ft.MinMax{1, 10}
	o.Depth = 3
	return o
}

func flatTreeOptions() ft.TreeOptions {
	o := ft.DefaultTreeOptions()
	o.FileCount = ft.MinMax{100, 100}
	o.Depth = -1
	return o
}

func show(f *ft.File) {
	fmt.Println(f.Path())
}
