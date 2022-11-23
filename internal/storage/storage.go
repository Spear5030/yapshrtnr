package storage

import (
	"bufio"
	"bytes"
	"encoding/gob"
	"fmt"
	"io"
	"os"
)

type link struct {
	Short string
	Long  string
}

type storage struct {
	URLs map[string]string
}

type fileStorage struct {
	filename string
	storage
}

func NewMemoryStorage() *storage {
	return &storage{
		make(map[string]string),
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
		buffer.Reset()
	}
	storage := NewMemoryStorage()
	return &fileStorage{
		filename: filename,
		storage:  *storage,
	}, nil
}

func (mStorage *storage) SetURL(short, long string) {
	mStorage.URLs[short] = long
}

func (mStorage *storage) GetURL(short string) string {
	if v, ok := mStorage.URLs[short]; ok {
		return v
	}
	return ""
}

func (fStorage *fileStorage) SetURL(short, long string) {
	fStorage.URLs[short] = long
	file, err := os.OpenFile(fStorage.filename, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0777)
	if err != nil {
		panic(err)
	}
	defer file.Close()
	var buffer bytes.Buffer
	link := link{
		Short: short,
		Long:  long,
	}
	if err := gob.NewEncoder(&buffer).Encode(link); err != nil {
		panic(err)
	}
	file.Write(append(buffer.Bytes(), 13))
}

func (fStorage *fileStorage) GetURL(short string) string {
	if v, ok := fStorage.URLs[short]; ok {
		return v
	}
	return ""
}
