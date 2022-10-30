package storage

/*type storage interface {
	SetURL(short, long string) string
	GetURL(short string) string
}*/

type Storage struct {
	URLs map[string]string
}

func NewStorage() *Storage {
	return &Storage{
		make(map[string]string),
	}
}

func (mStorage *Storage) SetURL(short, long string) {
	mStorage.URLs[short] = long
}

func (mStorage *Storage) GetURL(short string) string {
	if v, ok := mStorage.URLs[short]; ok {
		return v
	}
	return ""
}
