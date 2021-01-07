package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os/exec"
	"strings"
	"time"
)

type Runner struct {
	Map map[string]string
}

type MultiWriter struct {
	writers []io.Writer
}

func (w MultiWriter) Write(p []byte) (int, error) {
	maxWrite := 0
	var e error = nil
	for _, writer := range w.writers {
		write, err := writer.Write(p)
		if err != nil {
			e = err
		} else if write > maxWrite {
			maxWrite = write
		}
	}
	if maxWrite != len(p) {
		return maxWrite, e
	} else {
		return maxWrite, nil
	}
}

func (p Runner) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		path := r.URL.Path
		alias := path[1:]
		var cmd = p.Map[alias]
		if cmd == "" {
			w.WriteHeader(404)
			_, _ = w.Write([]byte(fmt.Sprintf("Command %s Not Found\n", alias)))
		} else {
			args, err := ioutil.ReadAll(r.Body)
			if err != nil {
				w.WriteHeader(500)
				_, _ = w.Write([]byte("Read Body Fail\n"))
			} else {
				tick := make(chan bool, 1)
				finished := false
				go func() {
					time.Sleep(8 * time.Second)
					if !finished {
						_, _ = w.Write([]byte("\nTo Be Continued\n"))
					}
					tick <- true
				}()
				go func() {
					log.Printf("EXEC: %s %s", cmd, args)
					w.WriteHeader(200)
					command := exec.Command(cmd, strings.Split(string(args), " ")...)
					multiWriter := MultiWriter{[]io.Writer{log.Writer(), w}}
					command.Stdout = multiWriter
					command.Stderr = multiWriter
					//command.Stdin = strings.NewReader("\n")
					runError := command.Run()
					if runError != nil {
						log.Println(runError)
					}
					finished = true
					tick <- true
				}()
				<-tick
			}
		}
	} else {
		w.WriteHeader(400)
		_, _ = w.Write([]byte("Only Support POST\n"))
	}
}

func main() {
	file, err := ioutil.ReadFile(".config")
	if err != nil {
		log.Fatal(err)
	}
	m := make(map[string]string, 0)
	programs := strings.Split(string(file), "\n")
	for _, program := range programs {
		if program != "" {
			index := strings.Index(program, "=")
			if index != -1 {
				m[program[:index]] = program[index+1:]
			} else {
				m[program] = program
			}
		}
	}
	log.Printf("load: %v", m)
	http.ListenAndServe("127.0.0.1:1888", Runner{Map: m})
}
