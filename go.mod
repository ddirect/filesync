module github.com/ddirect/filesync

replace github.com/ddirect/filemeta => ../filemeta

replace github.com/ddirect/format => ../format

replace github.com/ddirect/check => ../check

replace github.com/ddirect/protostream => ../protostream

replace github.com/ddirect/sys => ../sys

go 1.16

require (
	github.com/ddirect/check v0.0.0-00010101000000-000000000000
	github.com/ddirect/filemeta v0.0.0-00010101000000-000000000000
	github.com/ddirect/format v0.0.0-00010101000000-000000000000
	github.com/ddirect/protostream v0.0.0-00010101000000-000000000000
	github.com/ddirect/sys v0.0.0-00010101000000-000000000000
	google.golang.org/protobuf v1.27.1
)
