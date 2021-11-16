package rewriter

import (
	"io"
	"log"
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

func RewriteHtml(uri string, reader io.Reader) string {
	doc, err := goquery.NewDocumentFromReader(reader)
	if err != nil {
		log.Fatal(err) //todo
	}

	doc.Find("*").Each(func(i int, s *goquery.Selection) {
		if linkAttr, ok := s.Attr("href"); ok {
			s.SetAttr("href", replaceLink(uri, linkAttr))
		}
		if linkAttr, ok := s.Attr("src"); ok {
			s.SetAttr("src", replaceLink(uri, linkAttr))
		}
		if linkAttr, ok := s.Attr("action"); ok {
			s.SetAttr("action", replaceLink(uri, linkAttr))
		}
	})

	newBody, _ := doc.Html()
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
	} else if strings.Index(oldLink, "javascript") == 0 ||
		strings.Index(oldLink, "data:") == 0 ||
		strings.Index(oldLink, "#") == 0 {
		return oldLink //Exit early in this case. Nothing to do here
	} else {
		v.Add("u", parsedU.Scheme+"://"+parsedU.Hostname()+parsedU.Path+"/"+oldLink)
		requestedUrl.RawQuery = v.Encode()
	}

	return requestedUrl.String()
}
