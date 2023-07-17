package statflag

import (
	"strings"

	"github.com/gokrazy/stat"
	"github.com/gokrazy/stat/internal/cpu"
	"github.com/gokrazy/stat/internal/disk"
	"github.com/gokrazy/stat/internal/mem"
	"github.com/gokrazy/stat/internal/net"
	"github.com/gokrazy/stat/internal/sys"
	"github.com/gokrazy/stat/internal/thermal"
)

type ProcessAndFormatter interface {
	ProcessAndFormat(map[string][]byte) []stat.Col
}

func ModulesFromFlag(enabledModules string) ([]ProcessAndFormatter, bool) {
	if strings.Contains(enabledModules, "thermal") {
		return []ProcessAndFormatter{
			&cpu.Stats{},
			&disk.Stats{},
			&sys.Stats{},
			&net.Stats{},
			&mem.Stats{},
			&thermal.Stats{},
		}, true
	} else {
		return []ProcessAndFormatter{
			&cpu.Stats{},
			&disk.Stats{},
			&sys.Stats{},
			&net.Stats{},
			&mem.Stats{},
		}, false
	}
}
