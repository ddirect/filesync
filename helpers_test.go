package main

import (
	"fmt"

	ft "github.com/ddirect/filetest"
	"github.com/ddirect/xrand"
)

type errorfunc func(string, ...interface{})

func createTree(rnd xrand.Xrand) *ft.Dir {
	entryFactory := ft.NewEntryFactory(ft.NewRandomNameFactory(rnd, rnd.UniformFactory(15, 20), ft.LowerCaseChars))
	fileFactory := ft.NewFileFactory(entryFactory)
	filesFactory := ft.NewFilesFactory(fileFactory, rnd.UniformFactory(0, 10))

	var dirsFactory ft.DirsFactory
	dirFactory := ft.NewDirFactory(entryFactory, filesFactory, ft.FutureDirsFactory(3, &dirsFactory))
	dirsFactory = ft.NewDirsFactory(dirFactory, rnd.UniformFactory(1, 10))
	return ft.NewDirFactory(ft.NullEntryFactory(), filesFactory, dirsFactory)("", 0)
	//return dirFactory("", 0)
}

type dirStats struct {
	uniqueFiles  int
	uniqueHashes int
}

func commit1(tree *ft.Dir, rnd xrand.Xrand, root string) dirStats {
	tree.EachDirRecursive(ft.NewDirMakerFactory(root))
	files := tree.AllFilesSlice()
	rnd.Shuffle(len(files), func(i, j int) {
		files[i], files[j] = files[j], files[i]
	})
	createLimit := 30 * len(files) / 100
	cloneLimit := 60 * len(files) / 100
	linkLimit := len(files)

	create := ft.NewRandomFileFactory(rnd, root, rnd.UniformFactory(0, 5000))
	clone := ft.NewCloneFileOperation(root, root)
	link := ft.NewLinkFileOperation(root, root)

	i := 0
	for ; i < createLimit; i++ {
		create(files[i])
	}
	for ; i < cloneLimit; i++ {
		clone(files[rnd.Intn(createLimit)], files[i])
	}
	for ; i < linkLimit; i++ {
		link(files[rnd.Intn(cloneLimit)], files[i])
	}
	return dirStats{cloneLimit, createLimit}
}

func show(f *ft.File) {
	fmt.Println(f.Path())
}
