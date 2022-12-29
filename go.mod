module github.com/ddirect/filesync

replace github.com/ddirect/filemeta => ../filemeta

replace github.com/ddirect/format => ../format

replace github.com/ddirect/check => ../check

replace github.com/ddirect/protostream => ../protostream

replace github.com/ddirect/sys => ../sys

replace github.com/ddirect/filetest => ../filetest

replace github.com/ddirect/xrand => ../xrand

go 1.18

require (
	github.com/ddirect/check v0.0.0-00010101000000-000000000000
	github.com/ddirect/filemeta v0.0.0-00010101000000-000000000000
	github.com/ddirect/filetest v0.0.0-00010101000000-000000000000
	github.com/ddirect/format v0.0.0-00010101000000-000000000000
	github.com/ddirect/protostream v0.0.0-00010101000000-000000000000
	github.com/ddirect/sys v0.0.0-00010101000000-000000000000
	github.com/ddirect/xrand v0.0.0-00010101000000-000000000000
	google.golang.org/protobuf v1.27.1
)

require (
	github.com/google/go-cmp v0.5.6 // indirect
	golang.org/x/crypto v0.0.0-20210921155107-089bfa567519 // indirect
	golang.org/x/exp v0.0.0-20220205015713-f5f519d967d6 // indirect
	golang.org/x/sys v0.0.0-20211019181941-9d821ace8654 // indirect
)
