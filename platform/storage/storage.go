package storage

import (
	"errors"
	"os"
	"path"
	"strings"
)

type Storage struct {
	dir string
}

func New(name string) (*Storage, error) {
	name = strings.ToLower(name)

	err := os.MkdirAll(name, os.ModeDir)
	if err != nil && !errors.Is(err, os.ErrExist) {
		return nil, err
	}

	return &Storage{
		dir: name,
	}, nil
}

func (s *Storage) MakePath(p ...string) string {
	return path.Join(append([]string{s.dir}, p...)...)
}
