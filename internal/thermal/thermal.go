package thermal

import (
	"strings"

	"github.com/gokrazy/stat"
	"github.com/gokrazy/stat/internal/must"
)

type reading struct {
	thermalZone uint64
}

type Stats struct {
	cur reading
}

func (s *Stats) FileContents() []string {
	return []string{"/sys/class/thermal/thermal_zone0/temp"}
}

func (s *Stats) process(contents map[string][]byte) {
	s.cur = reading{}

	line := string(contents["/sys/class/thermal/thermal_zone0/temp"])
	if len(line) == 0 {
		return
	}

	therm := strings.TrimSpace(line)
	s.cur.thermalZone = must.Uint64(therm)
}

func (s *Stats) ProcessAndFormat(contents map[string][]byte) []stat.Col {
	s.process(contents)
	return []stat.Col{
		{Type: stat.ColPercentage, ValFloat64: float64(s.cur.thermalZone) / 1000, Width: 3, Scale: 34},
	}
}
