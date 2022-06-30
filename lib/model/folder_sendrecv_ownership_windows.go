package model

import (
	"os/user"
	"strconv"

	"github.com/syncthing/syncthing/lib/protocol"
)

func (f *sendReceiveFolder) syncOwnership(file *protocol.FileInfo, path string) error {
	var pd protocol.WindowsOSData
	if !file.LoadOSData(protocol.OsWindows, &pd) {
		// No owner data, nothing to do
		return nil
	}

	return nil
}
