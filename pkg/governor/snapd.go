package governor

import (
	"context"
	"net"
	"net/http"
	"net/http/httputil"
	"strings"
)

func NewSnapdProxy(requestPrefix string, snapdSocket string) http.Handler {
	proxy := &httputil.ReverseProxy{
		Director: func(req *http.Request) {
			path := strings.TrimPrefix(req.URL.Path, requestPrefix)

			req.URL.Host = "unix"
			req.URL.Path = path
			req.URL.Scheme = "http"
		},
		Transport: &http.Transport{
			DialContext: func(_ context.Context, _, _ string) (net.Conn, error) {
				return net.Dial("unix", snapdSocket)
			},
		},
	}

	return proxy
}
