package storage

import (
	"database/sql"
	//_ "github.com/jackc/pgx"
	_ "github.com/jackc/pgx/v5/stdlib"
)

type pgStorage struct {
	db *sql.DB
}

func NewPGXStorage(dsn string) (*pgStorage, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}
	return &pgStorage{db: db}, nil
}

func (pgStorage *pgStorage) Ping() error {
	err := pgStorage.db.Ping()
	if err != nil {
		panic(err)
	}
	return err
}

func (pgStorage *pgStorage) SetURL(user, short, long string) {
	user, short, long = "", "", ""
}

func (pgStorage *pgStorage) GetURL(short string) string {
	return short
}

func (pgStorage *pgStorage) GetURLsByUser(user string) (urls map[string]string) {
	user = ""
	return nil
}
