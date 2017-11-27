package main

import (
	"net/http"
	"net/url"

	"github.com/ynori7/reverse-proxy/client"
	"github.com/ynori7/reverse-proxy/rewriter"
)

type myClient struct {
	reverseProxy http.Handler
}

func (c myClient) GetWithProxy(w http.ResponseWriter, r *http.Request) {
	u := r.URL.Query().Get("u")
	requestedUrl, err := url.ParseRequestURI(u)
	if u == "" || err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Header().Set("Content-Type", "Text")
		w.Write([]byte("Missing the 'u' query parameter. Please supply a URL to proxy."))
		return
	}

	r.URL = requestedUrl
	r.Host = requestedUrl.Host

	c.reverseProxy.ServeHTTP(w, r)
}

func main() {
	reverseProxy := client.NewReverseProxyClient(
		"http://localhost:8001",
		rewriter.RewriteHtml,
	)

	c := myClient{reverseProxy: reverseProxy}

	http.HandleFunc("/", c.GetWithProxy)
	http.ListenAndServe(":8001", nil)
}
