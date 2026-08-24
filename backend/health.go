package main

import (
	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v4"
	"net/http"
)

func healthHandler(db *sqlx.DB) echo.HandlerFunc {
	return func(context echo.Context) error {
		if err := db.PingContext(context.Request().Context()); err != nil {
			return context.JSON(http.StatusServiceUnavailable, map[string]string{"status": "unhealthy", "database": "disconnected"})
		}
		return context.JSON(http.StatusOK, map[string]string{"status": "ok", "database": "connected"})
	}
}
