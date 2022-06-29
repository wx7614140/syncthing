//go:build !windows
// +build !windows

package fs

func NewOSDataGetter(underlying Filesystem) OSDataGetter {
	return NewPOSIXDataGetter(underlying)
}
