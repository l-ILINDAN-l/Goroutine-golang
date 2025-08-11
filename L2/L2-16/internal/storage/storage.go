package storage

import (
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

type Storage struct {
	outputDir string
}

func NewStorage(outputDir string) *Storage {
	return &Storage{outputDir: outputDir}
}

func (s *Storage) URLToPath(u *url.URL) string {
	path := s.outputDir
	path = filepath.Join(path, u.Host)

	if u.Path == "" || u.Path == "/" {
		return filepath.Join(path, "index.html")
	}

	path = filepath.Join(path, u.Path)

	if strings.HasSuffix(u.Path, "/") {
		path = filepath.Join(path, "index.html")
	}

	return path
}

func (s *Storage) Save(path string, reader io.Reader) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("error creationd directory: %v", err)
	}

	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("error creating file: %v", err)
	}
	defer file.Close()

	_, err = io.Copy(file, reader)
	if err != nil {
		return fmt.Errorf("error saving file: %v", err)
	}

	return nil
}
