package database

import (
	"net"
	"net/url"

	"github.com/quoctann/content-hub/pkg/config"
)

// URL builds the postgres:// connection URL used by both the pgx pool and the
// migrator, so they can never disagree. Credentials go through url.UserPassword,
// so passwords containing @ : / ? # or spaces are escaped instead of silently
// breaking the URL. A configured schema is sent as the search_path startup
// parameter, so every connection resolves unqualified names in that schema.
func URL(cfg config.Database) string {
	query := url.Values{}
	query.Set("sslmode", cfg.SSLMode)
	if cfg.Schema != "" {
		query.Set("search_path", cfg.Schema)
	}

	u := url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword(cfg.User, cfg.Password),
		Host:     net.JoinHostPort(cfg.Host, cfg.Port),
		Path:     "/" + cfg.Name,
		RawQuery: query.Encode(),
	}
	return u.String()
}
