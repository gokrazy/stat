package sys

import (
	"strings"

	"github.com/gokrazy/stat"
	"github.com/gokrazy/stat/internal/must"
)

type reading struct {
	// TODO: verify in the kernel source the data type for these
	int uint64
	csw uint64
}

type Stats struct {
	old, cur reading
}

func (s *Stats) FileContents() []string {
	return []string{"/proc/stat"}
}

func (s *Stats) process(contents map[string][]byte) {
	s.old = s.cur
	s.cur = reading{}

	lines := strings.Split(string(contents["/proc/stat"]), "\n")
	if len(lines) == 0 {
		return
	}

	for _, line := range lines {
		f := strings.Fields(line)
		if len(f) < 2 {
			continue
		}
		if f[0] == "intr" {
			s.cur.int = must.Uint64(f[1])
		}
		if f[0] == "ctxt" {
			s.cur.csw = must.Uint64(f[1])
		}
	}
}

func (s *Stats) ProcessAndFormat(contents map[string][]byte) []stat.Col {
	s.process(contents)
	return []stat.Col{
		stat.MetricCol(s.cur.int - s.old.int).WithWidth(5),
		stat.MetricCol(s.cur.csw - s.old.csw).WithWidth(5),
	}
}
