package fs

import (
	"fmt"
	"os/user"
	"strconv"
	"sync"
	"time"

	"github.com/syncthing/syncthing/lib/protocol"
)

const (
	positiveNameCachetime = time.Hour
	negativeNameCachetime = 5 * time.Minute
)

func NewPOSIXDataGetter(_ Filesystem) OSDataGetter {
	return &POSIXOSDataGetter{
		users:  newNameCache(),
		groups: newNameCache(),
	}
}

type POSIXOSDataGetter struct {
	users  *nameCache
	groups *nameCache
}

func (p *POSIXOSDataGetter) GetOSData(_ *protocol.FileInfo, stat FileInfo) (map[protocol.OS][]byte, error) {
	// Look up owner and group ID and names via the cache. We swallow errors
	// in the name lookup, because the end result is anyway the empty
	// string.

	ownerUID := stat.Owner()
	ownerName, _ := p.users.getOrPopulate(ownerUID, func(uid int) (string, error) {
		user, err := user.LookupId(strconv.Itoa(uid))
		if err != nil {
			return "", err
		}
		return user.Username, nil
	})

	groupID := stat.Group()
	groupName, _ := p.users.getOrPopulate(groupID, func(gid int) (string, error) {
		group, err := user.LookupGroupId(strconv.Itoa(gid))
		if err != nil {
			return "", err
		}
		return group.Name, nil
	})

	// Create the POSIX private data structure and store it marshalled.
	pd := &protocol.POSIXOSData{
		OwnerUID:  ownerUID,
		OwnerName: ownerName,
		GroupID:   groupID,
		GroupName: groupName,
	}
	bs, err := pd.Marshal()
	if err != nil {
		return nil, fmt.Errorf("surprising error marshalling private data: %w", err)
	}
	return map[protocol.OS][]byte{
		protocol.OsPosix: bs,
	}, nil
}

type nameCache struct {
	names map[int]nameCacheEntry
	mut   sync.RWMutex
}

func newNameCache() *nameCache {
	return &nameCache{
		names: make(map[int]nameCacheEntry),
	}
}

type nameCacheEntry struct {
	name string
	err  error
	when time.Time
}

func (cache *nameCache) getOrPopulate(id int, get func(int) (string, error)) (string, error) {
	// In the best case we'll have a cache hit, optimize for that case by
	// taking a read lock only and for a short while.
	now := time.Now()
	cache.mut.RLock()
	entry, ok := cache.names[id]
	cache.mut.RUnlock()
	switch {
	case ok && entry.err == nil && now.Sub(entry.when) < positiveNameCachetime:
		return entry.name, nil
	case ok && entry.err != nil && now.Sub(entry.when) < negativeNameCachetime:
		return "", entry.err
	}

	// We didn't, so take a write lock and populate the cache.
	cache.mut.Lock()
	defer cache.mut.Unlock()
	// Someone might still have made this before we did, so check again
	// before doing the heavy work.
	entry, ok = cache.names[id]
	switch {
	case ok && entry.err == nil && now.Sub(entry.when) < positiveNameCachetime:
		return entry.name, nil
	case ok && entry.err != nil && now.Sub(entry.when) < negativeNameCachetime:
		return "", entry.err
	}

	name, err := get(id)
	if err != nil {
		cache.names[id] = nameCacheEntry{err: err, when: now}
		return "", err
	}
	cache.names[id] = nameCacheEntry{name: name, when: now}
	return name, nil
}
