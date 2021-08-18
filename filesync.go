package main

import (
	"flag"
	"fmt"
	"os"
)

type NetAddr func() (string, string)

func netAddr(address string, unix bool) NetAddr {
	network := "tcp"
	if unix {
		network = "unix"
	}
	return func() (string, string) {
		return network, address
	}
}

func main() {
	var do string
	var basePath string
	var remoteAddress string
	var bindAddress string
	var nocache bool
	var unix bool
	flag.StringVar(&do, "do", "", "serve|diff|recv")
	flag.StringVar(&basePath, "base", "", "local path")
	flag.StringVar(&remoteAddress, "remote", "", "remote address")
	flag.StringVar(&bindAddress, "bind", "", "bind address")
	flag.BoolVar(&nocache, "nocache", false, "disable directory tree caching")
	flag.BoolVar(&unix, "unix", false, "use unix sockets instead of TCP")
	flag.Parse()

	if do == "" {
		flag.Usage()
		return
	}

	switch do {
	case "serve":
		Serve(ReadDb(basePath, !nocache), basePath, netAddr(bindAddress, unix))
	case "diff":
		Sync(ReadDb(basePath, !nocache), basePath, netAddr(remoteAddress, unix), DiffActionsFactory)
	case "recv":
		Sync(ReadDb(basePath, !nocache), basePath, netAddr(remoteAddress, unix), RecvActionsFactory)
	default:
		fmt.Fprintf(os.Stderr, "unknown operation '%s'\n", do)
	}
}


/*
Issues:
- sour and dest on-disk and global db size would be useful (but filemetatool does this already)
- test synching with missing metadata (tests run with filemeta.Refresh at the moment)
- separate deleted links from actual deletions (inferred by "no more links to file")
*/
