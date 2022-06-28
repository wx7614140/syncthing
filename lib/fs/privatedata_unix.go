package fs

import (
	"fmt"
	"io/fs"
	"os/user"
	"sync"
	"syscall"
	"time"

	"github.com/syncthing/syncthing/lib/protocol"
)

const (
	positiveNameCachetime = time.Hour
	negativeNameCachetime = 5 * time.Minute
)

func NewOSDataSetter() OSDataSetter {
	return &POSIXOSDataSetter{
		users:  newNameCache(),
		groups: newNameCache(),
	}
}

type POSIXOSDataSetter struct {
	users  *nameCache
	groups *nameCache
}

func (p *POSIXOSDataSetter) SetOSData(dst *protocol.FileInfo, fi fs.FileInfo) error {
	sys, ok := fi.Sys().(*syscall.Stat_t)
	if !ok {
		return fmt.Errorf("unexpected underlying stat type: %T", sys)
	}

	owner, err := p.users.getOrPopulate(sys.Uid, func(uid uint32) (string, error) {
		user, err := user.LookupId(fmt.Sprintf("%d", uid))
		if err != nil {
			return "", err
		}
		return user.Username, nil
	})

	group, err := p.users.getOrPopulate(sys.Uid, func(gid uint32) (string, error) {
		group, err := user.LookupGroupId(fmt.Sprintf("%d", gid))
		if err != nil {
			return "", err
		}
		return group.Name, nil
	})

	pd := &protocol.POSIXPrivateData{
		OwnerUID:  sys.Uid,
		OwnerName: owner,
		GroupID:   sys.Gid,
		GroupName: group,
	}
	bs, err := pd.Marshal()
	if err != nil {
		return fmt.Errorf("surprising error marshalling private data: %w", err)
	}

	dst.OsPrivateData[protocol.OsPosix] = bs
	return nil
}

type nameCache struct {
	names map[uint32]nameCacheEntry
	mut   sync.RWMutex
}

func newNameCache() *nameCache {
	return &nameCache{
		names: make(map[uint32]nameCacheEntry),
	}
}

type nameCacheEntry struct {
	name string
	err  error
	when time.Time
}

func (cache *nameCache) getOrPopulate(id uint32, get func(uint32) (string, error)) (string, error) {
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
