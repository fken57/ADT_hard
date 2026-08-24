package main

import (
	"fmt"
	"github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"net/url"
	"strings"
	"time"
)

func openDatabase(raw string) (*sqlx.DB, error) {
	dsn, err := normalizeDatabaseURL(raw)
	if err != nil {
		return nil, err
	}
	db, err := sqlx.Connect("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("connect to MariaDB: %w", err)
	}
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(30 * time.Minute)
	return db, nil
}
func normalizeDatabaseURL(raw string) (string, error) {
	if !strings.Contains(raw, "://") {
		parsed, err := mysql.ParseDSN(raw)
		if err != nil {
			return "", err
		}
		return normalizeMySQL(*parsed).FormatDSN(), nil
	}
	parsed, err := url.Parse(raw)
	if err != nil {
		return "", err
	}
	if parsed.Scheme != "mysql" && parsed.Scheme != "mariadb" {
		return "", fmt.Errorf("DATABASE_URL scheme must be mysql or mariadb")
	}
	password, _ := parsed.User.Password()
	name := strings.TrimPrefix(parsed.Path, "/")
	port := parsed.Port()
	if port == "" {
		port = "3306"
	}
	if parsed.User.Username() == "" || parsed.Hostname() == "" || name == "" {
		return "", fmt.Errorf("DATABASE_URL must include user, host, and database")
	}
	return normalizeMySQL(mysql.Config{User: parsed.User.Username(), Passwd: password, Net: "tcp", Addr: parsed.Hostname() + ":" + port, DBName: name, TLSConfig: parsed.Query().Get("tls")}).FormatDSN(), nil
}
func normalizeMySQL(config mysql.Config) *mysql.Config {
	config.ParseTime = true
	config.Loc = time.UTC
	config.MultiStatements = false
	config.AllowNativePasswords = true
	config.Collation = "utf8mb4_unicode_ci"
	config.Timeout = 5 * time.Second
	config.ReadTimeout = 15 * time.Second
	config.WriteTimeout = 15 * time.Second
	if config.Params == nil {
		config.Params = map[string]string{}
	}
	config.Params["charset"] = "utf8mb4"
	return &config
}
