package database

import (
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/quoctann/content-hub/pkg/config"
)

func TestURLRoundTripsSpecialCharacters(t *testing.T) {
	cfg := config.Database{
		Host:     "db.local",
		Port:     "5432",
		User:     "app",
		Password: "p@ss:w/rd?#1 x",
		Name:     "postgres",
		SSLMode:  "disable",
		Schema:   "meme",
	}

	poolCfg, err := pgxpool.ParseConfig(URL(cfg))
	if err != nil {
		t.Fatalf("ParseConfig(URL) error = %v", err)
	}
	conn := poolCfg.ConnConfig
	if conn.Password != cfg.Password {
		t.Errorf("password = %q, want %q", conn.Password, cfg.Password)
	}
	if conn.Host != cfg.Host || conn.Port != 5432 || conn.User != cfg.User || conn.Database != cfg.Name {
		t.Errorf("parsed %s@%s:%d/%s, want %s@%s:5432/%s", conn.User, conn.Host, conn.Port, conn.Database, cfg.User, cfg.Host, cfg.Name)
	}
	if got := conn.RuntimeParams["search_path"]; got != "meme" {
		t.Errorf("search_path = %q, want %q", got, "meme")
	}
}

func TestURLOmitsSearchPathWithoutSchema(t *testing.T) {
	poolCfg, err := pgxpool.ParseConfig(URL(config.Database{Host: "h", Port: "5432", User: "u", Password: "p", Name: "d", SSLMode: "disable"}))
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := poolCfg.ConnConfig.RuntimeParams["search_path"]; ok {
		t.Error("search_path should not be set when Schema is empty")
	}
}
