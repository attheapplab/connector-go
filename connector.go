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
	kGET = "GET"
	kPATCH = "PATCH"
	kPOST = "POST"
)

type procedure interface {
	Do(context.Context, http.ResponseWriter) context.Context
}

type Handler struct {
	origin string
	port   string
	tree   map[string]*node
}

func New() *Handler {
	return &Handler{
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

func (h *Handler) Handle(method string, path string, procedures ...procedure) {
	if !isPath(path) {
		panic(path)
	}
	root, ok := h.tree[method]
	if !ok {
		root := &node{}
		h.tree[method] = root
		h.Handle(method, path, procedures...)
		return
	}
	segments := splitPath(path)
	node, next := root.traverse(segments)
	node.add(next, procedures)
}

func (h *Handler) Delete(path string, procedures ...procedure) {
	h.Handle(kDELETE, path, procedures...)
}

func (h *Handler) Get(path string, procedures ...procedure) {
	h.Handle(kGET, path, procedures...)
}

func (h *Handler) Patch(path string, procedures ...procedure) {
	h.Handle(kPATCH, path, procedures...)
}

func (h *Handler) Post(path string, procedures ...procedure) {
	h.Handle(kPOST, path, procedures...)
}

func (h *Handler) ListenAndServe() {
	fmt.Println("Listening...")
	http.ListenAndServe(h.port, h)
}

func (h *Handler) ListenAndServeTLS(certfile string, keyfile string) {
	fmt.Println("Listening on TLS...")
	http.ListenAndServeTLS(h.port, certfile, keyfile, h)
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {	
	nodes, ok := h.tree[r.Method]
	if !ok && r.Method == http.MethodOptions {
		h.setCORSHeaders(w)
		return
	} else if !ok {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}
	segments := splitPath(r.URL.Path)
	node, _ := nodes.traverse(segments)
	if !node.isMatch {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}
	h.setCORSHeaders(w)
	ctx := extractContext(r)
	ctx = parseRequest(ctx, r)
	defer catchErrors(w)
	for _, procedure := range node.procedures {
		ctx = procedure.Do(ctx, w)
	}
}

func (h *Handler) setAccessControlAllowCredentials(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Credentials", "true")
}

func (h *Handler) setAccessControlAllowMethods(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Methods", "DELETE, PATCH")
}

func (h *Handler) setAccessControlAllowOrigin(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", h.origin)
}

func (h *Handler) setCORSHeaders(w http.ResponseWriter) {
	h.setAccessControlAllowCredentials(w)
	h.setAccessControlAllowOrigin(w)
	h.setAccessControlAllowMethods(w)
}

func (h *Handler) SetAccessControlOrigin(origin string) {
	h.origin = origin
}

func (h *Handler) SetPort(port string) {
	h.port = port
}
