package server

import (
	"fmt"
	"github.com/koding/websocketproxy"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
)

type WebProxy struct {
	path  string
	mux   *http.ServeMux
	vport int //vhost_http_port
}

func NewWebProxy(path string, mux *http.ServeMux) *WebProxy {
	s := WebProxy{
		path:  path,
		mux:   mux,
		vport: 58080,
	}
	return &s
}

func (r *WebProxy) Handle() {
	r.mux.Handle(fmt.Sprintf("%v/websocket/", r.path), r.wsProxy(fmt.Sprintf("ws://localhost:%v/websocket/", r.vport)))
	r.mux.Handle(fmt.Sprintf("%v/", r.path), http.StripPrefix(r.path, r.httpProxy(fmt.Sprintf("http://localhost:%v/", r.vport))))
}

func (r *WebProxy) wsProxy(remoteUrl string) http.Handler {
	target := toUrl(remoteUrl)
	handler := websocketproxy.NewProxy(target)

	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		fmt.Fprintf(os.Stdout, "wsProxy: %v\n", req.URL)

		handler.ServeHTTP(res, req)
	})
}

func (r *WebProxy) httpProxy(remoteUrl string) http.Handler {
	target := toUrl(remoteUrl)
	handler := httputil.NewSingleHostReverseProxy(target)
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		fmt.Fprintf(os.Stdout, "httpProxy: %v\n", req.URL)

		handler.ServeHTTP(res, req)
	})
}

func toUrl(s string) *url.URL {
	u, err := url.Parse(s)
	if err != nil {
		panic(err)
	}
	return u
}
