package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/gokrazy/stat"
	"github.com/gokrazy/stat/internal/statflag"
	"golang.org/x/sync/errgroup"
)

func formatCols(cols []stat.Col) string {
	formatted := make([]string, len(cols))
	for idx, col := range cols {
		formatted[idx] = fmt.Sprintf(
			`<td style="width: %dem">%s</td>`,
			col.Width,
			col.HTML())
	}
	return strings.Join(formatted, "\n")
}

func serveStats() error {
	var listen = flag.String("listen", ":6618", "[host]:port to serve HTML on")
	var enabledModules = flag.String("modules", "cpu,disk,sys,net,mem", "comma-separated list of modules to show. known modules: cpu,disk,sys,net,mem,thermal")
	flag.Parse()

	modules, err := statflag.ModulesFromFlag(*enabledModules)
	if err != nil {
		return err
	}

	var headers []string
	for _, mod := range modules {
		hdrs := mod.Headers()
		for idx, h := range hdrs {
			hdrs[idx] = strings.ReplaceAll(h, " ", "_")
		}
		headers = append(headers, hdrs...)
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

	statusTmpl, err := template.New("").Parse(statusTmpl)
	if err != nil {
		return err
	}
	hostname, _ := os.Hostname()
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		var buf bytes.Buffer
		if err := statusTmpl.Execute(&buf, struct {
			Hostname string
			Headers  []string
		}{
			Hostname: hostname,
			Headers:  headers,
		}); err != nil {
			log.Print(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		io.Copy(w, &buf)
	})
	var newEventMu sync.Mutex
	newEvent := sync.NewCond(&newEventMu)
	http.HandleFunc("/readings", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		newEventMu.Lock()
		for {
			b, err := json.Marshal(parts)
			if err != nil {
				log.Print(err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			rdr := bytes.NewReader(append(append([]byte("data: "), b...), []byte("\n\n")...))
			newEventMu.Unlock()

			if _, err := io.Copy(w, rdr); err != nil {
				log.Print(err)
				return
			}
			if f, ok := w.(http.Flusher); ok {
				f.Flush()
			}

			newEventMu.Lock()
			newEvent.Wait()
		}

	})
	var eg errgroup.Group
	eg.Go(func() error {
		for i := 0; ; i++ {
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

			newEventMu.Lock()
			for idx, mod := range modules {
				parts[idx] = formatCols(mod.ProcessAndFormat(contents))
			}
			newEventMu.Unlock()

			if i > 0 {
				const darkblue = "\033[0;34m"
				//fmt.Println(strings.Join(parts, darkblue+" | "))
				// TODO: clear colors at the end of line so that program can be interrupted
				newEvent.Broadcast()
			}

			time.Sleep(1 * time.Second)
		}
	})
	eg.Go(func() error {
		log.Printf("listening on %s", *listen)
		return http.ListenAndServe(*listen, nil)
	})
	return eg.Wait()
}

func main() {
	if err := serveStats(); err != nil {
		log.Fatal(err)
	}
}
