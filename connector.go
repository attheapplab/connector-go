package connector

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"
)

const (
	kbody = "body"
	kid = "id"
	kmethod = "method"
	kquery = "query"
	kresource = "resource"
	kresult = "result"
)

type procedure interface {
	Do(context.Context) context.Context
}

type handler map[string]map[string][]procedure

func NewHandler() *handler {
	return &handler{}
}

func (h handler) Handle(method string, resource string, collection ...procedure) {
	if routes, ok := h[method]; ok {
		routes[resource] = collection
		return
	}
	h[method] = make(map[string][]procedure)
	h.Handle(method, resource, collection...)
}

func (h handler) ListenAndServe() {
	port := os.Getenv("PORT")
	if port != "" {
		log.Printf("$PORT=%s\n", port)
	}
	log.Println("Listening...")
	http.ListenAndServe(port, h)
}

func extractBody(r *http.Request) map[string]string {
	bodyParams := make(map[string]string)
	json.NewDecoder(r.Body).Decode(&bodyParams)
	return bodyParams
}

func extractId(r *http.Request) string {
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

func readResult(ctx context.Context) []byte {
	result, ok := ctx.Value(kresult).([]byte)
	if !ok {
		return []byte{}
	}
	return result
}

func logRequest(method string, resource string) {
	log.Printf("[%s] %s\n", method, resource)
}

func parseRequest(r *http.Request) context.Context {
	ctx := r.Context()
	body := extractBody(r)
	ctx = context.WithValue(ctx, kbody, body)
	id := extractId(r)
	ctx = context.WithValue(ctx, kid, id)
	method := extractMethod(r)
	ctx = context.WithValue(ctx, kmethod, method)
	query := extractQuery(r)
	ctx = context.WithValue(ctx, kquery, query)
	resource := extractResource(r)
	ctx = context.WithValue(ctx, kresource, resource)
	return ctx
}

func setHeaders(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:8000")
	w.Header().Set("Content-Type", "application/json")
}

func (h handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	method := extractMethod(r)
	collections, ok := h[method]
	if !ok {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}
	resource := extractResource(r)
	collection, ok := collections[resource]
	if !ok {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}
	logRequest(method, resource)
	ctx := parseRequest(r)
	for _, procedure := range collection {
		ctx = procedure.Do(ctx)
	}
	result := readResult(ctx)
	setHeaders(w, r)
	w.Write(result)
}
