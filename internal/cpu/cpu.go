package cpu

import (
	"strings"

	"github.com/gokrazy/stat"
	"github.com/gokrazy/stat/internal/must"
)

type reading struct {
	// TODO: verify in the kernel source the data type for these
	usr uint64 // user + nice + irq + softirq
	sys uint64 // system
	idl uint64 // idle
	wai uint64 // I/O wait
	stl uint64 // steal

	sum uint64
}

type Stats struct {
	old, cur reading
}

func (s *Stats) Headers() []string {
	return []string{"usr", "sys", "idl", "wai", "stl"}
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

	// As per proc(5):
	f := strings.Fields(strings.TrimSpace(lines[0]))
	user := must.Uint64(f[1])
	nice := must.Uint64(f[2])
	system := must.Uint64(f[3])
	idle := must.Uint64(f[4])
	iowait := must.Uint64(f[5])
	irq := must.Uint64(f[6])
	softirq := must.Uint64(f[7])
	steal := must.Uint64(f[8])

	s.cur.usr = user + nice + irq + softirq
	s.cur.sys = system
	s.cur.idl = idle
	s.cur.wai = iowait
	s.cur.stl = steal
	s.cur.sum = s.cur.usr + s.cur.sys + s.cur.idl + s.cur.wai + s.cur.stl
}

func (s *Stats) ProcessAndFormat(contents map[string][]byte) []stat.Col {
	s.process(contents)
	total := float64(s.cur.sum - s.old.sum)
	return []stat.Col{
		{Type: stat.ColPercentage, ValFloat64: 100 * float64(s.cur.usr-s.old.usr) / total, Width: 3, Scale: 34},
		{Type: stat.ColPercentage, ValFloat64: 100 * float64(s.cur.sys-s.old.sys) / total, Width: 3, Scale: 34},
		{Type: stat.ColPercentage, ValFloat64: 100 * float64(s.cur.idl-s.old.idl) / total, Width: 3, Scale: 34},
		{Type: stat.ColPercentage, ValFloat64: 100 * float64(s.cur.wai-s.old.wai) / total, Width: 3, Scale: 34},
		{Type: stat.ColPercentage, ValFloat64: 100 * float64(s.cur.stl-s.old.stl) / total, Width: 3, Scale: 34},
	}
}
