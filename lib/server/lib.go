package server

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/digitalcircle-com-br/caroot"
	"github.com/digitalcircle-com-br/devserver/lib/config"
	"github.com/digitalcircle-com-br/devserver/lib/lpath"
	"github.com/digitalcircle-com-br/devserver/lib/mime"
)

var httpsServer http.Server

var muxes map[string]*http.ServeMux

type routeInfo struct {
	srcHost string
	srcPath string
	epProto string
	epHost  string
	epPath  string
}

var cli http.Client

func hStatic(ri routeInfo) func(writer http.ResponseWriter, r *http.Request) {
	return func(rw http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		nurlstr := strings.Replace(r.URL.Path, ri.srcPath, "", 1)
		fpath := strings.Split(nurlstr, "?")[0]
		ext := filepath.Ext(fpath)
		bs, err := os.ReadFile(path.Join(ri.epPath, fpath))
		if err != nil {
			rw.WriteHeader(http.StatusInternalServerError)
			rw.Write([]byte(err.Error()))
			return
		}
		rw.Header().Set("Content-Type", mime.Mime(ext))
		rw.Write(bs)
	}
}

func hRaw(ri routeInfo) func(rw http.ResponseWriter, r *http.Request) {
	return func(rw http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		nurlstr := strings.Replace(r.URL.Path, ri.srcPath, "", 1)
		fpath := strings.Split(nurlstr, "?")[0]

		bs, err := os.ReadFile(path.Join(ri.epPath, fpath+"_"+strings.ToUpper(r.Method)))
		if err != nil {
			bs, err = os.ReadFile(path.Join(ri.epPath, fpath))
		}
		res, err := http.ReadResponse(bufio.NewReader(bytes.NewReader(bs)), r)
		if err != nil {
			rw.WriteHeader(http.StatusInternalServerError)
			rw.Write([]byte(err.Error()))
			return
		}

		buf := &bytes.Buffer{}
		reslen, _ := io.Copy(buf, res.Body)
		for k, v := range res.Header {
			if strings.ToLower(k) == "content-length" {
				rw.Header().Add("content-length", fmt.Sprintf("%v", reslen))
			} else {
				for _, vv := range v {
					rw.Header().Add(k, vv)
				}
			}
		}
		rw.WriteHeader(res.StatusCode)
		defer res.Body.Close()
		rw.Write(buf.Bytes())
	}
}

func hRProxy(ri routeInfo) func(rw http.ResponseWriter, r *http.Request) {
	return func(rw http.ResponseWriter, r *http.Request) {
		var err error
		nurlstr := strings.Replace(r.URL.Path, ri.srcPath, "", 1)
		nurlstr = ri.epPath + nurlstr

		if r.URL.RawQuery != "" {
			nurlstr = nurlstr + "?" + r.URL.RawQuery
		}
		if !strings.HasPrefix(nurlstr, "/") {
			nurlstr = "/" + nurlstr
		}

		nurlstr = ri.epProto + "://" + ri.epHost + nurlstr

		u, err := url.Parse(nurlstr)

		if err != nil {
			http.Error(rw, fmt.Sprintf("error parsing url  %#v: %s", ri, err.Error()), http.StatusInternalServerError)
			return
		}
		r.URL = u
		r.Host = r.URL.Host
		r.RequestURI = ""
		r.Proto = "HTTP/1.1"
		res, err := cli.Do(r)
		if err != nil {
			http.Error(rw, fmt.Sprintf("error proxying  %#v: %s", ri, err.Error()), http.StatusInternalServerError)
			return
		}

		for k, v := range res.Header {
			for _, vv := range v {
				rw.Header().Add(k, vv)
			}
		}
		rw.WriteHeader(res.StatusCode)
		defer res.Body.Close()
		io.Copy(rw, res.Body)
	}

}

func buildEp(ri routeInfo) {

	mx, ok := muxes[ri.srcHost]
	if !ok {
		mx = &http.ServeMux{}
		muxes[ri.srcHost] = mx
	}

	log.Printf("Building route: %s[%s] => [%s]%s[%s]", ri.srcHost, ri.srcPath, ri.epProto, ri.epHost, ri.epPath)

	if ri.epHost == "" {
		if runtime.GOOS != "windows" {
			ri.epPath = "/" + ri.epPath
		}

	}
	ri.epHost = lpath.Resolve(ri.epHost)

	switch ri.epProto {
	case "static":
		mx.HandleFunc(ri.srcPath, mwPerf(mwCors(hStatic(ri))))
	case "raw":
		mx.HandleFunc(ri.srcPath, mwPerf(mwCors(hRaw(ri))))
	default:
		mx.HandleFunc(ri.srcPath, mwPerf(mwCors(hRProxy(ri))))
	}

}

func buildEps() {
	muxes = make(map[string]*http.ServeMux)
	for k, v := range config.Cfg.Routes {
		ri := routeInfo{}
		parts := strings.Split(k, "/")
		ri.srcHost = parts[0]
		ri.srcPath = strings.Join(parts[1:], "/")

		if !strings.HasSuffix(ri.srcPath, "/") {
			ri.srcPath = ri.srcPath + "/"
		}

		if !strings.HasPrefix(ri.srcPath, "/") {
			ri.srcPath = "/" + ri.srcPath
		}
		parts = strings.Split(v, "://")
		ri.epProto = parts[0]
		parts = strings.Split(parts[1], "/")
		ri.epHost = parts[0]
		ri.epPath = strings.Join(parts[1:], "/")

		buildEp(ri)
	}
}

func StartHttpsServer() error {
	_, err := config.LoadCfg()
	if err != nil {
		return err
	}
	tlscfg := &tls.Config{
		GetCertificate: func(info *tls.ClientHelloInfo) (*tls.Certificate, error) {
			ca := caroot.GetOrGenFromRoot(info.ServerName)
			return ca, nil
		},
	}

	buildEps()

	mx := &http.ServeMux{}

	mx.HandleFunc("/", func(rw http.ResponseWriter, r *http.Request) {

		h := strings.Split(r.Host, ":")[0]
		ep, ok := muxes[h]

		if !ok {
			ep, ok = muxes["*"]
			if !ok {
				http.Error(rw, fmt.Sprintf("unknown route for %s", h), http.StatusBadGateway)
				return
			}
		}

		ep.ServeHTTP(rw, r)
		return

	})

	httpsServer = http.Server{
		Addr:      config.Cfg.Addr,
		Handler:   mx,
		TLSConfig: tlscfg,
	}

	go func() {
		err := httpsServer.ListenAndServeTLS("", "")
		if err != nil {
			log.Printf("Finishing server: %s", err.Error())
		}
	}()
	return nil
}

func StopHttpServer() {
	httpsServer.Close()
}
