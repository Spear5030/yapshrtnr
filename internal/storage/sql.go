package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/Spear5030/yapshrtnr/internal/domain"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"log"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type pgStorage struct {
	db         *sql.DB
	chanForDel chan urlsForDelete
	deleteWork chan bool
}

type urlsForDelete struct {
	user   string
	shorts []string
	ctx    context.Context // не нравится, но не сообразил как реализовать передачу контекста
}

type URL struct {
	domain.URL
	CorrelationID string `db:"correlation_id"`
}
type ResultBatch struct {
	long          string
	correlationID string
}

type DuplicationError struct {
	Duplication string
	Err         error
}

func (derr *DuplicationError) Error() string {
	return fmt.Sprintf("%v", derr.Err)
}

type Pinger interface {
	Ping() error
}

func NewDuplicationError(dup string, err error) error {
	return &DuplicationError{
		Duplication: dup,
		Err:         err,
	}
}

func NewPGXStorage(dsn string) (*pgStorage, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	err = db.Ping()
	if err != nil {
		log.Println(err)
		return nil, err
	}
	pgS := pgStorage{
		db:         db,
		chanForDel: make(chan urlsForDelete),
		deleteWork: make(chan bool),
	}
	go pgS.WorkWithDeleteBatch()
	return &pgS, nil
}

func (pgStorage *pgStorage) Ping() error {
	err := pgStorage.db.Ping()
	if err != nil {
		panic(err)
	}
	return err
}

func (pgStorage *pgStorage) DeleteURLs(ctx context.Context, user string, shorts []string) {
	time.AfterFunc(10*time.Millisecond, func() {
		pgStorage.deleteWork <- true
	})
	chunk := urlsForDelete{user, shorts, ctx}
	pgStorage.chanForDel <- chunk
}

func (pgStorage *pgStorage) WorkWithDeleteBatch() {
	urlsByUser := make(map[string][]string)
	ctxByUser := make(map[string]context.Context)
	for {
		select {
		case x := <-pgStorage.chanForDel:
			log.Println(x.user, " will delete ", x.shorts)
			urlsByUser[x.user] = append(urlsByUser[x.user], x.shorts...)
			ctxByUser[x.user] = x.ctx
		case <-pgStorage.deleteWork:
			log.Println(urlsByUser)
			if len(urlsByUser) == 0 {
				return
			}
			for user, shorts := range urlsByUser {
				log.Println(user, " deleted ", shorts)
				err := pgStorage.DeleteBatchURLs(ctxByUser[user], user, shorts)
				if err != nil {
					log.Println(err)
				}
			}
			urlsByUser = make(map[string][]string)
		}
	}
}

func (pgStorage *pgStorage) DeleteBatchURLs(ctx context.Context, user string, shorts []string) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	query := `UPDATE urls SET deleted = true WHERE userID = $1 AND short = ANY $2);`
	_, err := pgStorage.db.ExecContext(ctx, query, user, shorts)
	if err != nil {
		return err
	}
	return nil
}

func (pgStorage *pgStorage) SetURL(ctx context.Context, user, short, long string) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
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
			return NewDuplicationError(short, err)
		} else {
			log.Print(err.Error())
			return err
		}
	}
	return nil
}

func (pgStorage *pgStorage) GetURL(ctx context.Context, short string) (string, bool) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	sql := `SELECT long, deleted FROM urls WHERE short=$1;`
	row := pgStorage.db.QueryRowContext(ctx, sql, short)
	var long string
	var deleted bool

	err := row.Scan(&long, &deleted)
	if err != nil {
		log.Println(err)
	}
	return long, deleted
}

func (pgStorage *pgStorage) GetURLsByUser(ctx context.Context, user string) (urls map[string]string) {
	return nil
}

func (pgStorage *pgStorage) SetBatchURLs(ctx context.Context, urls []domain.URL) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

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
