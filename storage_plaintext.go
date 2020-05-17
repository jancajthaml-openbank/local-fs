// Copyright (c) 2016-2020, Jan Cajthaml <jan.cajthaml@gmail.com>
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
)

type plaintextFileReader struct {
	source *os.File
}

func (reader *plaintextFileReader) Read(p []byte) (int, error) {
	if reader == nil {
		return 0, fmt.Errorf("cannot read into nil pointer")
	}
	if reader.source == nil {
		return 0, fmt.Errorf("no source to read from")
	}
	n, err := reader.source.Read(p)
	if n == 0 {
		return 0, io.EOF
	}
	if err != nil {
		reader.source.Close()
		return n, err
	}
	return n, nil
}

// Storage is a fascade to access plaintext storage
type PlaintextStorage struct {
	Root       string
	bufferSize int
}

// NewStorage returns new storage over given root
func NewPlaintextStorage(root string) PlaintextStorage {
	if root == "" || os.MkdirAll(filepath.Clean(root), os.ModePerm) != nil {
		panic("unable to assert root storage directory")
	}
	return PlaintextStorage{
		Root:       root,
		bufferSize: 8192,
	}
}

// ListDirectory returns sorted slice of item names in given absolute path
// default sorting is ascending
func (storage PlaintextStorage) ListDirectory(path string, ascending bool) ([]string, error) {
	return listDirectory(storage.Root+"/"+path, storage.bufferSize, ascending)
}

// CountFiles returns number of items in directory
func (storage PlaintextStorage) CountFiles(path string) (int, error) {
	return countFiles(storage.Root+"/"+path, storage.bufferSize)
}

// Exists returns true if path exists
func (storage PlaintextStorage) Exists(path string) (bool, error) {
	return nodeExists(storage.Root + "/" + path)
}

// TouchFile creates files given absolute path if file does not already exist
func (storage PlaintextStorage) TouchFile(path string) error {
	return touch(storage.Root + "/" + path)
}

// DeleteFile removes file given absolute path if that file does exists
func (storage PlaintextStorage) DeleteFile(path string) error {
	return os.Remove(filepath.Clean(storage.Root + "/" + path))
}

// GetFileReader creates file io.Reader
func (storage PlaintextStorage) GetFileReader(path string) (*plaintextFileReader, error) {
	f, err := os.OpenFile(filepath.Clean(storage.Root+"/"+path), os.O_RDONLY, os.ModePerm)
	if err != nil {
		return nil, err
	}

	reader := new(plaintextFileReader)
	reader.source = f

	return reader, nil
}

// ReadFileFully reads whole file given path
func (storage PlaintextStorage) ReadFileFully(path string) ([]byte, error) {
	f, err := os.OpenFile(filepath.Clean(storage.Root+"/"+path), os.O_RDONLY, os.ModePerm)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	fi, err := f.Stat()
	if err != nil {
		return nil, err
	}
	buf := make([]byte, fi.Size())
	_, err = f.Read(buf)
	if err != nil && err != io.EOF {
		return nil, err
	}
	return buf, nil
}

// WriteFileExclusive writes data given path to a file if that file does not
// already exists
func (storage PlaintextStorage) WriteFileExclusive(path string, data []byte) error {
	cleanedPath := filepath.Clean(storage.Root + "/" + path)
	if err := os.MkdirAll(filepath.Dir(cleanedPath), os.ModePerm); err != nil {
		return err
	}
	f, err := os.OpenFile(cleanedPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, os.ModePerm)
	if err != nil {
		return err
	}
	defer f.Close()
	if _, err := f.Write(data); err != nil {
		return err
	}
	return nil
}

// WriteFile rewrite file with data given absolute path to a file if that file
// exist
func (storage PlaintextStorage) WriteFile(path string, data []byte) (err error) {
	cleanedPath := filepath.Clean(storage.Root + "/" + path)
	var f *os.File
	f, err = os.OpenFile(cleanedPath, os.O_WRONLY|os.O_TRUNC, os.ModePerm)
	if err != nil {
		return
	}
	defer f.Close()
	_, err = f.Write(data)
	return
}

// AppendFile appens data given absolute path to a file, creates it if it does
// not exist
func (storage PlaintextStorage) AppendFile(path string, data []byte) (err error) {
	cleanedPath := filepath.Clean(storage.Root + "/" + path)
	err = os.MkdirAll(filepath.Dir(cleanedPath), os.ModePerm)
	if err != nil {
		return err
	}
	var f *os.File
	f, err = os.OpenFile(cleanedPath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, os.ModePerm)
	if err != nil {
		return
	}
	defer f.Close()
	_, err = f.Write(data)
	return
}
