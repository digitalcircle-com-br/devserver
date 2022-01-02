package server

import (
	"log"
	"net/http"
	"sync"

	_ "embed"

	"github.com/digitalcircle-com-br/random"
	"github.com/gorilla/websocket"
)

//go:embed weblog.html
var html []byte

type webLogWriter struct {
}

type jsonLogMsg struct {
	Msg string `json:"msg"`
}

func (w *webLogWriter) Write(b []byte) (int, error) {
	mx.RLock()
	for _, v := range outs {
		v.WriteJSON(jsonLogMsg{string(b)})
	}
	mx.RUnlock()
	return len(b), nil
}

var mx sync.RWMutex
var outs map[string]*websocket.Conn

func WebLogSetup() {
	w := &webLogWriter{}
	log.SetOutput(w)
}

var upgrader = websocket.Upgrader{} // use default options
func WebLogMux() *http.ServeMux {
	ret := &http.ServeMux{}
	outs = make(map[string]*websocket.Conn)

	ret.HandleFunc("/__log/index.html", func(rw http.ResponseWriter, r *http.Request) {
		rw.Header().Add("Content-type", "text/html")
		rw.Write(html)
	})

	ret.HandleFunc("/", func(rw http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/index.html" {
			rw.Header().Add("Content-type", "text/html")
			rw.Write(html)
			return
		}
		con, err := upgrader.Upgrade(rw, r, nil)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}

		id := random.StrTSNano(8)

		mx.Lock()
		outs[id] = con
		mx.Unlock()
		defer func() {
			mx.Lock()
			delete(outs, id)
			mx.Unlock()
		}()

		con.WriteJSON(jsonLogMsg{"Staring log"})

		<-r.Context().Done()
	})

	return ret
}
