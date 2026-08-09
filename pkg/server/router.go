package server

import "net/http"

// Router defines a framework‑agnostic HTTP router interface (the *port*).
// Handlers receive a Context that abstracts the underlying request/response.
// This allows us to swap Gin, Echo, Fiber, etc. without touching the
// bootstrap or use‑case layers.

type Router interface {
	// GET registers a handler for HTTP GET requests.
	GET(path string, handler HandlerFunc)
	// POST registers a handler for HTTP POST requests.
	POST(path string, handler HandlerFunc)
	// PUT registers a handler for HTTP PUT requests.
	PUT(path string, handler HandlerFunc)
	// DELETE registers a handler for HTTP DELETE requests.
	DELETE(path string, handler HandlerFunc)
	// PATCH registers a handler for HTTP PATCH requests.
	PATCH(path string, handler HandlerFunc)
	// HEAD registers a handler for HTTP HEAD requests.
	HEAD(path string, handler HandlerFunc)
	// Group creates a sub‑router with a common prefix.
	Group(path string) RouterGroup
	// Use registers one or more middleware functions globally.
	Use(middleware ...MiddlewareFunc)
}

// RouterGroup is a sub‑router that can also register middleware.
type RouterGroup interface {
	GET(path string, handler HandlerFunc)
	POST(path string, handler HandlerFunc)
	PUT(path string, handler HandlerFunc)
	DELETE(path string, handler HandlerFunc)
	PATCH(path string, handler HandlerFunc)
	// HEAD registers a handler for HTTP HEAD requests.
	HEAD(path string, handler HandlerFunc)
	// Use registers one or more middleware functions for this group.
	Use(middleware ...MiddlewareFunc)
	// Group creates a sub-router with a common prefix.
	Group(path string) RouterGroup
}

// Context abstracts the request/response handling.
// It mirrors a subset of gin.Context that we need for our handlers.
type Context interface {
	// Param returns a path parameter (e.g. ":id").
	Param(key string) string
	// Query returns a query string parameter.
	Query(key string) string
	// Bind parses the request body into the provided struct.
	Bind(obj interface{}) error
	// BindQuery parses and validates query parameters into the provided struct.
	BindQuery(obj interface{}) error
	// BindURI parses and validates path parameters into the provided struct.
	BindURI(obj interface{}) error
	// JSON writes a JSON response with the given HTTP status code.
	JSON(code int, obj interface{})
	// Status writes the HTTP status code with no body.
	Status(code int)
	// Request returns the underlying *http.Request.
	Request() *http.Request
	// SetHeader sets a response header.
	SetHeader(key, value string)
	// SetCookie sets a cookie in the response. Parameters mirror http.SetCookie
	//
	// name, value: application data to store in the cookie
	// maxAge: in seconds, if not set, once browser closed, cookie will be deleted
	// path: the URL path for which the cookie is valid, e.g "/admin" or "/"
	// secure: if true then cookie only sent over HTTPS
	// httpOnly: if true then cookie is inaccessible to JavaScript (prevent XSS)
	SetCookie(name, value string, maxAge int, path string, secure, httpOnly bool)
	Cookie(name string) (string, error)
	SetValue(key string, value interface{})
	GetValue(key string) interface{}
}

// HandlerFunc is the signature for route handlers.
type HandlerFunc func(Context)

// MiddlewareFunc is the signature for middleware.
type MiddlewareFunc func(Context) (Context, error)
