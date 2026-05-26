package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// GinRouter implements the Router interface using the Gin framework.
type GinRouter struct {
	engine *gin.Engine
}

// NewGinRouter creates a new GinRouter instance.
func NewGinRouter(engine *gin.Engine) *GinRouter {
	return &GinRouter{engine: engine}
}

func (r *GinRouter) Engine() *gin.Engine {
	return r.engine
}

func (r *GinRouter) GET(path string, handler HandlerFunc) {
	r.engine.GET(path, wrapHandler(handler))
}

func (r *GinRouter) POST(path string, handler HandlerFunc) {
	r.engine.POST(path, wrapHandler(handler))
}

func (r *GinRouter) PUT(path string, handler HandlerFunc) {
	r.engine.PUT(path, wrapHandler(handler))
}

func (r *GinRouter) DELETE(path string, handler HandlerFunc) {
	r.engine.DELETE(path, wrapHandler(handler))
}

func (r *GinRouter) PATCH(path string, handler HandlerFunc) {
	r.engine.PATCH(path, wrapHandler(handler))
}

func (r *GinRouter) Group(path string) RouterGroup {
	return &GinRouterGroup{group: r.engine.Group(path)}
}

func (r *GinRouter) Use(middleware ...MiddlewareFunc) {
	for _, m := range middleware {
		r.engine.Use(wrapMiddleware(m))
	}
}

// GinRouterGroup implements the RouterGroup interface.
type GinRouterGroup struct {
	group *gin.RouterGroup
}

func (g *GinRouterGroup) GET(path string, handler HandlerFunc) {
	g.group.GET(path, wrapHandler(handler))
}

func (g *GinRouterGroup) POST(path string, handler HandlerFunc) {
	g.group.POST(path, wrapHandler(handler))
}

func (g *GinRouterGroup) PUT(path string, handler HandlerFunc) {
	g.group.PUT(path, wrapHandler(handler))
}

func (g *GinRouterGroup) DELETE(path string, handler HandlerFunc) {
	g.group.DELETE(path, wrapHandler(handler))
}

func (g *GinRouterGroup) PATCH(path string, handler HandlerFunc) {
	g.group.PATCH(path, wrapHandler(handler))
}

func (g *GinRouterGroup) Use(middleware ...MiddlewareFunc) {
	for _, m := range middleware {
		g.group.Use(wrapMiddleware(m))
	}
}

func (g *GinRouterGroup) Group(path string) RouterGroup {
	return &GinRouterGroup{group: g.group.Group(path)}
}

// GinContext implements the Context interface.
type GinContext struct {
	ctx *gin.Context
}

func (c *GinContext) Param(key string) string {
	return c.ctx.Param(key)
}

func (c *GinContext) Query(key string) string {
	return c.ctx.Query(key)
}

func (c *GinContext) Bind(obj interface{}) error {
	return c.ctx.ShouldBind(obj)
}

func (c *GinContext) JSON(code int, obj interface{}) {
	c.ctx.JSON(code, obj)
}

func (c *GinContext) Status(code int) {
	c.ctx.Status(code)
}

func (c *GinContext) Request() *http.Request {
	return c.ctx.Request
}

func (c *GinContext) SetHeader(key, value string) {
	c.ctx.Header(key, value)
}

func (c *GinContext) SetCookie(name, value string, maxAge int, path string, secure, httpOnly bool) {
	c.ctx.SetCookie(name, value, maxAge, path, "", secure, httpOnly)
}

func (c *GinContext) Cookie(name string) (string, error) {
	cookie, err := c.ctx.Cookie(name)
	if err != nil {
		return "", err
	}
	return cookie, nil
}

func (c *GinContext) SetValue(key string, value interface{}) {
	c.ctx.Set(key, value)
}

func (c *GinContext) GetValue(key string) interface{} {
	val, _ := c.ctx.Get(key)
	return val
}

// Helper functions to wrap handlers and middleware

func wrapHandler(h HandlerFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		h(&GinContext{ctx: c})
	}
}

func wrapMiddleware(m MiddlewareFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, err := m(&GinContext{ctx: c})
		if err != nil {
			if httpErr, ok := err.(*HTTPError); ok {
				c.AbortWithStatusJSON(httpErr.Code, gin.H{"error": httpErr.Message})
			} else {
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			}
			return
		}
		if ctx != nil {
			// In some cases we might want to update the context,
			// but for Gin we usually just continue.
		}
		c.Next()
	}
}
