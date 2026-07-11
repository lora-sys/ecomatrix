package http

import (
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

func newAppWithCORS(cfg corsConfig) *fiber.App {
	app := fiber.New()
	app.Use(corsMiddleware(cfg, nil))
	app.Get("/ping", func(c *fiber.Ctx) error { return c.SendString("pong") })
	return app
}

func TestCORS_Wildcard_AllowsAnyOrigin(t *testing.T) {
	app := newAppWithCORS(corsConfig{AllowedOrigins: []string{"*"}})
	req := httptest.NewRequest("GET", "/ping", nil)
	req.Header.Set("Origin", "http://anywhere.example")
	resp, _ := app.Test(req)
	assert.Equal(t, 200, resp.StatusCode)
	assert.Equal(t, "http://anywhere.example", resp.Header.Get("Access-Control-Allow-Origin"))
}

func TestCORS_ExactMatch_AllowsListedOrigin(t *testing.T) {
	app := newAppWithCORS(corsConfig{AllowedOrigins: []string{
		"http://localhost:3100",
		"http://localhost:3000",
	}})
	req := httptest.NewRequest("GET", "/ping", nil)
	req.Header.Set("Origin", "http://localhost:3100")
	resp, _ := app.Test(req)
	assert.Equal(t, 200, resp.StatusCode)
	assert.Equal(t, "http://localhost:3100", resp.Header.Get("Access-Control-Allow-Origin"))
}

func TestCORS_ExactMatch_BlocksUnknownOrigin(t *testing.T) {
	app := newAppWithCORS(corsConfig{AllowedOrigins: []string{
		"http://localhost:3100",
	}})
	req := httptest.NewRequest("GET", "/ping", nil)
	req.Header.Set("Origin", "http://evil.example")
	resp, _ := app.Test(req)
	// We don't return 403; we just don't add CORS headers. The browser will
	// block the response on the client side.
	assert.Equal(t, 200, resp.StatusCode)
	assert.Empty(t, resp.Header.Get("Access-Control-Allow-Origin"))
}

func TestCORS_EmptyAllowlist_NoHeaders(t *testing.T) {
	app := newAppWithCORS(corsConfig{AllowedOrigins: nil})
	req := httptest.NewRequest("GET", "/ping", nil)
	req.Header.Set("Origin", "http://anything.example")
	resp, _ := app.Test(req)
	assert.Equal(t, 200, resp.StatusCode)
	assert.Empty(t, resp.Header.Get("Access-Control-Allow-Origin"))
}

func TestCORS_Preflight_AllowedOrigin_Returns204(t *testing.T) {
	app := newAppWithCORS(corsConfig{AllowedOrigins: []string{"http://localhost:3100"}})
	req := httptest.NewRequest("OPTIONS", "/ping", nil)
	req.Header.Set("Origin", "http://localhost:3100")
	resp, _ := app.Test(req)
	assert.Equal(t, 204, resp.StatusCode)
	assert.Equal(t, "http://localhost:3100", resp.Header.Get("Access-Control-Allow-Origin"))
}

func TestCORS_Preflight_DisallowedOrigin_Returns403(t *testing.T) {
	app := newAppWithCORS(corsConfig{AllowedOrigins: []string{"http://localhost:3100"}})
	req := httptest.NewRequest("OPTIONS", "/ping", nil)
	req.Header.Set("Origin", "http://evil.example")
	resp, _ := app.Test(req)
	assert.Equal(t, 403, resp.StatusCode)
}

func TestCORS_NoOrigin_NoHeaders(t *testing.T) {
	app := newAppWithCORS(corsConfig{AllowedOrigins: []string{"*"}})
	req := httptest.NewRequest("GET", "/ping", nil)
	resp, _ := app.Test(req)
	assert.Equal(t, 200, resp.StatusCode)
	// No Origin header on request → no CORS headers on response.
	assert.Empty(t, resp.Header.Get("Access-Control-Allow-Origin"))
}
