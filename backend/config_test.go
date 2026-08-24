package main

import "testing"

func TestLoadConfigRequiresDatabase(t *testing.T) {
	t.Setenv("DATABASE_URL", "")
	t.Setenv("NS_MARIADB_HOSTNAME", "")
	t.Setenv("NS_MARIADB_PORT", "")
	t.Setenv("NS_MARIADB_USER", "")
	t.Setenv("NS_MARIADB_PASSWORD", "")
	t.Setenv("NS_MARIADB_DATABASE", "")
	if _, err := loadConfig(); err == nil {
		t.Fatal("expected missing DATABASE_URL to fail")
	}
}

func TestLoadConfigUsesNeoShowcaseMariaDB(t *testing.T) {
	t.Setenv("DATABASE_URL", "")
	t.Setenv("NS_MARIADB_HOSTNAME", "mariadb.internal")
	t.Setenv("NS_MARIADB_PORT", "3306")
	t.Setenv("NS_MARIADB_USER", "shojin")
	t.Setenv("NS_MARIADB_PASSWORD", "p@ss/word")
	t.Setenv("NS_MARIADB_DATABASE", "atcoder_shojin")

	config, err := loadConfig()
	if err != nil {
		t.Fatal(err)
	}
	want := "mariadb://shojin:p%40ss%2Fword@mariadb.internal:3306/atcoder_shojin"
	if config.DatabaseURL != want {
		t.Fatalf("DatabaseURL=%q, want %q", config.DatabaseURL, want)
	}
}
func TestLoadConfigUsesDevelopmentDefaults(t *testing.T) {
	t.Setenv("DATABASE_URL", "mariadb://user:pass@localhost/db")
	t.Setenv("PORT", "")
	t.Setenv("FRONTEND_ORIGIN", "")
	config, err := loadConfig()
	if err != nil {
		t.Fatal(err)
	}
	if config.Port != "8080" || config.FrontendOrigin != "http://localhost:3000" {
		t.Fatalf("config=%#v", config)
	}
}
func TestLoadConfigRejectsInvalidExternalURL(t *testing.T) {
	t.Setenv("DATABASE_URL", "mariadb://user:pass@localhost/db")
	t.Setenv("ATCODER_PROBLEMS_URL", "file:///tmp/data")
	if _, err := loadConfig(); err == nil {
		t.Fatal("expected invalid URL to fail")
	}
}
