// Package statexp is an experimental API for the gokrazy/stat package.
package statexp

import (
	"github.com/gokrazy/stat"
	"github.com/gokrazy/stat/internal/cpu"
	"github.com/gokrazy/stat/internal/disk"
	"github.com/gokrazy/stat/internal/mem"
	"github.com/gokrazy/stat/internal/net"
	"github.com/gokrazy/stat/internal/sys"
)

type ProcessAndFormatter interface {
	ProcessAndFormat(map[string][]byte) []stat.Col
}

func DefaultModules() []ProcessAndFormatter {
	return []ProcessAndFormatter{
		&cpu.Stats{},
		&disk.Stats{},
		&sys.Stats{},
		&net.Stats{},
		&mem.Stats{},
	}
}
