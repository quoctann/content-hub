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
