// Copyright (C) 2022 The Syncthing Authors.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this file,
// You can obtain one at https://mozilla.org/MPL/2.0/.

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
