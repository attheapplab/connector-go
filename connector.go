package connector

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
)

const (
	k401 = "401" // Unauthorized
	k409 = "409" // Conflict
	k500 = "500" // Internal Server Error
)

// Write to context:
const (
	kbody       = "body"
	kcookies    = "cookies"
	kidentifier = "identifier"
	kmethod     = "method"
	kquery      = "query"
	kresource   = "resource"
)

const (
	kDELETE = "DELETE"
	kGET    = "GET"
	kPATCH  = "PATCH"
	kPOST   = "POST"
)

type procedure interface {
	Do(context.Context, http.ResponseWriter) context.Context
}

type Router struct {
	ch            *cors
	listeningPort string
	tree          map[string]*node
}

func New() *Router {
	return &Router{
		ch: &cors{
			allowedMethods:   defaultCorsMethods,
			allowedOrigins:   defaultCorsOrigins,
			optionStatusCode: defaultCorsOptionStatusCode,
		},
		tree: make(map[string]*node),
	}
}

func extractBody(r *http.Request) map[string]interface{} {
	bodyParams := make(map[string]interface{})
	json.NewDecoder(r.Body).Decode(&bodyParams)
	return bodyParams
}

func extractContext(r *http.Request) context.Context {
	return r.Context()
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

func extractQuery(r *http.Request) map[string]interface{} {
	queryParams := make(map[string]interface{})
	query := r.URL.Query()
	for key, _ := range query {
		queryParams[key] = query.Get(key)
	}
	return queryParams
}

func extractResource(r *http.Request) string {
	return strings.Split(r.URL.Path, "/")[1]
}

func writeBody(ctx context.Context, body map[string]interface{}) context.Context {
	return context.WithValue(ctx, kbody, body)
}

func writeCookies(ctx context.Context, cookies map[string]string) context.Context {
	return context.WithValue(ctx, kcookies, cookies)
}

func writeIdentifier(ctx context.Context, identifier string) context.Context {
	return context.WithValue(ctx, kidentifier, identifier)
}

func writeMethod(ctx context.Context, method string) context.Context {
	return context.WithValue(ctx, kmethod, method)
}

func writeQuery(ctx context.Context, query map[string]interface{}) context.Context {
	return context.WithValue(ctx, kquery, query)
}

func writeResource(ctx context.Context, resource string) context.Context {
	return context.WithValue(ctx, kresource, resource)
}

func parseRequest(ctx context.Context, r *http.Request) context.Context {
	body := extractBody(r)
	ctx = writeBody(ctx, body)
	cookies := extractCookies(r)
	ctx = writeCookies(ctx, cookies)
	identifier := extractIdentifier(r)
	ctx = writeIdentifier(ctx, identifier)
	method := extractMethod(r)
	ctx = writeMethod(ctx, method)
	query := extractQuery(r)
	ctx = writeQuery(ctx, query)
	resource := extractResource(r)
	ctx = writeResource(ctx, resource)
	return ctx
}

func catchErrors(w http.ResponseWriter) {
	err := recover()
	if err == nil {
		return
	}
	switch err {
	case k401:
		send401(w)
	case k409:
		send409(w)
	case k500:
		send500(w)
	default:
		fmt.Println(err)
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

func isPath(path string) bool {
	return len(path) > 0 && path[0:1] == "/"
}

func splitPath(path string) []string {
	return strings.Split(path, "/")[1:]
}

func (r *Router) Handle(method string, path string, procedures ...procedure) {
	if !isPath(path) {
		panic(path)
	}
	root, ok := r.tree[method]
	if !ok {
		root := &node{}
		r.tree[method] = root
		r.Handle(method, path, procedures...)
		return
	}
	segments := splitPath(path)
	node, next := root.traverse(segments)
	node.add(next, procedures)
}

func (r *Router) Delete(path string, procedures ...procedure) {
	r.Handle(kDELETE, path, procedures...)
}

func (r *Router) Get(path string, procedures ...procedure) {
	r.Handle(kGET, path, procedures...)
}

func (r *Router) Patch(path string, procedures ...procedure) {
	r.Handle(kPATCH, path, procedures...)
}

func (r *Router) Post(path string, procedures ...procedure) {
	r.Handle(kPOST, path, procedures...)
}

func (r *Router) ListenAndServe() {
	handler := r.UseCORS()
	fmt.Println("Listening...")
	http.ListenAndServe(r.listeningPort, handler)
}

func (r *Router) ListenAndServeTLS(certfile string, keyfile string) {
	handler := r.UseCORS()
	fmt.Println("Listening on TLS...")
	http.ListenAndServeTLS(r.listeningPort, certfile, keyfile, handler)
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	nodes, ok := r.tree[req.Method]
	if !ok {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}

	segments := splitPath(req.URL.Path)
	node, _ := nodes.traverse(segments)
	if !node.isMatch {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}

	ctx := extractContext(req)
	ctx = parseRequest(ctx, req)
	defer catchErrors(w)
	for _, procedure := range node.procedures {
		ctx = procedure.Do(ctx, w)
	}
}

func (r *Router) SetListeningPort(listeningPort string) {
	r.listeningPort = listeningPort
}
