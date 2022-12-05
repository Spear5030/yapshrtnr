package storage

import (
	"context"
	"database/sql"
	"errors"
	"github.com/Spear5030/yapshrtnr/internal/domain"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"log"
	"time"

	//_ "github.com/jackc/pgx"
	_ "github.com/jackc/pgx/v5/stdlib"
)

type pgStorage struct {
	db *sql.DB
}

type URL struct {
	domain.URL
	CorrelationID string `db:"correlation_id"`
}
type ResultBatch struct {
	long          string
	correlationID string
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

func (pgStorage *pgStorage) SetURL(user, short, long string) (result string) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `INSERT INTO urls(short, long, userID) 
          			VALUES($1, $2, $3);`
	_, err := pgStorage.db.ExecContext(ctx, query, short, long, user)
	var pgErr *pgconn.PgError
	if err != nil {
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			query := `SELECT short FROM urls WHERE long=$1;`
			row := pgStorage.db.QueryRowContext(ctx, query, long)
			row.Scan(&short)
			return short
		} else {
			log.Print(err.Error())
		}
	}
	return ""
}

func (pgStorage *pgStorage) GetURL(short string) string {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	sql := `SELECT long FROM urls WHERE short=$1;`
	row := pgStorage.db.QueryRowContext(ctx, sql, short)
	var long string
	row.Scan(&long)
	return long
}

func (pgStorage *pgStorage) GetURLsByUser(user string) (urls map[string]string) {
	return nil
}

func (pgStorage *pgStorage) SetBatchURLs(ctx context.Context, urls []domain.URL) error {
	tx, err := pgStorage.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	stmt, err := tx.PrepareContext(ctx, "INSERT INTO urls(short, long, userID) VALUES($1,$2,$3);")
	if err != nil {
		return err
	}
	defer stmt.Close()
	for _, url := range urls {
		if _, err = stmt.ExecContext(ctx, url.Short, url.Long, url.User); err != nil {
			return err
		}
	}
	return tx.Commit()
}
