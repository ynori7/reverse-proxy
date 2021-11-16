package main

import (
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/ynori7/reverse-proxy/client"
	"github.com/ynori7/reverse-proxy/resources"
	"github.com/ynori7/reverse-proxy/rewriter"
)

const (
	defaultPort    = "8081"
	defaultBaseURL = "http://localhost:8081"
)

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

	requestedURL, err := url.ParseRequestURI(u)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Header().Set("Content-Type", "Text")
		w.Write([]byte("Missing the 'u' query parameter. Please supply a URL to proxy."))
		return
	}

	r.URL = requestedURL
	r.Host = requestedURL.Host

	c.reverseProxy.ServeHTTP(w, r)
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		log.Println("$PORT was empty. Using default")
		port = defaultPort
	}
	baseURL := os.Getenv("BASE_URL")
	if baseURL == "" {
		log.Println("$BASE_URL was empty. Using default")
		baseURL = defaultBaseURL
	}

	reverseProxy := client.NewReverseProxyClient(
		baseURL,
		rewriter.RewriteHtml,
	)

	c := myClient{reverseProxy: reverseProxy}

	log.Println("Starting server...")

	http.HandleFunc("/", c.GetWithProxy)
	log.Println(http.ListenAndServe(":"+port, nil))
}
