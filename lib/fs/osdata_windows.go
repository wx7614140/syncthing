// Copyright (C) 2022 The Syncthing Authors.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this file,
// You can obtain one at https://mozilla.org/MPL/2.0/.

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
		l.Infoln("WindowsOSDataGetter: underlying filesystem is not a BasicFilesystem")
		return nil, fmt.Errorf("underlying filesystem is not a BasicFilesystem")
	}

	rootedName, err := basic.rooted(cur.Name)
	if err != nil {
		l.Infoln("WindowsOSDataGetter: rooted failed:", err)
		return nil, err
	}
	hdl, err := windows.Open(rootedName, windows.O_RDONLY, 0)
	if err != nil {
		l.Infoln("Failed to open", rootedName, "for reading:", err)
		return nil, err
	}
	defer windows.Close(hdl)

	sd, err := windows.GetSecurityInfo(hdl, windows.SE_FILE_OBJECT, windows.OWNER_SECURITY_INFORMATION)
	if err != nil {
		l.Infoln("Failed to get security info for", rootedName, ":", err)
		return nil, err
	}
	owner, _, err := sd.Owner()
	if err != nil {
		l.Infoln("Failed to get owner for", rootedName, ":", err)
		return nil, err
	} else {
		l.Infoln("Owner for", rootedName, "is", owner)
	}

	pd := &protocol.WindowsOSData{}

	if us, err := user.LookupId(owner.String()); err == nil {
		l.Infoln("Found owner for", rootedName, ":", us.Username, us.Uid)
		pd.OwnerName = us.Username
	} else if gr, err := user.LookupGroupId(owner.String()); err == nil {
		l.Infoln("Found group for", rootedName, ":", gr.Name, gr.Gid)
		pd.OwnerName = gr.Name
		pd.OwnerIsGroup = true
	} else {
		l.Infoln("Failed to find owner for", rootedName, ":", err)
	}

	bs, err := pd.Marshal()
	if err != nil {
		return nil, fmt.Errorf("surprising error marshalling private data: %w", err)
	}

	l.Infoln("OS data for", rootedName, "is", pd)
	return map[protocol.OS][]byte{
		protocol.OsWindows: bs,
	}, nil
}
