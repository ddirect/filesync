# filesync
`filesync` synchronizes a `source` tree of files with a `destination` tree; the trees typically reside on separate systems. It performs copies, moves and deletions on the `destination` tree so that it matches the `source` tree.

IMPORTANT: the two trees must be scanned in advance with `filemetatool`, to ensure that all file hashes are up to date (see https://github.com/ddirect/filemetatool for more info). 

## Requirements
`filesync` uses the `filemeta` library which currently requires Linux and file systems with extended attributes (e.g. EXT4).

## Synopsis
`filesync -do OPERATION [OPTION]... -base FILE_TREE`

## How it works
`filesync` is run in server mode on the source system and as client on the destination one.

In both modes, `filesync` first scans the tree and creates an in-memory data structure (called "db") containing information about the tree of files, including all file hashes.

In server mode (`serv` operation), it then listens and serves connections (TCP or Unix sockets) coming from client instances.

In client mode, it connects to a `filesync` server, requests the remote db and computes the operations which would be needed to make the local file system match the remote one. It then either shows the operations (`diff` or `fulldiff` operations) or it performs them right away (`recv` operation).

`filesync` attempts to avoid copying the file data from the remote system: if a file with the same hash exists locally, a hard link to it is created instead of performing a copy. Similarly, if two files with the same hash exist on the remote system but not on the local one, a single copy is performed and a hard link to the first file is created for the second one.

## Scenarios
- **deduplication:** if the destination tree is empty, `filesync` copies the source tree deduplicating any file (it creates multiple hard links to the same inode for files containing the same data)
- **synchronization:** if the destination tree contains a backup of the files of the source tree taken at an earlier time, `filesync` efficiently updates the backup minimizing the number of needed operations

## Examples
### Server side
```
filesync -bind :6001 -do serve -base <source_tree>
```
Starts `filesync` in server mode on TCP port 6001, for `source_tree`.

### Client side
```
filesync -do recv -remote <remote_host>:6001 -base <destination_tree>
```
Starts `filesync` in client mode and synchronizes the tree on `remote_host`, port 6001 onto `destination_tree`.
