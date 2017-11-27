package rewriter

import (
	"bytes"
	"io"
	"io/ioutil"
	"net/url"
	"strings"

	"golang.org/x/net/html"
)

func RewriteHtml(uri string, reader io.Reader) string {
	body, err := ioutil.ReadAll(reader)
	if err != nil {
		return "Something went wrong while parsing the response body: " + err.Error()
	}
	newBody := string(body)

	urls := make(map[string]string)
	z := html.NewTokenizer(ioutil.NopCloser(bytes.NewBuffer(body)))

LOOP:
	for {
		tt := z.Next()

		switch {
		case tt == html.ErrorToken:
			// End of the document, we're done
			break LOOP
		case tt == html.StartTagToken:
			t := z.Token()

			// Check if the token is an <a>, <script>, or <link> tag
			mightHaveLink := t.Data == "a" || t.Data == "script" || t.Data == "link" || t.Data == "img"
			if !mightHaveLink {
				continue
			}

			for _, a := range t.Attr {
				if a.Key == "href" || a.Key == "src" {
					urls[a.Val] = replaceLink(uri, a.Val)
				}
			}
		}
	}

	for key, val := range urls {
		newBody = strings.Replace(newBody, "\""+key+"\"", "\""+val+"\"", -1)
		newBody = strings.Replace(newBody, "'"+key+"'", "'"+val+"'", -1)
	}

	return newBody
}

func replaceLink(baseUri string, oldLink string) string {
	requestedUrl, _ := url.ParseRequestURI(baseUri)
	v := url.Values{}
	u := requestedUrl.Query().Get("u")
	parsedU, _ := url.ParseRequestURI(u)

	if strings.Index(oldLink, "http") == 0 {
		v.Add("u", oldLink)
		requestedUrl.RawQuery = v.Encode()
	} else if strings.Index(oldLink, "//") == 0 {
		v.Add("u", "http:"+oldLink)
		requestedUrl.RawQuery = v.Encode()
	} else if strings.Index(oldLink, "/") == 0 {
		v.Add("u", parsedU.Scheme+"://"+parsedU.Hostname()+oldLink)
		requestedUrl.RawQuery = v.Encode()
	} else if strings.Index(oldLink, "javascript") == 0 {
		return oldLink //Exit early in this case. Nothing to do here
	} else {
		v.Add("u", parsedU.Scheme+"://"+parsedU.Hostname()+parsedU.Path+"/"+oldLink)
		requestedUrl.RawQuery = v.Encode()
	}

	return requestedUrl.String()
}
