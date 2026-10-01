package migrations

import (
	"embed"
	"fmt"

	"github.com/jmoiron/sqlx"
)

//go:embed 001_init.sql
var fs embed.FS

// Apply выполняет все миграции при старте приложения.
// SQL встроен в бинарник через go:embed, поэтому путь не зависит от рабочей директории.
func Apply(db *sqlx.DB) error {
	data, err := fs.ReadFile("001_init.sql")
	if err != nil {
		return fmt.Errorf("read embedded migration: %w", err)
	}
	if _, err := db.Exec(string(data)); err != nil {
		return fmt.Errorf("execute migration 001_init: %w", err)
	}
	return nil
}