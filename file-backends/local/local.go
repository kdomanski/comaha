package local

import (
	log "github.com/Sirupsen/logrus"
	"math/rand"
	"os"
	"path"
)

type localFileBackend struct {
	path string
}

func New(path string) *localFileBackend {
	return &localFileBackend{path: path}
}

var randStringRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func randomString(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = randStringRunes[rand.Intn(len(randStringRunes))]
	}
	return string(b)
}

func (b *localFileBackend) Store(data []byte) (string, error) {
	id := randomString(32)
	filepath := path.Join(b.path, id)
	file, err := os.Create(filepath)
	if err != nil {
		return "", err
	}

	defer file.Close()

	_, err = file.Write(data)
	if err != nil {
		return "", err
	}

	log.Debugf("FILE: saved %v bytes to file '%v'", len(data), filepath)

	return id, nil
}

func (b *localFileBackend) Delete(id string) error {
	filepath := path.Join(b.path, id)
	err := os.Remove(filepath)
	return err
}

func (b *localFileBackend) GetUpdateURL(localURL string) string {
	return localURL + "/file?id="
}
