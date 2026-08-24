package main

import (
	"github.com/labstack/echo/v4"
	"os"
	"path/filepath"
	"strings"
)

func registerStaticFrontend(server *echo.Echo, dir string) {
	if dir == "" {
		return
	}
	root, err := filepath.Abs(dir)
	if err != nil {
		return
	}
	server.GET("/*", func(context echo.Context) error {
		requestPath := strings.TrimPrefix(filepath.Clean("/"+context.Param("*")), string(filepath.Separator))
		if requestPath == "apis" || strings.HasPrefix(requestPath, "apis"+string(filepath.Separator)) {
			return echo.ErrNotFound
		}
		candidate := filepath.Join(root, requestPath)
		relative, err := filepath.Rel(root, candidate)
		if err != nil || strings.HasPrefix(relative, "..") || filepath.IsAbs(relative) {
			return echo.ErrNotFound
		}
		if info, err := os.Stat(candidate); err == nil && !info.IsDir() {
			return context.File(candidate)
		}
		if filepath.Ext(requestPath) != "" || strings.HasPrefix(requestPath, "static"+string(filepath.Separator)) {
			return echo.ErrNotFound
		}
		return context.File(filepath.Join(root, "index.html"))
	})
}
