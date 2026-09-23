package config

import (
	"strings"
	"testing"
)

func TestLoadJWTSecret(t *testing.T) {
	tests := []struct {
		name    string
		secret  string
		wantErr string
	}{
		{name: "empty secret is rejected", secret: "", wantErr: "SECURITY_JWT_SECRET"},
		{name: "short secret is rejected", secret: "too-short", wantErr: "at least 32"},
		{name: "32 char secret is accepted", secret: strings.Repeat("a", 32)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("DATABASE_PASSWORD", "x")
			t.Setenv("SECURITY_JWT_SECRET", tt.secret)

			_, err := Load()
			if tt.wantErr == "" {
				if err != nil {
					t.Fatalf("Load() error = %v, want nil", err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("Load() error = %v, want error containing %q", err, tt.wantErr)
			}
		})
	}
}

func TestLoadTrustedProxies(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		want    []string
		wantErr bool
	}{
		{name: "empty means none", value: "", want: []string{}},
		{name: "cidr and ip with blanks", value: " 10.42.0.0/16, ,10.0.0.1,", want: []string{"10.42.0.0/16", "10.0.0.1"}},
		{name: "garbage is rejected", value: "not-an-ip", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("DATABASE_PASSWORD", "x")
			t.Setenv("SECURITY_JWT_SECRET", strings.Repeat("a", 32))
			t.Setenv("SERVER_TRUSTED_PROXIES", tt.value)

			cfg, err := Load()
			if tt.wantErr {
				if err == nil {
					t.Fatal("Load() error = nil, want error")
				}
				return
			}
			if err != nil {
				t.Fatalf("Load() error = %v", err)
			}
			if strings.Join(cfg.Server.TrustedProxies, "|") != strings.Join(tt.want, "|") {
				t.Fatalf("TrustedProxies = %q, want %q", cfg.Server.TrustedProxies, tt.want)
			}
		})
	}
}
