package main

import (
	"github.com/go-sql-driver/mysql"
	"testing"
	"time"
)

func TestNormalizeDatabaseURL(t *testing.T) {
	dsn, err := normalizeDatabaseURL("mariadb://shojin:p%40ss@db.example:3307/training?tls=true")
	if err != nil {
		t.Fatal(err)
	}
	config, err := mysql.ParseDSN(dsn)
	if err != nil {
		t.Fatal(err)
	}
	if config.User != "shojin" || config.Passwd != "p@ss" || config.Addr != "db.example:3307" || config.DBName != "training" || !config.ParseTime || config.Loc != time.UTC || config.MultiStatements {
		t.Fatalf("config=%#v", config)
	}
}
func TestNormalizeDatabaseURLRejectsOtherScheme(t *testing.T) {
	if _, err := normalizeDatabaseURL("postgres://user:pass@localhost/db"); err == nil {
		t.Fatal("expected invalid scheme")
	}
}
