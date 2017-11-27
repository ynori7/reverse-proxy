package main

import (
	"net/http"
	"net/url"
	"strings"

	"github.com/ynori7/reverse-proxy/client"
	"github.com/ynori7/reverse-proxy/rewriter"
)

type myClient struct {
	reverseProxy http.Handler
}

func (c myClient) GetWithProxy(w http.ResponseWriter, r *http.Request) {
	u := r.URL.Query().Get("u")
	if u == "" {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(`
<html>
<head><title>Scott's Go Proxy</title></head>
<body>
<form action='/' method='GET' style='width:1000px;margin:100px auto;'>
	<input type='text' name='u' size='100' placeholder='Enter a URL here...'/>
	<input type='submit' value='Go!'/>
</form>
</body>
</html>
		`))
		return
	}

	if strings.Index(u, "http") < 0 {
		u = "http://" + u //try to make the url valid and hope the server will redirect if necessary
	}

	requestedUrl, err := url.ParseRequestURI(u)
	if err != nil {
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
