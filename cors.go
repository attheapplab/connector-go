package connector

import (
	"net/http"
)

type cors struct {
	h                http.Handler
	allowedMethods   []string
	allowedOrigins   []string
	allowCredentials bool
	optionStatusCode int
}

var (
	defaultCorsOptionStatusCode = 200
	defaultCorsMethods          = []string{"GET", "HEAD", "POST"}
	defaultCorsOrigins          = []string{}
)

const (
	corsOptionMethod           = "OPTIONS"
	corsAllowOriginHeader      = "Access-Control-Allow-Origin"
	corsAllowMethodsHeader     = "Access-Control-Allow-Methods"
	corsAllowCredentialsHeader = "Access-Control-Allow-Credentials"
	corsRequestMethodHeader    = "Access-Control-Request-Method"
	corsOriginHeader           = "Origin"
	corsOriginMatchAll         = "*"
)

func (ch *cors) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	origin := r.Header.Get(corsOriginHeader)
	if !ch.isOriginAllowed(origin) {
		if r.Method != corsOptionMethod {
			ch.h.ServeHTTP(w, r)
		}
		return
	}

	if r.Method == corsOptionMethod {
		if _, ok := r.Header[corsRequestMethodHeader]; !ok {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		method := r.Header.Get(corsRequestMethodHeader)
		if !ch.isMatch(method, ch.allowedMethods) {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		if !ch.isMatch(method, defaultCorsMethods) {
			w.Header().Set(corsAllowMethodsHeader, method)
		}
	}

	if ch.allowCredentials {
		w.Header().Set(corsAllowCredentialsHeader, "true")
	}

	returnOrigin := origin
	if len(ch.allowedOrigins) == 0 {
		returnOrigin = "*"
	}
	w.Header().Set(corsAllowOriginHeader, returnOrigin)

	if r.Method == corsOptionMethod {
		w.WriteHeader(ch.optionStatusCode)
		return
	}
	ch.h.ServeHTTP(w, r)
}

func (ch *cors) isOriginAllowed(origin string) bool {
	if origin == "" {
		return false
	}

	if len(ch.allowedOrigins) == 0 {
		return true
	}

	for _, allowedOrigin := range ch.allowedOrigins {
		if allowedOrigin == origin || allowedOrigin == corsOriginMatchAll {
			return true
		}
	}

	return false
}

func (ch *cors) isMatch(needle string, haystack []string) bool {
	for _, v := range haystack {
		if v == needle {
			return true
		}
	}
	return false
}

func (r *Router) UseCORS() http.Handler {
	r.ch.h = r
	return r.ch
}

func (r *Router) AllowCredentials() {
	r.ch.allowCredentials = true
}

func (r *Router) AllowMethods(methods ...string) {
	r.ch.allowedMethods = defaultCorsMethods
	for _, method := range methods {
		if r.ch.isMatch(method, r.ch.allowedMethods) {
			continue
		}
		r.ch.allowedMethods = append(r.ch.allowedMethods, method)
	}
}

func (r *Router) AllowOrigins(origins ...string) {
	for _, v := range origins {
		if v == corsOriginMatchAll {
			r.ch.allowedOrigins = []string{corsOriginMatchAll}
		}
	}
	r.ch.allowedOrigins = origins
}
