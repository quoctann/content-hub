package middleware

import (
	"fmt"
	"strings"

	"github.com/quoctann/content-hub/pkg/server"
)

type CSPOptions struct {
	DefaultSrc string
	ScriptSrc  string
	StyleSrc   string
	ImgSrc     string
	FontSrc    string
	ConnectSrc string
	FrameSrc   string
	MediaSrc   string
	ObjectSrc  string
}

func SecurityHeaders(opts CSPOptions) server.MiddlewareFunc {
	csp := fmt.Sprintf(
		"default-src %s; script-src %s; style-src %s; img-src %s; font-src %s; connect-src %s; frame-src %s; media-src %s; object-src %s; frame-ancestors 'none'",
		buildDirective("'self'", opts.DefaultSrc),
		buildDirective("'self'", opts.ScriptSrc),
		buildDirective("'self' 'unsafe-inline'", opts.StyleSrc),
		buildDirective("'self' data: https:", opts.ImgSrc),
		buildDirective("'self'", opts.FontSrc),
		buildDirective("'self'", opts.ConnectSrc),
		buildDirective("", opts.FrameSrc),
		buildDirective("'self'", opts.MediaSrc),
		buildDirective("'none'", opts.ObjectSrc),
	)

	return func(c server.Context) (server.Context, error) {
		c.SetHeader("X-Content-Type-Options", "nosniff")
		c.SetHeader("X-Frame-Options", "DENY")
		c.SetHeader("Referrer-Policy", "strict-origin-when-cross-origin")
		c.SetHeader("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		c.SetHeader("Permissions-Policy", "geolocation=(), microphone=(), camera=()")
		c.SetHeader("Content-Security-Policy", csp)
		return c, nil
	}
}

func buildDirective(base string, extra string) string {
	result := base
	if extra == "" {
		return result
	}
	for _, s := range strings.Split(extra, ",") {
		if trimmed := strings.TrimSpace(s); trimmed != "" {
			result += " " + trimmed
		}
	}
	return result
}
