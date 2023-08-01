package net

import (
	"strings"

	"github.com/gokrazy/stat"
	"github.com/gokrazy/stat/internal/must"
)

type reading struct {
	recv uint64
	send uint64
}

type Stats struct {
	old, cur reading
}

func (s *Stats) Headers() []string {
	return []string{" recv", " send"}
}

func (s *Stats) FileContents() []string {
	return []string{"/proc/net/dev"}
}

func (s *Stats) process(contents map[string][]byte) {
	s.old = s.cur
	s.cur = reading{}

	lines := strings.Split(string(contents["/proc/net/dev"]), "\n")
	if len(lines) == 0 {
		return
	}
	var totalRecv uint64
	var totalSend uint64
	for _, line := range lines {
		line = strings.TrimSpace(strings.ReplaceAll(line, ":", ""))

		// As per proc(5):
		f := strings.Fields(line)
		if len(f) < 10 {
			continue
		}
		recv := must.Uint64(f[1])
		send := must.Uint64(f[9])
		totalRecv += recv
		totalSend += send
	}

	s.cur.recv = totalRecv
	s.cur.send = totalSend
}

func (s *Stats) ProcessAndFormat(contents map[string][]byte) []stat.Col {
	s.process(contents)
	return []stat.Col{
		stat.ByteCol(s.cur.recv - s.old.recv).WithWidth(5),
		stat.ByteCol(s.cur.send - s.old.send).WithWidth(5),
	}
}
