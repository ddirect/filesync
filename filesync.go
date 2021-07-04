package main

import (
	"flag"
	"fmt"
	"net/url"
	"os"

	"github.com/ddirect/check"
)

type NetAddr func() (string, string)

func splitNetAddr(x string) NetAddr {
	u, err := url.Parse(x)
	check.E(err)
	return func() (string, string) {
		return u.Scheme, u.Host
	}
}

func main() {
	var do string
	var basePath string
	var remoteAddress string
	var bindAddress string
	flag.StringVar(&do, "do", "", "serve|recv")
	flag.StringVar(&basePath, "base", "", "local path")
	flag.StringVar(&remoteAddress, "remote", "", "remote address")
	flag.StringVar(&bindAddress, "bind", "", "bind address")
	flag.Parse()

	if do == "" {
		flag.Usage()
		return
	}

	switch do {
	case "serve":
		serve(ReadDb(basePath), basePath, splitNetAddr(bindAddress))
	case "recv":
		recv(ReadDb(basePath), basePath, splitNetAddr(remoteAddress))
	default:
		fmt.Fprintf(os.Stderr, "unknown operation '%s'\n", do)
	}
}
