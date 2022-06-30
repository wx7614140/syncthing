package model

import (
	"os/user"
	"strings"

	"github.com/syncthing/syncthing/lib/protocol"
)

func (f *sendReceiveFolder) syncOwnership(file *protocol.FileInfo, path string) error {
	var pd protocol.WindowsOSData
	if !file.LoadOSData(protocol.OsWindows, &pd) {
		// No owner data, nothing to do
		return nil
	}

	// Try to look up the user by name. The username will be something like
	// DOMAIN\user and we'll first try with the whole thing. If that fails,
	// we'll try again with just the user part without domain -- this
	// handled the case where the domain is just the local workstation name,
	// and it differs between two boxes. If that also fails, we'll just use
	// the SID and hope for the best.

	ownerSID := pd.OwnerSID
	if pd.OwnerName != "" {
		us, err := user.Lookup(pd.OwnerName)
		if err != nil {
			parts := strings.Split(pd.OwnerName, "\\")
			if len(parts) == 2 {
				us, err = user.Lookup(parts[1])
			}
		}
		if err == nil {
			ownerSID = us.Uid
		}
	}

	return f.mtimefs.Lchown(path, ownerSID, "")
}
