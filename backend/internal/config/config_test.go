package config

import "testing"

func TestLoadUsesPlatformPortBeforeLocalPort(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://test:test@localhost/test?sslmode=disable")
	t.Setenv("PORT", "10000")
	t.Setenv("BACKEND_PORT", "8080")
	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Port != "10000" {
		t.Fatalf("got port %q, want platform port 10000", cfg.Port)
	}
}

func TestLoadFallsBackToLocalPort(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://test:test@localhost/test?sslmode=disable")
	t.Setenv("PORT", "")
	t.Setenv("BACKEND_PORT", "8080")
	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Port != "8080" {
		t.Fatalf("got port %q, want local port 8080", cfg.Port)
	}
}
