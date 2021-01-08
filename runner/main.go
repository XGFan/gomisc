package main

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os/exec"
	"strings"
	"sync"
	"time"
)

type Runner struct {
	Map map[string]string
}

type MultiWriter struct {
	writers []io.Writer
}

func removeEmpty(slice *[]string) {
	i := 0
	for _, v := range *slice {
		if v != "" {
			(*slice)[i] = v
			i++
		}
	}
	*slice = (*slice)[:i]
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
				mutex := sync.Mutex{}
				mutex.Lock()
				go func() {
					defer mutex.Unlock()
					time.Sleep(5 * time.Second)
					_, _ = w.Write([]byte("\nTo Be Continued\n"))
				}()
				go func() {
					defer mutex.Unlock()
					log.Printf("EXEC: %s %s", cmd, args)
					w.WriteHeader(200)
					argsForCommand := strings.Split(string(args), " ")
					removeEmpty(&argsForCommand)
					command := exec.Command(cmd, argsForCommand...)
					multiWriter := MultiWriter{[]io.Writer{log.Writer(), w}}
					command.Stdout = multiWriter
					command.Stderr = multiWriter
					runError := command.Run()
					if runError != nil {
						log.Println(runError)
					}
				}()
				mutex.Lock()
			}
		}
	} else {
		w.WriteHeader(400)
		_, _ = w.Write([]byte("Only Support POST\n"))
	}
}

func main() {
	addrAndPort := flag.String("b", "127.0.0.1:1888", "bind address and port")
	config := flag.String("c", ".config", "config file location")
	flag.Parse()
	file, err := ioutil.ReadFile(*config)
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
	log.Printf("Listen: %s\n", *addrAndPort)
	log.Printf("Load %s : %v ", *config, m)
	http.ListenAndServe(*addrAndPort, Runner{Map: m})
}
