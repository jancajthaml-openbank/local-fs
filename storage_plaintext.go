// Copyright (c) 2017-2021, Jan Cajthaml <jan.cajthaml@gmail.com>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package storage

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"syscall"
	"time"
)

// PlaintextStorage is a fascade to access plaintext storage
type PlaintextStorage struct {
	Storage
	root       string
	bufferSize int
}

// NewPlaintextStorage returns new storage over given root
func NewPlaintextStorage(root string) (Storage, error) {
	if root == "" {
		return NilStorage{}, fmt.Errorf("invalid root directory")
	}
	if os.MkdirAll(filepath.Clean(root), os.ModePerm) != nil {
		return NilStorage{}, fmt.Errorf("unable to assert root storage directory")
	}
	return PlaintextStorage{
		root:       root,
		bufferSize: 8192,
	}, nil
}

// Chmod sets chmod flag on given file
func (storage PlaintextStorage) Chmod(path string, mod os.FileMode) error {
	return chmod(storage.root+"/"+path, mod)
}

// ListDirectory returns sorted slice of item names in given absolute path
// default sorting is ascending
func (storage PlaintextStorage) ListDirectory(path string, ascending bool) ([]string, error) {
	return listDirectory(storage.root+"/"+path, storage.bufferSize, ascending)
}

// CountFiles returns number of items in directory
func (storage PlaintextStorage) CountFiles(path string) (int, error) {
	return countFiles(storage.root+"/"+path, storage.bufferSize)
}

// Exists returns true if path exists
func (storage PlaintextStorage) Exists(path string) (bool, error) {
	return nodeExists(storage.root + "/" + path)
}

// LastModification returns time of last modification
func (storage PlaintextStorage) LastModification(path string) (time.Time, error) {
	return modTime(storage.root + "/" + path)
}

// TouchFile creates files given absolute path if file does not already exist
func (storage PlaintextStorage) TouchFile(path string) error {
	return touch(storage.root + "/" + path)
}

// Delete removes given absolute path if that file does exists
func (storage PlaintextStorage) Delete(path string) error {
	return os.RemoveAll(filepath.Clean(storage.root + "/" + path))
}

// ReadFileFully reads whole file given path
func (storage PlaintextStorage) ReadFileFully(path string) ([]byte, error) {
	filename := filepath.Clean(storage.root + "/" + path)
	fd, err := syscall.Open(filename, syscall.O_RDONLY|syscall.O_NONBLOCK, 0600)
	if err != nil {
		return nil, err
	}
	defer syscall.Close(fd)
	if err = syscall.Flock(fd, syscall.LOCK_EX); err != nil {
		return nil, err
	}
	defer syscall.Flock(fd, syscall.LOCK_UN)
	var fs syscall.Stat_t
	if err = syscall.Fstat(fd, &fs); err != nil {
		return nil, err
	}
	buf := make([]byte, fs.Size)
	if _, err = syscall.Read(fd, buf); err != nil && err != io.EOF {
		return nil, err
	}
	return buf, nil
}

// WriteFileExclusive writes data given path to a file if that file does not
// already exists
func (storage PlaintextStorage) WriteFileExclusive(path string, data []byte) error {
	filename := filepath.Clean(storage.root + "/" + path)
	if err := os.MkdirAll(filepath.Dir(filename), os.ModePerm); err != nil {
		return err
	}
	fd, err := syscall.Open(filename, syscall.O_CREAT|syscall.O_WRONLY|syscall.O_EXCL|syscall.O_NONBLOCK, 0600)
	if err != nil {
		return err
	}
	defer func() {
		syscall.Close(fd)
		syscall.Fsync(fd)
	}()
	if err = syscall.Flock(fd, syscall.LOCK_EX); err != nil {
		return err
	}
	defer syscall.Flock(fd, syscall.LOCK_UN)
	if _, err := syscall.Write(fd, data); err != nil {
		return err
	}
	return nil
}

// WriteFile writes data given absolute path to a file, creates it if it does
// not exist
func (storage PlaintextStorage) WriteFile(path string, data []byte) error {
	filename := filepath.Clean(storage.root + "/" + path)
	if err := os.MkdirAll(filepath.Dir(filename), os.ModePerm); err != nil {
		return err
	}
	fd, err := syscall.Open(filename, syscall.O_CREAT|syscall.O_WRONLY|syscall.O_TRUNC|syscall.O_NONBLOCK, 0600)
	if err != nil {
		return err
	}
	defer func() {
		syscall.Close(fd)
		syscall.Fsync(fd)
	}()
	if err = syscall.Flock(fd, syscall.LOCK_EX); err != nil {
		return err
	}
	defer syscall.Flock(fd, syscall.LOCK_UN)
	if _, err := syscall.Write(fd, data); err != nil {
		return err
	}
	return nil
}

// AppendFile appens data given absolute path to a file, creates it if it does
// not exist
func (storage PlaintextStorage) AppendFile(path string, data []byte) error {
	filename := filepath.Clean(storage.root + "/" + path)
	if err := os.MkdirAll(filepath.Dir(filename), os.ModePerm); err != nil {
		return err
	}
	fd, err := syscall.Open(filename, syscall.O_CREAT|syscall.O_WRONLY|syscall.O_APPEND|syscall.O_NONBLOCK, 0600)
	if err != nil {
		return err
	}
	defer func() {
		syscall.Close(fd)
		syscall.Fsync(fd)
	}()
	if err = syscall.Flock(fd, syscall.LOCK_EX); err != nil {
		return err
	}
	defer syscall.Flock(fd, syscall.LOCK_UN)
	if _, err := syscall.Write(fd, data); err != nil {
		return err
	}
	return nil
}
