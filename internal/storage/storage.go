package storage

import (
	"bufio"
	"bytes"
	"context"
	"encoding/gob"
	"fmt"
	"github.com/Spear5030/yapshrtnr/internal/domain"
	"io"
	"os"
)

type link struct {
	User  string
	Short string
	Long  string
}

type storage struct {
	URLs  map[string]string
	Users map[string][]string
}

type fileStorage struct {
	filename string
	storage
}

func NewMemoryStorage() *storage {
	return &storage{
		make(map[string]string),
		make(map[string][]string),
	}
}

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

func (mStorage *storage) SetURL(user, short, long string) {
	mStorage.URLs[short] = long
	mStorage.Users[user] = append(mStorage.Users[user], short)
}

func (mStorage *storage) GetURL(short string) string {
	if v, ok := mStorage.URLs[short]; ok {
		return v
	}
	return ""
}

func (mStorage *storage) GetURLsByUser(user string) (urls map[string]string) {
	urls = make(map[string]string)
	if shorts, ok := mStorage.Users[user]; ok {
		for _, short := range shorts {
			urls[short] = mStorage.URLs[short]
		}
		//url := link{}
		//url.Short = short
		//url.Long = mStorage.URLs[short]
		//result = append(result, url)
	}
	return
}

func (fStorage *fileStorage) SetURL(user, short, long string) {
	fStorage.URLs[short] = long
	fStorage.Users[user] = append(fStorage.Users[user], short)
	file, err := os.OpenFile(fStorage.filename, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0777)
	if err != nil {
		panic(err)
	}
	defer file.Close()
	var buffer bytes.Buffer
	link := link{
		User:  user,
		Short: short,
		Long:  long,
	}
	if err := gob.NewEncoder(&buffer).Encode(link); err != nil {
		panic(err)
	}
	file.Write(append(buffer.Bytes(), 13))
}

func (fStorage *fileStorage) GetURLsByUser(user string) (urls map[string]string) {
	urls = make(map[string]string)
	if shorts, ok := fStorage.Users[user]; ok {

		for _, short := range shorts {
			urls[short] = fStorage.URLs[short]
			//url := link{}
			//url.Short = short
			//url.Long = fStorage.URLs[short]
			//result = append(result, url)
		}
	}
	return
}

func (fStorage *fileStorage) GetURL(short string) string {
	if v, ok := fStorage.URLs[short]; ok {
		return v
	}
	return ""
}

func (fStorage *fileStorage) Ping() error {
	return nil
}

func (mStorage *storage) Ping() error {
	return nil
}

func (mStorage *storage) SetBatchURLs(ctx context.Context, urls []domain.URL) error {
	return nil
}

func (fStorage *fileStorage) SetBatchURLs(ctx context.Context, urls []domain.URL) error {
	return nil
}