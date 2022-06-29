package fs

import (
	"github.com/syncthing/syncthing/lib/protocol"
)

type OSDataSetter interface {
	// SetOSData sets the operating system private data for the current
	// operating system onto the destination FileInfo, and leaves privata
	// data for other OSes untouched.
	SetOSData(dst *protocol.FileInfo, fi FileInfo) error
}