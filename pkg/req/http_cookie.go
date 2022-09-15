package req

import (
	"bytes"
	"io"
	"log"
	"net/http"
	"net/http/cookiejar"
)

type HTTPCookie struct {
	Client   *http.Client
	Request  *http.Request
	Response *http.Response

	url string

	CookiesJar *cookiejar.Jar
	Cookies    []*http.Cookie
}

func (h *HTTPCookie) Get(data []byte) *HTTPCookie {
	h.Request, _ = http.NewRequest("GET", h.url, bytes.NewReader(data))
	return h
}

func (h *HTTPCookie) Do() *HTTPCookie {
	h.Response, _ = h.Client.Do(h.Request)
	h.Cookies = h.CookiesJar.Cookies(h.Request.URL)
	log.Println(h.Cookies)
	return h
}

// Post
func (h *HTTPCookie) Post(data io.Reader, contentType string) *HTTPCookie {
	h.Request, _ = http.NewRequest("POST", h.url, data)
	switch contentType {
	case "urlencoded":
		h.Request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	default:
		h.Request.Header.Add("Content-Type", "application/json")
	}
	return h
}

func (h *HTTPCookie) SetCookie(r *http.Request) {
	for _, c := range h.Cookies {
		r.AddCookie(c)
	}
}

func (h *HTTPCookie) GetCookies(name string) *http.Cookie {
	for _, c := range h.Cookies {
		if c.Name == name {
			return c
		}
	}
	return nil
}

func (h *HTTPCookie) SetHeader(name string) string {
	return h.Response.Header.Get(name)
}

func (h *HTTPCookie) Url(url string) *HTTPCookie {
	h.url = url
	h.Request, _ = http.NewRequest("GET", h.url, nil)
	return h
}

func (h *HTTPCookie) ToJson() *HTTPCookie {
	return h
}
