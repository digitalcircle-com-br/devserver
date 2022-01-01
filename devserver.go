package main

import (
	"bufio"
	"bytes"
	"crypto/tls"
	_ "embed"
	"fmt"
	"io"
	"log"
	"mime"
	"net/http"
	"net/url"
	"os"
	"os/user"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/digitalcircle-com-br/caroot"
	"github.com/getlantern/systray"
	"github.com/natefinch/lumberjack"
	"github.com/skratchdot/open-golang/open"
	"gopkg.in/yaml.v2"
)

//go:embed lib/logo_dc_square.png
var logo_png []byte

//go:embed lib/logo_dc_square.ico
var logo_ico []byte

var builtinMimeTypesLower = map[string]string{
	".css":  "text/css; charset=utf-8",
	".gif":  "image/gif",
	".htm":  "text/html; charset=utf-8",
	".html": "text/html; charset=utf-8",
	".jpg":  "image/jpeg",
	".js":   "application/javascript",
	".wasm": "application/wasm",
	".pdf":  "application/pdf",
	".png":  "image/png",
	".svg":  "image/svg+xml",
	".xml":  "text/xml; charset=utf-8",
}

func Mime(ext string) string {
	if v, ok := builtinMimeTypesLower[ext]; ok {
		return v
	}
	return mime.TypeByExtension(ext)
}

type cfg struct {
	Addr   string            `json:"addr" yaml:"addr"`
	Log    string            `json:"log" yaml:"log"`
	Routes map[string]string `json:"routes" yaml:"routes"`
}

var Cfg *cfg = &cfg{}

func LoadCfg() (bool, error) {
	Cfg = &cfg{}
	bs, err := os.ReadFile("config.yaml")
	if err != nil {
		return false, err
	}
	err = yaml.Unmarshal(bs, Cfg)
	return true, err
}

func SaveCfg() error {
	bs, _ := yaml.Marshal(Cfg)
	return os.WriteFile("config.yaml", bs, 0600)
}

var httpsServer http.Server

var muxes map[string]*http.ServeMux

type routeInfo struct {
	srcHost string
	srcPath string
	epProto string
	epHost  string
	epPath  string
}

