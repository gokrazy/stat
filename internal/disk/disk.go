package disk

import (
	"regexp"
	"strings"

	"github.com/gokrazy/stat"
	"github.com/gokrazy/stat/internal/must"
)

type reading struct {
	read  uint64
	write uint64
}

type Stats struct {
	old, cur reading
}

func (s *Stats) FileContents() []string {
	return []string{"/proc/diskstats"}
}

var diskfilterRe = regexp.MustCompile(`^([hsv]d[a-z]+\d+|cciss/c\d+d\d+p\d+|dm-\d+|md\d+|loop\d+p\d+|nvme\d+n\d+p\d+|mmcblk\d+p\d+|VxVM\d+)$`)

func (s *Stats) process(contents map[string][]byte) {
	s.old = s.cur
	s.cur = reading{}

	lines := strings.Split(string(contents["/proc/diskstats"]), "\n")
	if len(lines) == 0 {
		return
	}

	var totalSectorsRead uint64
	var totalSectorsWritten uint64
	for _, line := range lines {
		line = strings.TrimSpace(line)

		// As per https://www.kernel.org/doc/Documentation/iostats.txt:
		f := strings.Fields(line)
		if len(f) < 10 {
			continue
		}
		if diskfilterRe.MatchString(f[2]) {
			// filters out e.g. dm-*
			continue
		}

		sectorsRead := must.Uint64(f[5])
		sectorsWritten := must.Uint64(f[9])
		totalSectorsRead += sectorsRead
		totalSectorsWritten += sectorsWritten
	}

	// TODO: reference linux kernel source if this is indeed hard-coded
	const sectorSize = 512

	s.cur.read = totalSectorsRead * sectorSize
	s.cur.write = totalSectorsWritten * sectorSize

	return
}

func (s *Stats) ProcessAndFormat(contents map[string][]byte) []stat.Col {
	s.process(contents)
	return []stat.Col{
		stat.ByteCol(s.cur.read - s.old.read).WithWidth(5),
		stat.ByteCol(s.cur.write - s.old.write).WithWidth(5),
	}
}
