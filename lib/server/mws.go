package server

import (
	"log"
	"net/http"
	"strings"
	"time"
)

func mwCors(h func(rw http.ResponseWriter, r *http.Request)) func(rw http.ResponseWriter, r *http.Request) {
	return func(rw http.ResponseWriter, r *http.Request) {
		orig := r.Header.Get("Origin")
		if orig == "" {
			orig = strings.Split(r.Host, ":")[0]
		}

		rw.Header().Set("Access-Control-Allow-Origin", orig)
		rw.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		rw.Header().Set("Access-Control-Allow-Headers", "Last-Modified, Expires, Accept, Cache-Control, Content-Type, Content-Language,Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, Pragma, Upgrade")
		rw.Header().Set("Access-Control-Allow-Credentials", "true")
		h(rw, r)
	}
}

func mwPerf(h func(rw http.ResponseWriter, r *http.Request)) func(rw http.ResponseWriter, r *http.Request) {
	return func(rw http.ResponseWriter, r *http.Request) {
		start := time.Now()
		h(rw, r)
		dur := time.Now().Sub(start)
		log.Printf("[%s]%s: %s", r.Method, r.URL, dur.String())
	}
}
