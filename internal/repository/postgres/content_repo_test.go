package postgres

import "testing"

func TestContentOrderSQL(t *testing.T) {
	tests := []struct {
		name   string
		hasFTS bool
		want   string
	}{
		{
			name: "browse order",
			want: " ORDER BY c.created_at DESC, c.id DESC",
		},
		{
			name:   "full text order",
			hasFTS: true,
			want:   " ORDER BY rank DESC, c.created_at DESC, c.id DESC",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := contentOrderSQL(tt.hasFTS); got != tt.want {
				t.Errorf("contentOrderSQL() = %q, want %q", got, tt.want)
			}
		})
	}
}
