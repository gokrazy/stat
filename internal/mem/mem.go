package mem

import (
	"strings"

	"github.com/gokrazy/stat"
	"github.com/gokrazy/stat/internal/must"
)

type reading struct {
	used uint64
	free uint64
	buff uint64
	cach uint64
}

type Stats struct {
	old, cur reading
}

func (s *Stats) Headers() []string {
	return []string{" used", " free", " buff", " cach"}
}

func (s *Stats) FileContents() []string {
	return []string{"/proc/meminfo"}
}

func (s *Stats) process(contents map[string][]byte) {
	s.old = s.cur
	s.cur = reading{}

	lines := strings.Split(string(contents["/proc/meminfo"]), "\n")
	if len(lines) == 0 {
		return
	}
	var (
		memTotal     uint64
		shmem        uint64
		sreclaimable uint64
	)
	for _, line := range lines {
		line = strings.TrimSpace(strings.ReplaceAll(line, ":", ""))

		// As per proc(5):
		f := strings.Fields(line)
		if len(f) < 2 {
			continue
		}
		switch f[0] {
		case "MemFree":
			s.cur.free = must.Uint64(f[1]) * 1024
		case "Buffers":
			s.cur.buff = must.Uint64(f[1]) * 1024
		case "Cached":
			s.cur.cach = must.Uint64(f[1]) * 1024
		case "MemTotal":
			memTotal = must.Uint64(f[1]) * 1024
		case "Shmem":
			shmem = must.Uint64(f[1]) * 1024
		case "SReclaimable":
			sreclaimable = must.Uint64(f[1]) * 1024
		}
	}

	s.cur.used = memTotal - s.cur.free - s.cur.buff - s.cur.cach - sreclaimable + shmem
}

func (s *Stats) ProcessAndFormat(contents map[string][]byte) []stat.Col {
	s.process(contents)
	return []stat.Col{
		{Type: stat.ColGauge, Unit: stat.UnitBytesFloat, ValFloat64: float64(s.cur.used), Width: 5},
		{Type: stat.ColGauge, Unit: stat.UnitBytesFloat, ValFloat64: float64(s.cur.free), Width: 5},
		{Type: stat.ColGauge, Unit: stat.UnitBytesFloat, ValFloat64: float64(s.cur.buff), Width: 5},
		{Type: stat.ColGauge, Unit: stat.UnitBytesFloat, ValFloat64: float64(s.cur.cach), Width: 5},
	}
}
