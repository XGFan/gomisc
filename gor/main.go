package main

import (
	"context"
	"crypto/tls"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"time"
)

func main() {
	ns, err := net.LookupIP("home.test4x.com")
	if err != nil {
		log.Fatalf("nslookup home.test4x.com fail, %s", err)
	}
	if len(ns) == 0 {
		log.Fatalf("nslookup home.test4x.com fail, 0 IP return")
	}
	ip := ns[0].String()
	log.Printf("start up with %s\n", ip)
	handler := func(p *httputil.ReverseProxy) func(http.ResponseWriter, *http.Request) {
		return func(w http.ResponseWriter, r *http.Request) {
			p.ServeHTTP(w, r)
		}
	}

	proxy := &httputil.ReverseProxy{
		Director: func(req *http.Request) {
			newIP := req.Header.Get("X-NEW-IP")
			if newIP != "" {
				ip = newIP
				log.Printf("change ip to %s", ip)
			}
			req.URL.Scheme = "https"
			req.URL.Host = ip + ":44443"
		},
		Transport: &http.Transport{
			DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				return net.DialTimeout(network, addr, time.Second*2)
			},
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}
	http.HandleFunc("/", handler(proxy))
	err = http.ListenAndServe("127.0.0.1:8080", nil)
	if err != nil {
		panic(err)
	}
}
