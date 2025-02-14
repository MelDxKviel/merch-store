package config

import (
	"testing"
)

func TestLoadConfig_Success(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://user:pass@localhost/db")
	t.Setenv("JWT_SECRET", "mysecret")

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("Ожидалась ошибка nil, получена: %v", err)
	}

	if cfg.DatabaseURL != "postgres://user:pass@localhost/db" {
		t.Errorf("Ожидалось DATABASE_URL %s, получено %s", "postgres://user:pass@localhost/db", cfg.DatabaseURL)
	}
	if cfg.JWTSecret != "mysecret" {
		t.Errorf("Ожидалось JWT_SECRET %s, получено %s", "mysecret", cfg.JWTSecret)
	}
}

func TestLoadConfig_NoDatabaseURL(t *testing.T) {
	t.Setenv("DATABASE_URL", "")
	t.Setenv("JWT_SECRET", "mysecret")

	cfg, err := LoadConfig()
	if err == nil {
		t.Fatalf("Ожидалась ошибка из-за отсутствия DATABASE_URL, получена nil")
	}
	if cfg != nil {
		t.Errorf("Ожидалось, что cfg будет nil при ошибке, получено: %v", cfg)
	}
}

func TestLoadConfig_NoJWTSecret(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://user:pass@localhost/db")
	t.Setenv("JWT_SECRET", "")

	cfg, err := LoadConfig()
	if err == nil {
		t.Fatalf("Ожидалась ошибка из-за отсутствия JWT_SECRET, получена nil")
	}
	if cfg != nil {
		t.Errorf("Ожидалось, что cfg будет nil при ошибке, получено: %v", cfg)
	}
}
