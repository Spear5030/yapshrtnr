package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/Spear5030/yapshrtnr/internal/domain"

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
}

// URL структура с CorrelationID для связи списков сокращенных и полных URL при пакетной обработке
type URL struct {
	domain.URL
	CorrelationID string `db:"correlation_id"`
}

// ResultBatch результирующая структура с CorrelationID для связи с коротким URL
type ResultBatch struct {
	long          string
	correlationID string
}

// DuplicationError ошибка при конфликте уже сохраненного URL. содержит в себе сокращенный ранее идентификатор
type DuplicationError struct {
	Duplication string
	Err         error
}

// Error для интерфейса Error
func (derr *DuplicationError) Error() string {
	return fmt.Sprintf("%v", derr.Err)
}

// Pinger интерфейс для БД
type Pinger interface {
	Ping() error
}

// NewDuplicationError возвращает ошибку DuplicationError
func NewDuplicationError(dup string, err error) error {
	return &DuplicationError{
		Duplication: dup,
		Err:         err,
	}
}

// NewPGXStorage возвращает хранилище PostrgeSQL. Запускает горутину-воркер для удаления
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
		chanForDel: make(chan urlsForDelete), //канал, в который отправляются задачи(пользователь, слайс URL)
		deleteWork: make(chan bool),          //канал по которому стартуем саму операцию удаления()
	}
	go pgS.WorkWithDeleteBatch(context.Background()) // функция с циклом for-select - ожидает значения в каналах chanForDel и deleteWork
	return &pgS, nil
}

// Ping реализует интерфейс Pinger
func (pgStorage *pgStorage) Ping() error {
	err := pgStorage.db.Ping()
	if err != nil {
		panic(err)
	}
	return err
}

// Shutdown форсит пакетное удаление
func (pgStorage *pgStorage) Shutdown() error {
	log.Println("Shutdown Postgre storage")
	pgStorage.deleteWork <- true
	return pgStorage.db.Close()
}

// DeleteURLs отправляет в канал список удаляемых URL. Через 500мс через канал deleteWork выполняет отложенный запуск
func (pgStorage *pgStorage) DeleteURLs(ctx context.Context, user string, shorts []string) {
	time.AfterFunc(500*time.Millisecond, func() {
		pgStorage.deleteWork <- true
	})
	chunk := urlsForDelete{user, shorts}
	pgStorage.chanForDel <- chunk
}

// WorkWithDeleteBatch Сбор URL из канала chanForDel. По каналу deleteWork удаление в хранилище
func (pgStorage *pgStorage) WorkWithDeleteBatch(ctx context.Context) {
	urlsByUser := make(map[string][]string)
	for {
		select {
		case <-ctx.Done():
			log.Println("Context done")
		case x := <-pgStorage.chanForDel:
			urlsByUser[x.user] = append(urlsByUser[x.user], x.shorts...)
		case <-pgStorage.deleteWork:
			for user, shorts := range urlsByUser {
				err := pgStorage.DeleteBatchURLs(user, shorts)
				if err != nil {
					log.Println(err)
				}
			}
			urlsByUser = make(map[string][]string)
		}
	}
}

// DeleteBatchURLs Пакетное удаление массива URL
func (pgStorage *pgStorage) DeleteBatchURLs(user string, shorts []string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	query := `UPDATE urls SET deleted = true WHERE 
                                   userID = $1 AND short = any ($2);`
	_, err := pgStorage.db.ExecContext(ctx, query, user, shorts)
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

// SetURL запись URL в PostgeSQL
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

// GetURL Получение оригинального URL по короткой записи. Возвращает вторым аргументом bool - удален ли URL
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

// GetURLsByUser не имплементировано на БД
func (pgStorage *pgStorage) GetURLsByUser(ctx context.Context, user string) (urls map[string]string) {
	return nil
}

// SetBatchURLs Пакетная запись URL в PostgreSQL
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

// GetUsersCount возвращает количество пользователей
func (pgStorage *pgStorage) GetUsersCount(ctx context.Context) (int, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var count int

	sql := `SELECT COUNT(userid) from urls;`
	err := pgStorage.db.QueryRowContext(ctx, sql).Scan(&count)
	if err != nil {
		return -1, err
	}
	return count, nil
}

// GetUrlsCount возвращает количество ссылок
func (pgStorage *pgStorage) GetUrlsCount(ctx context.Context) (int, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var count int

	sql := `SELECT COUNT(long) from urls;`
	err := pgStorage.db.QueryRowContext(ctx, sql).Scan(&count)
	if err != nil {
		return -1, err
	}
	return count, nil
}
