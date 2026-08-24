package main

import (
	"github.com/labstack/echo/v4"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestStaticFrontendPreservesAPI404AndSupportsDeepLinks(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "index.html"), []byte("app"), 0o600); err != nil {
		t.Fatal(err)
	}
	server := echo.New()
	registerStaticFrontend(server, dir)
	tests := []struct {
		path   string
		status int
		body   string
	}{
		{path: "/training/session", status: http.StatusOK, body: "app"},
		{path: "/apis/missing", status: http.StatusNotFound},
	}
	for _, test := range tests {
		request := httptest.NewRequest(http.MethodGet, test.path, nil)
		response := httptest.NewRecorder()
		server.ServeHTTP(response, request)
		if response.Code != test.status {
			t.Fatalf("%s status=%d", test.path, response.Code)
		}
		if test.body != "" && response.Body.String() != test.body {
			t.Fatalf("%s body=%q", test.path, response.Body.String())
		}
	}
}
