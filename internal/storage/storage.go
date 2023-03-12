// Package storage реализует слой хранения
package storage

import (
	"bufio"
	"bytes"
	"context"
	"encoding/gob"
	"fmt"
	"io"
	"os"

	"github.com/Spear5030/yapshrtnr/internal/domain"
)

type link struct {
	User  string
	Short string
	Long  string
}

type storage struct {
	URLs    map[string]string
	Users   map[string][]string
	Deleted map[string]string
}

type fileStorage struct {
	filename string
	storage
}

// NewMemoryStorage возвращает хранилище в памяти
func NewMemoryStorage() *storage {
	return &storage{
		make(map[string]string),
		make(map[string][]string),
		make(map[string]string),
	}
}

// NewFileStorage возвращает файловое хранилище
func NewFileStorage(filename string) (*fileStorage, error) {
	file, err := os.OpenFile(filename, os.O_RDONLY|os.O_APPEND|os.O_CREATE, 0777)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	rd := bufio.NewReader(file)
	var buffer bytes.Buffer
	var url link
	urls := make(map[string]string)
	users := make(map[string]string)
	for {
		b, err := rd.ReadBytes(13) // "\n"
		if err != nil {
			if err == io.EOF {
				break
			} else {
				fmt.Println(err)
				break
			}
		}
		buffer.Write(b)
		gob.NewDecoder(&buffer).Decode(&url)
		urls[url.Short] = url.Long
		users[url.User] = url.User

		buffer.Reset()
	}
	storage := NewMemoryStorage()
	return &fileStorage{
		filename: filename,
		storage:  *storage,
	}, nil
}

// SetURL записывает связь short и long в map памяти.
func (mStorage *storage) SetURL(ctx context.Context, user, short, long string) error {
	mStorage.URLs[short] = long
	mStorage.Users[user] = append(mStorage.Users[user], short)
	return nil
}

// GetURL возвращает полный URL из хранилища памяти.
func (mStorage *storage) GetURL(ctx context.Context, short string) (string, bool) {
	if _, ok := mStorage.Deleted[short]; ok {
		return "", true
	}

	if v, ok := mStorage.URLs[short]; ok {
		return v, false
	}
	return "", false
}

// GetURLsByUser возвращает список URL созданных определенным пользователем из хранилища памяти.
func (mStorage *storage) GetURLsByUser(ctx context.Context, user string) (urls map[string]string) {
	urls = make(map[string]string)
	if shorts, ok := mStorage.Users[user]; ok {
		for _, short := range shorts {
			urls[short] = mStorage.URLs[short]
		}
	}
	return
}

// SetURL записывает связь short и long в map памяти. После сбрасывает изменения в файл
func (fStorage *fileStorage) SetURL(ctx context.Context, user, short, long string) error {
	fStorage.URLs[short] = long
	fStorage.Users[user] = append(fStorage.Users[user], short)
	file, err := os.OpenFile(fStorage.filename, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0777)
	if err != nil {
		panic(err)
	}
	defer file.Close()
	var buffer bytes.Buffer
	linkToEncode := link{
		User:  user,
		Short: short,
		Long:  long,
	}
	if err := gob.NewEncoder(&buffer).Encode(linkToEncode); err != nil {
		panic(err)
	}
	file.Write(append(buffer.Bytes(), 13))
	return nil
}

// GetURLsByUser возвращает список URL созданных определенным пользователем из хранилища.
func (fStorage *fileStorage) GetURLsByUser(ctx context.Context, user string) (urls map[string]string) {
	urls = make(map[string]string)
	if shorts, ok := fStorage.Users[user]; ok {

		for _, short := range shorts {
			urls[short] = fStorage.URLs[short]
		}
	}
	return
}

// GetURL возвращает полный URL.
func (fStorage *fileStorage) GetURL(ctx context.Context, short string) (string, bool) {
	if v, ok := fStorage.URLs[short]; ok {
		return v, false
	}
	return "", false
}

// Ping не имплементировано для данного хранилища
func (mStorage *storage) Ping() error {
	return nil
}

// SetBatchURLs пакетное сохранение ссылок в памяти
func (mStorage *storage) SetBatchURLs(ctx context.Context, urls []domain.URL) error {
	for _, u := range urls {
		mStorage.URLs[u.Short] = u.Long
		mStorage.Users[u.User] = append(mStorage.Users[u.User], u.Short)
	}
	return nil
}

// SetBatchURLs не имплементировано для данного хранилища
func (fStorage *fileStorage) SetBatchURLs(ctx context.Context, urls []domain.URL) error {
	return nil
}

// DeleteURLs пакетное удаление ссылок в памяти
func (mStorage *storage) DeleteURLs(ctx context.Context, user string, shorts []string) {
	for _, short := range shorts {
		for _, s := range mStorage.Users[user] {
			if s == short {
				mStorage.Deleted[short] = user
			}
		}
	}
}

// DeleteURLs не имплементировано для данного хранилища
func (fStorage *fileStorage) DeleteURLs(ctx context.Context, user string, shorts []string) {
}

// Shutdown не имплементировано для данного хранилища
func (mStorage *storage) Shutdown() error {
	return nil
}

// Shutdown не имплементировано для данного хранилища
func (fStorage *fileStorage) Shutdown() error {
	return nil
}
