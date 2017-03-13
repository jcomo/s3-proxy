package main

import (
	"io"
	"net/http"
	"strconv"
	"time"
)

type HostDispatchingHandler struct {
	hosts map[string]http.Handler
}

func NewHostDispatchingHandler() *HostDispatchingHandler {
	return &HostDispatchingHandler{
		hosts: make(map[string]http.Handler),
	}
}

func (h *HostDispatchingHandler) HandleHost(host string, handler http.Handler) {
	h.hosts[host] = handler
}

func (h *HostDispatchingHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	handler, ok := h.hosts[getHost(r)]
	if !ok {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	handler.ServeHTTP(w, r)
}

func NewBasicAuthHandler(users []User, next http.Handler) http.HandlerFunc {
	m := make(map[string]string)
	for _, u := range users {
		m[u.Name] = u.Password
	}

	return func(w http.ResponseWriter, r *http.Request) {
		username, password, ok := r.BasicAuth()
		if !ok {
			challenge(w, r)
			return
		}

		p, ok := m[username]
		if !ok {
			challenge(w, r)
			return
		}

		if password != p {
			challenge(w, r)
			return
		}

		next.ServeHTTP(w, r)
	}
}

func NewProxyHandler(proxy S3Proxy) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		obj, err := proxy.Get(r.URL.Path)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		setHeader(w, "Cache-Control", s2s(obj.CacheControl))
		setHeader(w, "Content-Disposition", s2s(obj.ContentDisposition))
		setHeader(w, "Content-Encoding", s2s(obj.ContentEncoding))
		setHeader(w, "Content-Language", s2s(obj.ContentLanguage))
		setHeader(w, "Content-Length", i2s(obj.ContentLength))
		setHeader(w, "Content-Range", s2s(obj.ContentRange))
		setHeader(w, "Content-Type", s2s(obj.ContentType))
		setHeader(w, "ETag", s2s(obj.ETag))
		setHeader(w, "Expires", s2s(obj.Expires))
		setHeader(w, "Last-Modified", t2s(obj.LastModified))

		io.Copy(w, obj.Body)
	}
}

func challenge(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("WWW-Authenticate", `Basic realm="`+getHost(r)+`"`)
	http.Error(w, "", http.StatusUnauthorized)
}

func getHost(r *http.Request) string {
	host := r.Header.Get("Host")
	if host == "" {
		host = r.Host
	}

	return host
}

func s2s(s *string) string {
	if s != nil {
		return *s
	} else {
		return ""
	}
}

func i2s(i *int64) string {
	if i != nil {
		return strconv.FormatInt(*i, 10)
	} else {
		return ""
	}
}

func t2s(t *time.Time) string {
	if t != nil {
		return t.UTC().Format(http.TimeFormat)
	} else {
		return ""
	}
}

func setHeader(w http.ResponseWriter, key, value string) {
	if value != "" {
		w.Header().Add(key, value)
	}
}
