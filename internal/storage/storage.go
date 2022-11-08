package storage

/*type storage interface {
	SetURL(short, long string) string
	GetURL(short string) string
}*/

type storage struct {
	URLs map[string]string
}

func New() *storage {
	return &storage{
		make(map[string]string),
	}
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
