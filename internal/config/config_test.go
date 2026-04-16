package config

import (
	"testing"
)

func TestDatabaseURLWithName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		databaseURL string
		dbName      string
		want        string
	}{
		{
			name:        "ホスト・ポート・クエリを維持してDB名だけ置換",
			databaseURL: "postgres://postgres:postgres@localhost:5432/inventory?sslmode=disable",
			dbName:      "userdb",
			want:        "postgres://postgres:postgres@localhost:5432/userdb?sslmode=disable",
		},
		{
			name:        "本番ホストでもDB名が正しく置換される",
			databaseURL: "postgres://appuser:secret@db.example.com:5432/maindb?sslmode=require",
			dbName:      "orderdb",
			want:        "postgres://appuser:secret@db.example.com:5432/orderdb?sslmode=require",
		},
		{
			name:        "デフォルトURL（ポートなし）でもDB名が置換される",
			databaseURL: "postgres://postgres:postgres@localhost/inventory",
			dbName:      "paymentdb",
			want:        "postgres://postgres:postgres@localhost/paymentdb",
		},
		{
			name:        "無効なURLのときはフォールバックURLを返す",
			databaseURL: "://invalid-url",
			dbName:      "fallbackdb",
			want:        "postgres://postgres:postgres@localhost:5432/fallbackdb?sslmode=disable",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			c := Config{DatabaseURL: tt.databaseURL}
			got := c.DatabaseURLWithName(tt.dbName)
			if got != tt.want {
				t.Errorf("DatabaseURLWithName(%q)\n got  %q\n want %q", tt.dbName, got, tt.want)
			}
		})
	}
}

func TestDatabaseURLWithName_DoesNotMutateOriginalURL(t *testing.T) {
	t.Parallel()

	original := "postgres://postgres:postgres@localhost:5432/inventory?sslmode=disable"
	c := Config{DatabaseURL: original}

	c.DatabaseURLWithName("otherdb")

	if c.DatabaseURL != original {
		t.Errorf("DatabaseURL was mutated: got %q, want %q", c.DatabaseURL, original)
	}
}
