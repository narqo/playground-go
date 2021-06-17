package main

import (
	"os"
	"testing"
)

func TestIntegrationFoo(t *testing.T) {
	if os.Getenv("INTEGRATION_TEST") == "" {
		t.Skip("INTEGRATION_TEST is not set")
	}

	dsn := os.Getenv("POSTGRES_DSN")
	if dsn == "" {
		t.Fatal("POSTGRES_DSN is not set")
	}
	t.Logf("postgres dsn %q", dsn)
}
