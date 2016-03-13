package main

type payload struct {
	ID      string
	Version string
	SHA1    string
	SHA256  string
	Size    int64
}

type fileBackend interface {
	//StorageURL() string
	Store(data []byte) (string, error)
	Delete(id string) error
	GetUpdateURL(localURL string) string
}
