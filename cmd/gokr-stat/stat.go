package main

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"
	"unsafe"

	"github.com/gokrazy/stat"
	"github.com/gokrazy/stat/internal/statflag"
)

func formatCols(cols []stat.Col) string {
	formatted := make([]string, len(cols))
	for idx, col := range cols {
		formatted[idx] = col.String()
	}
	return strings.Join(formatted, " ")
}

const (
	TIOCGWINSZ = 0x5413
)

type window struct {
	Row    uint16
	Col    uint16
	Xpixel uint16
	Ypixel uint16
}

func terminalSize() (*window, error) {
	w := new(window)
	tio := syscall.TIOCGWINSZ
	res, _, err := syscall.Syscall(syscall.SYS_IOCTL,
		uintptr(syscall.Stdin),
		uintptr(tio),
		uintptr(unsafe.Pointer(w)),
	)
	if int(res) == -1 {
		if err == syscall.ENOTTY {
			return &window{Row: 80}, nil
		}
		return nil, err
	}
	return w, nil
}

func printStats() error {
	var enabledModules = flag.String("modules", "cpu,disk,sys,net,mem", "comma-separated list of modules to show. known modules: cpu,disk,sys,net,mem,thermal")
	flag.Parse()

	ts, err := terminalSize()
	if err != nil {
		return err
	}
	var rowsMu sync.Mutex
	rows := int(ts.Row) - 1
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGWINCH)
	go func() {
		for range ch {
			ts, err := terminalSize()
			if err != nil {
				log.Print(err)
				continue
			}
			rowsMu.Lock()
			rows = int(ts.Row) - 1
			rowsMu.Unlock()
		}
	}()

	modules, hasThermal := statflag.ModulesFromFlag(*enabledModules)

	header := func() {
		const blue = "\033[1;34m"
		fmt.Printf(blue + "usr sys idl wai stl | ")
		fmt.Printf(" read  writ | ")
		fmt.Printf(" int   csw  | ")
		fmt.Printf(" recv  send | ")
		if hasThermal {
			fmt.Printf(" used  free  buff  cach | ")
			fmt.Printf(" cpu\n")
		} else {
			fmt.Printf(" used  free  buff  cach\n")
		}
	}

	parts := make([]string, len(modules))
	files := make(map[string]*os.File)
	for _, mod := range modules {
		// When a stats module implements the FileContents() interface, we
		// ensure all returned file contents are read and passed to
		// ProcessAndFormat.
		fc, ok := mod.(interface{ FileContents() []string })
		if !ok {
			continue
		}
		for _, f := range fc.FileContents() {
			if _, ok := files[f]; ok {
				continue // already requested
			}
			fl, err := os.Open(f)
			if err != nil {
				return err
			}
			files[f] = fl
		}
	}
	for i := 0; ; i++ {
		rowsMu.Lock()
		showHeader := i%(int(rows)-1) == 0
		rowsMu.Unlock()
		if showHeader {
			header()
		}
		contents := make(map[string][]byte)
		for path, fl := range files {
			if _, err := fl.Seek(0, io.SeekStart); err != nil {
				return err
			}
			b, err := ioutil.ReadAll(fl)
			if err != nil {
				return err
			}
			contents[path] = b
		}

		for idx, mod := range modules {
			parts[idx] = formatCols(mod.ProcessAndFormat(contents))
		}

		if i > 0 {
			const darkblue = "\033[0;34m"
			fmt.Println(strings.Join(parts, darkblue+" | "))
			// TODO: clear colors at the end of line so that program can be interrupted
		}

		time.Sleep(1 * time.Second)
	}
}

func main() {
	if os.Getenv("GOKRAZY_FIRST_START") == "1" {
		// Do not supervise this process: it is meant for interactive usage on a
		// terminal. If you are looking for a daemon, use gokr-webstat instead.
		os.Exit(125)
	}

	if err := printStats(); err != nil {
		log.Fatal(err)
	}
}
