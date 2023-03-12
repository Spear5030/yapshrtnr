// Package migrate отвечает за миграции.
package migrate

import (
	"database/sql"
	"embed"
	"github.com/pressly/goose/v3"
	"io/fs"
	"log"

	_ "github.com/jackc/pgx/v5/stdlib"
)

//go:embed migrations
var Migrations embed.FS

// Migrate функция миграции
func Migrate(dsn string, path fs.FS) error {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		log.Print(err)
		return err
	}
	defer db.Close()
	goose.SetBaseFS(path)

	return goose.Up(db, "migrations")
}
