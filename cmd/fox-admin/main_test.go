package main

import (
	"slices"
	"testing"

	systemmiddleware "fox-admin/internal/middleware"

	"github.com/surge-go/fox"
)

func TestNewCORSConfigAllowsAndExposesAuthHeaders(t *testing.T) {
	cfg := newCORSConfig()
	for _, header := range []string{
		systemmiddleware.DefaultDeviceIDHeaderName,
		systemmiddleware.DefaultRefreshHeaderName,
	} {
		if !slices.Contains(cfg.AllowHeaders, header) {
			t.Fatalf("AllowHeaders = %#v, want %s", cfg.AllowHeaders, header)
		}
	}
	for _, header := range []string{
		systemmiddleware.DefaultAccessResponseHeaderName,
		systemmiddleware.DefaultRefreshResponseHeaderName,
		systemmiddleware.AccessExpiresAtHeaderName,
		systemmiddleware.RefreshExpiresAtHeaderName,
		systemmiddleware.TokenTypeHeaderName,
	} {
		if !slices.Contains(cfg.ExposeHeaders, header) {
			t.Fatalf("ExposeHeaders = %#v, want %s", cfg.ExposeHeaders, header)
		}
	}
}

func TestRegisterSwaggerRoutes(t *testing.T) {
	printRoutes := false
	engine := fox.New(&fox.Config{
		Addr:        ":0",
		Mode:        fox.ModeTest,
		PrintRoutes: &printRoutes,
	})

	registerSwaggerRoutes(engine)
}
