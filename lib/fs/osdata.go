package fs

import (
	"github.com/syncthing/syncthing/lib/protocol"
)

type OSDataGetter interface {
	// SetOSData sets the operating system private data for the current
	// operating system onto the destination FileInfo, and leaves privata
	// data for other OSes untouched.
	GetOSData(cur *protocol.FileInfo, stat FileInfo) (map[protocol.OS][]byte, error)
}
