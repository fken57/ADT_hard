package main

import "testing"

func TestLoadConfigRequiresDatabase(t *testing.T) {
	t.Setenv("DATABASE_URL", "")
	if _, err := loadConfig(); err == nil {
		t.Fatal("expected missing DATABASE_URL to fail")
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
