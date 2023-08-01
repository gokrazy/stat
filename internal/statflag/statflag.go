package statflag

import (
	"fmt"
	"strings"

	"github.com/gokrazy/stat"
	"github.com/gokrazy/stat/internal/cpu"
	"github.com/gokrazy/stat/internal/disk"
	"github.com/gokrazy/stat/internal/mem"
	"github.com/gokrazy/stat/internal/net"
	"github.com/gokrazy/stat/internal/sys"
	"github.com/gokrazy/stat/internal/thermal"
)

type Module interface {
	ProcessAndFormat(map[string][]byte) []stat.Col
	Headers() []string
}

func ModulesFromFlag(enabledModules string) ([]Module, error) {
	var modules []Module
	for _, name := range strings.Split(strings.TrimSpace(enabledModules), ",") {
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}
		switch name {
		case "cpu":
			modules = append(modules, &cpu.Stats{})
		case "disk":
			modules = append(modules, &disk.Stats{})
		case "sys":
			modules = append(modules, &sys.Stats{})
		case "net":
			modules = append(modules, &net.Stats{})
		case "mem":
			modules = append(modules, &mem.Stats{})
		case "thermal":
			modules = append(modules, &thermal.Stats{})
		default:
			return nil, fmt.Errorf("unknown module: %q", name)
		}
	}
	return modules, nil
}
