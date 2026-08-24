package main

import (
	"context"
	"embed"
	"fmt"
	"github.com/jmoiron/sqlx"
	"io/fs"
	"path"
	"sort"
	"strings"
)

//go:embed migrations/*.sql
var migrationFiles embed.FS

func runMigrations(ctx context.Context, db *sqlx.DB) error {
	entries, err := fs.ReadDir(migrationFiles, "migrations")
	if err != nil {
		return err
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].Name() < entries[j].Name() })
	if _, err = db.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS schema_migrations(version VARCHAR(255) CHARACTER SET ascii COLLATE ascii_bin PRIMARY KEY,applied_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6))`); err != nil {
		return err
	}
	for _, entry := range entries {
		if entry.IsDir() || path.Ext(entry.Name()) != ".sql" {
			continue
		}
		version := strings.TrimSuffix(entry.Name(), ".sql")
		var count int
		if err = db.GetContext(ctx, &count, "SELECT COUNT(*) FROM schema_migrations WHERE version=?", version); err != nil {
			return err
		}
		if count > 0 {
			continue
		}
		content, readErr := migrationFiles.ReadFile(path.Join("migrations", entry.Name()))
		if readErr != nil {
			return readErr
		}
		tx, beginErr := db.BeginTxx(ctx, nil)
		if beginErr != nil {
			return beginErr
		}
		for _, statement := range strings.Split(string(content), ";") {
			statement = strings.TrimSpace(statement)
			if statement == "" {
				continue
			}
			if _, err = tx.ExecContext(ctx, statement); err != nil {
				break
			}
		}
		if err == nil {
			_, err = tx.ExecContext(ctx, "INSERT INTO schema_migrations(version) VALUES(?)", version)
		}
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("migration %s: %w", version, err)
		}
		if err = tx.Commit(); err != nil {
			return err
		}
	}
	return nil
}
