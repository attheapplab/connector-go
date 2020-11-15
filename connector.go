package connector

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
)

const (
	k401 = "401" // Unauthorized
	k409 = "409" // Conflict
	k500 = "500" // Internal Server Error
)

const (
	kbody = "body"
	kcookies = "cookies"
	kidentifier = "identifier"
	kmethod = "method"
	kquery = "query"
	kresource = "resource"
)

type procedure interface {
	Do(context.Context, http.ResponseWriter) context.Context
}

type connector map[string]map[string][]procedure

func New() *connector {
	return &connector{}
}

func (c connector) Handle(method string, resource string, collection ...procedure) {
	routes, ok := c[method]
	if !ok {
		c[method] = make(map[string][]procedure)
		c.Handle(method, resource, collection...)
		return
	}
	routes[resource] = collection
}

func (c connector) ListenAndServe() {
	port := os.Getenv("PORT")
	if port != "" {
		log.Printf("$PORT=%s\n", port)
	}
	fmt.Println("Listening...")
	http.ListenAndServe(port, c)
}

func createContext(r *http.Request) context.Context {
	return r.Context()
}

func extractBody(r *http.Request) map[string]string {
	bodyParams := make(map[string]string)
	json.NewDecoder(r.Body).Decode(&bodyParams)
	return bodyParams
}

func extractCookies(r *http.Request) map[string]string {
	cookies := make(map[string]string)
	for _, cookie := range r.Cookies() {
		cookies[cookie.Name] = cookie.Value
	}
	return cookies
}

func extractIdentifier(r *http.Request) string {
	path := strings.Split(r.URL.Path, "/")
	if len(path) < 3 {
		return ""
	}
	return path[2]
}

func extractMethod(r *http.Request) string {
	return r.Method
}

func extractQuery(r *http.Request) map[string]string {
	queryParams := make(map[string]string)
	query := r.URL.Query()
	for key, _ := range query {
		queryParams[key] = query.Get(key)
	}
	return queryParams
}

func extractResource(r *http.Request) string {
	return strings.Split(r.URL.Path, "/")[1]
}

func parseRequest(ctx context.Context, r *http.Request) context.Context {
	body := extractBody(r)
	ctx = context.WithValue(ctx, kbody, body)
	cookies := extractCookies(r)
	ctx = context.WithValue(ctx, kcookies, cookies)
	identifier := extractIdentifier(r)
	ctx = context.WithValue(ctx, kidentifier, identifier)
	method := extractMethod(r)
	ctx = context.WithValue(ctx, kmethod, method)
	query := extractQuery(r)
	ctx = context.WithValue(ctx, kquery, query)
	resource := extractResource(r)
	ctx = context.WithValue(ctx, kresource, resource)
	return ctx
}

func handleErrors(w http.ResponseWriter) {
	switch r := recover(); r {
	case k401:
		send401(w)
	case k409:
		send409(w)
	case k500:
		send500(w)
	}
}

func logError(status string) {
	log.Println(status)
}

func send401(w http.ResponseWriter) {
	status := http.StatusText(http.StatusUnauthorized)
	logError(status)
	http.Error(w, status, http.StatusUnauthorized)
}

func send409(w http.ResponseWriter) {
	status := http.StatusText(http.StatusConflict)
	logError(status)
	http.Error(w, status, http.StatusConflict)
}

func send500(w http.ResponseWriter) {
	status := http.StatusText(http.StatusInternalServerError)
	logError(status)
	http.Error(w, status, http.StatusInternalServerError)
}

func setCORSHeaders(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:8000")
	w.Header().Set("Access-Control-Allow-Methods", "PATCH")
}

func (c connector) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	method := extractMethod(r)
	resource := extractResource(r)
	collections, ok := c[method]
	if !ok && method == http.MethodOptions {
		setCORSHeaders(w)
		return
	} else if !ok {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}
	collection, ok := collections[resource]
	if !ok {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}
	setCORSHeaders(w)
	ctx := createContext(r)
	ctx = parseRequest(ctx, r)
	defer handleErrors(w)
	for _, procedure := range collection {
		ctx = procedure.Do(ctx, w)
	}
}
