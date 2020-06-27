package stat

import (
	"fmt"
	"strconv"
	"strings"
)

const (
	UnitNone = iota
	UnitBytes
	UnitBytesFloat
	UnitMetric
)

var colorNameToANSI = map[string]string{
	"darkgray": "\033[1;30m",
	"red":      "\033[1;31m",
	"green":    "\033[1;32m",
	"yellow":   "\033[1;33m",
	"blue":     "\033[1;34m",
	"magenta":  "\033[1;35m",
	"cyan":     "\033[1;36m",
	"white":    "\033[1;37m",
}

var colorNameToHTML = map[string]string{
	"darkgray": "#555753",
	"red":      "#EF2929",
	"green":    "#8AE234",
	"yellow":   "#FCE94F",
	"blue":     "#729FCF",
	"magenta":  "#EE38DA",
	"cyan":     "#34E2E2",
	"white":    "#EEEEEC",
}

const (
	black       = "\033[0;30m"
	darkred     = "\033[0;31m"
	darkgreen   = "\033[0;32m"
	darkyellow  = "\033[0;33m"
	darkblue    = "\033[0;34m"
	darkmagenta = "\033[0;35m"
	darkcyan    = "\033[0;36m"
	gray        = "\033[0;37m"

	darkgray = "darkgray"
	red      = "red"
	green    = "green"
	yellow   = "yellow"
	blue     = "blue"
	magenta  = "magenta"
	cyan     = "cyan"
	white    = "white"

	blackbg   = "\033[40m"
	redbg     = "\033[41m"
	greenbg   = "\033[42m"
	yellowbg  = "\033[43m"
	bluebg    = "\033[44m"
	magentabg = "\033[45m"
	cyanbg    = "\033[46m"
	whitebg   = "\033[47m"
)

var colors = []string{
	red,
	yellow,
	green,
	blue,
	cyan,
	white,
	darkred,
	darkgreen,
}

type ColType int

const (
	ColGauge ColType = iota
	ColPercentage
)

type Col struct {
	Type       ColType
	ValU64     uint64
	ValInt     int     // used by ColPercentage
	ValFloat64 float64 // used by UnitBytes
	Unit       int
	Width      int
	Scale      int
}

func (c Col) WithWidth(w int) Col {
	ret := c
	ret.Width = w
	return ret
}

var unitsuffix = []string{"B", "k", "M", "G", "T"}

func (c Col) String() string {
	return c.colorize(func(col, text string) string {
		return colorNameToANSI[col] + text
	})
}

func (c Col) HTML() string {
	return c.colorize(func(col, text string) string {
		return fmt.Sprintf(`<span style="color: %s">%s</span>`, colorNameToHTML[col], text)
	})
}

func (c Col) colorize(color func(col, text string) string) string {
	switch c.Type {
	case ColPercentage:
		v := fmt.Sprintf("%"+strconv.Itoa(c.Width)+"d", int(c.ValFloat64))
		col := colors[int(c.ValFloat64/float64(c.Scale))%len(colors)]
		if strings.TrimSpace(v) == "0" {
			return color(darkgray, v)
		}
		if c.ValFloat64 >= 100 {
			return color(white, v)
		}
		return color(col, v)

	case ColGauge:
		if c.Unit == UnitBytes || c.Unit == UnitBytesFloat ||
			c.Unit == UnitMetric {
			base := 1024
			if c.Unit == UnitMetric {
				base = 1000
			}
			width := c.Width
			if c.ValU64 == 0 && c.ValFloat64 == 0 {
				return color(darkgray, fmt.Sprintf("%"+strconv.Itoa(width)+"d", 0))
			}
			width-- // for the unit suffix
			var f string
			var cl int
			if c.Unit == UnitBytesFloat {
				f, cl = fchg(c.ValFloat64, width, base)
			} else {
				f, cl = dchg(c.ValU64, width, base)
			}
			if len(f) < width {
				f = strings.Repeat(" ", width-len(f)) + f
			}
			if cl > 0 || c.Unit == UnitBytes || c.Unit == UnitBytesFloat {
				if cl < len(unitsuffix) {
					f += color(darkgray, unitsuffix[cl])
				} else {
					f += "?"
				}
			} else {
				f += " " // empty suffix
			}
			col := colors[cl%len(colors)]
			return color(col, f)
		}
		return fmt.Sprintf("%4d", c.ValU64)

	default:
		return "?BUG?"
	}
}

func ByteCol(v uint64) Col {
	return Col{
		Type:   ColGauge,
		Unit:   UnitBytes,
		ValU64: v,
	}
}

func MetricCol(v uint64) Col {
	return Col{
		Type:   ColGauge,
		Unit:   UnitMetric,
		ValU64: v,
	}
}

// -----------------------------------------------------------------------------
// dchg and fchg were ported directly from dstat.
// -----------------------------------------------------------------------------

func dchg(v uint64, width int, base int) (string, int) {
	c := 0
	for {
		ret := strconv.FormatUint(v, 10)
		if len(ret) <= width {
			return ret, c
		}
		v = v / uint64(base)
		c++
	}
}

func fchg(v float64, width int, base int) (string, int) {
	if v == 0 {
		return "0", 0
	}
	c := 0
	var ret string
	for {
		ret = strconv.Itoa(int(v))
		if len(ret) <= width {
			i := width - len(ret) - 1
			for ; i > 0; i-- {
				ret = strconv.FormatFloat(v, 'f', i, 64)
				if len(ret) <= width {
					break
				}
			}
			if i == 0 {
				ret = strconv.Itoa(int(v))
			}
			break
		}
		v = v / float64(base)
		c++
	}
	return ret, c
}
