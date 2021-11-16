package main

import (
        "log"
	"net/http"
	"net/url"
        "os"
	"strings"

	"github.com/ynori7/reverse-proxy/client"
	"github.com/ynori7/reverse-proxy/rewriter"
	"github.com/ynori7/reverse-proxy/resources"
)

const baseUrl = "https://sfinlay-test.herokuapp.com"

type myClient struct {
	reverseProxy http.Handler
}

func (c myClient) GetWithProxy(w http.ResponseWriter, r *http.Request) {
	u := r.URL.Query().Get("u")
	if u == "" {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(resources.LandingPageHTML))
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
	port := os.Getenv("PORT")
	if port == "" {
		log.Fatal("$PORT must be set")
	}

	reverseProxy := client.NewReverseProxyClient(
		baseUrl+":"+port,
		rewriter.RewriteHtml,
	)

	c := myClient{reverseProxy: reverseProxy}

        log.Println("Starting server...")

	http.HandleFunc("/", c.GetWithProxy)
	log.Println(http.ListenAndServe(":"+port, nil))
}
