package fs

import (
	"github.com/syncthing/syncthing/lib/protocol"
)

type WindowsOSDataSetter struct {
}

func (p *WindowsOSDataSetter) SetOSData(dst *protocol.FileInfo, fi FileInfo) error {
	return nil
}
