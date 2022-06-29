package fs

import (
	"github.com/syncthing/syncthing/lib/protocol"
	"golang.org/x/sys/windows"
)

func NewOSDataSetter() OSDataSetter {
	return &WindowsOSDataSetter{}
}

type WindowsOSDataSetter struct {
}

func (p *WindowsOSDataSetter) SetOSData(dst *protocol.FileInfo, fi FileInfo) error {
	hdl, err := windows.Open(dst.Name, windows.O_RDONLY, 0)
	if err != nil {
		return err
	}
	defer windows.Close(hdl)

	sd, err := windows.GetSecurityInfo(hdl, windows.SE_FILE_OBJECT, windows.DACL_SECURITY_INFORMATION)
	if err != nil {
		return err
	}
	owner, _, err := sd.Owner()
	if err != nil {
		return err
	}

	pd := &protocol.WindowsPrivateData{
		OwnerSid: owner.String(),
	}
	bs, _ := pd.Marshal()
	if dst.OsPrivateData == nil {
		dst.OsPrivateData = make(map[protocol.OS][]byte)
	}
	dst.OsPrivateData[protocol.OsWindows] = bs

	return nil
}
