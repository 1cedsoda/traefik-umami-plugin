package traefik_umami_plugin_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	traefik_umami_plugin "github.com/1cedsoda/traefik-umami-plugin"
)

// func TestDemo(t *testing.T) {
// 	cfg := traefik_umami_plugin.CreateConfig()
// 	cfg.Headers["X-Host"] = "[[.Host]]"
// 	cfg.Headers["X-Method"] = "[[.Method]]"
// 	cfg.Headers["X-URL"] = "[[.URL]]"
// 	cfg.Headers["X-URL"] = "[[.URL]]"
// 	cfg.Headers["X-Demo"] = "test"

// 	ctx := context.Background()
// 	next := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {})

// 	handler, err := traefik_umami_plugin.New(ctx, next, cfg, "demo-plugin")
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	recorder := httptest.NewRecorder()

// 	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://localhost", nil)
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	handler.ServeHTTP(recorder, req)

// 	assertHeader(t, req, "X-Host", "localhost")
// 	assertHeader(t, req, "X-URL", "http://localhost")
// 	assertHeader(t, req, "X-Method", "GET")
// 	assertHeader(t, req, "X-Demo", "test")
// }

// func assertHeader(t *testing.T, req *http.Request, key, expected string) {
// 	t.Helper()

// 	if req.Header.Get(key) != expected {
// 		t.Errorf("invalid header value: %s", req.Header.Get(key))
// 	}
// }

func setup(cfg *traefik_umami_plugin.Config) (ctx context.Context, handler http.Handler, recorder *httptest.ResponseRecorder) {

	ctx = context.Background()
	next := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {})

	handler, err := traefik_umami_plugin.New(ctx, next, cfg, "traefik-umami-plugin")
	if err != nil {
		panic(err)
	}

	recorder = httptest.NewRecorder()

	return ctx, handler, recorder
}

func createConfig() *traefik_umami_plugin.Config {
	cfg := traefik_umami_plugin.CreateConfig()
	// cfg.UmamiConfig.UmamiHost = "http://localhost:3000"
	// cfg.UmamiConfig.WebsiteId = "myWebsiteId"
	cfg.ForwardPath = "_umami"
	return cfg
}

func TestDoForward(t *testing.T) {
	cfg := createConfig()
	ctx, handler, recorder := setup(cfg)

	// Make Request that should not be forwarded
	req1, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://localhost:443/_umami", nil)
	if err != nil {
		t.Fatal(err)
	}
	handler.ServeHTTP(recorder, req1)

	fmt.Println(recorder.Body.String())
}
