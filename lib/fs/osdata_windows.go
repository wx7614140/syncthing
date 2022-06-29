package fs

import (
	"fmt"
	"os/user"

	"github.com/syncthing/syncthing/lib/protocol"
	"golang.org/x/sys/windows"
)

func NewOSDataGetter(underlying Filesystem) OSDataGetter {
	return &WindowsOSDataGetter{fs: underlying}
}

type WindowsOSDataGetter struct {
	fs Filesystem
}

func (p *WindowsOSDataGetter) GetOSData(cur *protocol.FileInfo, stat FileInfo) (map[protocol.OS][]byte, error) {
	// The underlying filesystem must be a BasicFilesystem.
	basic, ok := p.fs.(*BasicFilesystem)
	if !ok {
		return nil, fmt.Errorf("underlying filesystem is not a BasicFilesystem")
	}

	rootedName, err := basic.rooted(cur.Name)
	if err != nil {
		return nil, err
	}
	hdl, err := windows.Open(rootedName, windows.O_RDONLY, 0)
	if err != nil {
		return nil, err
	}
	defer windows.Close(hdl)

	sd, err := windows.GetSecurityInfo(hdl, windows.SE_FILE_OBJECT, windows.DACL_SECURITY_INFORMATION)
	if err != nil {
		return nil, err
	}
	owner, _, err := sd.Owner()
	if err != nil {
		return nil, err
	}

	pd := &protocol.WindowsOSData{
		OwnerSID: owner.String(),
	}

	user, err := user.LookupId(owner.String())
	if err == nil {
		pd.OwnerName = user.Username
	}

	bs, err := pd.Marshal()
	if err != nil {
		return nil, fmt.Errorf("surprising error marshalling private data: %w", err)
	}
	return map[protocol.OS][]byte{
		protocol.OsWindows: bs,
	}, nil
}
