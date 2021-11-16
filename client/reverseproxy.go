package client

import (
	"bytes"
	"compress/gzip"
	"crypto/tls"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"

	"github.com/andybalholm/brotli"
)

type TransformResponseFunc func(uri string, reader io.Reader) string

type ReverseProxy struct {
	*httputil.ReverseProxy

	baseUrl               string
	transformResponseFunc TransformResponseFunc
}

func (t *ReverseProxy) ModifyResponse(resp *http.Response) error {
	log.Printf("Handling %s. Status %d", resp.Request.URL.String(), resp.StatusCode)
	resp.Header.Set("X-Proxied-By", "Scott's Go Proxy")

	if resp.StatusCode == 301 || resp.StatusCode == 302 {
		v := url.Values{}
		v.Add("u", resp.Header.Get("Location"))
		resp.Header.Set("Location", t.baseUrl+"?"+v.Encode())
		return nil
	}

	//only rewrite the response for html content
	if strings.Index(resp.Header.Get("Content-Type"), "text/html") < 0 {
		return nil
	}

	v := url.Values{}
	v.Add("u", resp.Request.URL.String())
	uri := t.baseUrl + "?" + v.Encode()

	t.replaceBody(resp, t.transformResponseFunc(uri, t.getBodyReader(resp)))
	return nil
}

func (t *ReverseProxy) replaceBody(resp *http.Response, newBody string) {
	var reader io.Reader
	var contentLength int
	switch resp.Header.Get("Content-Encoding") {
	case "gzip":
		reader = new(bytes.Buffer)
		w := gzip.NewWriter(reader.(io.Writer))
		contentLength, _ = w.Write([]byte(newBody))
		w.Close()
	case "br":
		reader = new(bytes.Buffer)
		w := brotli.NewWriter(reader.(io.Writer))
		contentLength, _ = w.Write([]byte(newBody))
		w.Close()
	default:
		reader = strings.NewReader(newBody)
		contentLength = len(newBody)
	}
	body := ioutil.NopCloser(reader)
	resp.Body = body
	resp.ContentLength = int64(contentLength)
	resp.Header.Del("Content-Length")
}

func (t *ReverseProxy) getBodyReader(resp *http.Response) io.Reader {
	var reader io.Reader

	switch resp.Header.Get("Content-Encoding") {
	case "gzip":
		reader, _ = gzip.NewReader(resp.Body)
	case "br":
		reader = brotli.NewReader(resp.Body)
	default:
		reader = resp.Body
	}

	return reader
}

func NewReverseProxyClient(baseUrl string, transformResponseFunc TransformResponseFunc) *ReverseProxy {
	tr := &http.Transport{
		Dial: (&net.Dialer{
			Timeout: 5 * time.Second,
		}).Dial,
		TLSHandshakeTimeout:   3 * time.Second,
		ResponseHeaderTimeout: 3 * time.Second,
		TLSClientConfig:       &tls.Config{InsecureSkipVerify: true},
	}

	proxy := new(ReverseProxy)
	proxy.baseUrl = baseUrl
	proxy.transformResponseFunc = transformResponseFunc
	proxy.ReverseProxy = &httputil.ReverseProxy{
		Director:       func(req *http.Request) {},
		Transport:      tr,
		ModifyResponse: proxy.ModifyResponse,
	}

	return proxy
}