func mwCors(h func(rw http.ResponseWriter, r *http.Request)) func(rw http.ResponseWriter, r *http.Request) {
	return func(rw http.ResponseWriter, r *http.Request) {
		orig := r.Header.Get("Origin")
		if orig == "" {
			orig = strings.Split(r.Host, ":")[0]
		}

		rw.Header().Set("Access-Control-Allow-Origin", orig)
		rw.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		rw.Header().Set("Access-Control-Allow-Headers", "Last-Modified, Expires, Accept, Cache-Control, Content-Type, Content-Language,Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, Pragma")
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

var cli http.Client

func buildEp(ri routeInfo) {

	mx, ok := muxes[ri.srcHost]
	if !ok {
		mx = &http.ServeMux{}
		muxes[ri.srcHost] = mx
	}

	log.Printf("Building route: %s[%s] => [%s]%s[%s]", ri.srcHost, ri.srcPath, ri.epProto, ri.epHost, ri.epPath)

	switch ri.epProto {
	case "static":
		if ri.epHost == "" {
			if runtime.GOOS != "windows" {
				ri.epPath = "/" + ri.epPath
			}

		}
		mx.HandleFunc(ri.srcPath, mwPerf(mwCors(func(writer http.ResponseWriter, r *http.Request) {
			defer r.Body.Close()
			nurlstr := strings.Replace(r.URL.Path, ri.srcPath, "", 1)
			fpath := strings.Split(nurlstr, "?")[0]
			ext := filepath.Ext(fpath)
			bs, err := os.ReadFile(path.Join(ri.epPath, fpath))
			if err != nil {
				writer.WriteHeader(http.StatusInternalServerError)
				writer.Write([]byte(err.Error()))
				return
			}
			writer.Header().Set("Content-Type", Mime(ext))
			writer.Write(bs)
		})))
	case "raw":
		if ri.epHost == "" {
			if runtime.GOOS != "windows" {
				ri.epPath = "/" + ri.epPath
			}

		}
		mx.HandleFunc(ri.srcPath, mwPerf(mwCors(func(rw http.ResponseWriter, r *http.Request) {
			defer r.Body.Close()
			nurlstr := strings.Replace(r.URL.Path, ri.srcPath, "", 1)
			fpath := strings.Split(nurlstr, "?")[0]

			bs, err := os.ReadFile(path.Join(ri.epPath, fpath+"_"+r.Method))
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

		})))
	default:
		mx.HandleFunc(ri.srcPath, mwPerf(mwCors(func(rw http.ResponseWriter, r *http.Request) {
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

		})))
	}

}

func buildEps() {
	muxes = make(map[string]*http.ServeMux)
	for k, v := range Cfg.Routes {
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
	_, err := LoadCfg()
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

		//nurlstr := ep + r.URL.Path
		//if r.URL.RawQuery != "" {
		//	nurlstr = nurlstr + "?" + r.URL.RawQuery
		//}
		//u, err := url.Parse(nurlstr)
		//
		//if err != nil {
		//	http.Error(rw, fmt.Sprintf("error parsing url  %s", err.Error()), http.StatusInternalServerError)
		//}
		//r.URL = u
		//r.Host = r.URL.Host
		//r.RequestURI = ""
		//res, err := cli.Do(r)
		//if err != nil {
		//	http.Error(rw, fmt.Sprintf("error proxying  %s", ep), http.StatusInternalServerError)
		//}
		//
		//for k, v := range res.Header {
		//	for _, vv := range v {
		//		rw.Header().Add(k, vv)
		//	}
		//}
		//rw.WriteHeader(res.StatusCode)
		//defer res.Body.Close()
		//io.Copy(rw, res.Body)

	})

	httpsServer = http.Server{
		Addr:      Cfg.Addr,
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

var wd string

func Start() error {
	usr, err := user.Current()
	if err != nil {
		return err
	}
	wd = path.Join(usr.HomeDir, ".devserver")
	os.MkdirAll(wd, os.ModePerm)
	os.Chdir(wd)

	found, err := LoadCfg()
	if !found {
		Cfg.Addr = ":8443"
		Cfg.Log = "devserver.log"
		SaveCfg()

	} else if err != nil {
		return err
	}

	if Cfg.Log != "-" {
		log.SetOutput(&lumberjack.Logger{
			Filename:   Cfg.Log,
			MaxSize:    25, // megabytes
			MaxBackups: 10,
			MaxAge:     28,    //days
			Compress:   false, // disabled by default
		})
	}

	caroot.InitCA("caroot", func(ca string) {
		log.Printf("Initiating CA: %s", ca)
	})

	//port := flag.Int("port", 8080, "Port for control listener")

	StartHttpsServer()

	systray.Run(onReady, onExit)
	return nil
}

func onReady() {
	if runtime.GOOS == "windows" {
		systray.SetIcon(logo_ico)
	} else {
		systray.SetIcon(logo_png)
	}

	systray.SetTitle("DC - DevServer")
	systray.SetTooltip("Digital Circle - Development Server & Gateway")

	mRestart := systray.AddMenuItem("Restart", "")
	mOpenDir := systray.AddMenuItem("Open Dir", "")
	systray.AddSeparator()
	systray.AddMenuItem("Digital Circle® - V:0.0.5", "")
	systray.AddSeparator()
	mQuit := systray.AddMenuItem("Quit", "")

	for {
		select {
		case <-mRestart.ClickedCh:
			StopHttpServer()
			StartHttpsServer()
		case <-mOpenDir.ClickedCh:
			open.Run(wd)
		case <-mQuit.ClickedCh:
			systray.Quit()
		}
	}

}

func onExit() {
	// clean up here
}

func main() {
	Start()
}
