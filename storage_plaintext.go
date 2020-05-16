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
	"bytes"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"sort"
	"syscall"
	"unsafe"
)

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
func (storage PlaintextStorage) ListDirectory(path string, ascending bool) (result []string, err error) {
	var (
		n  int
		dh *os.File
		de *syscall.Dirent
	)

	dh, err = os.Open(filepath.Clean(storage.Root + "/" + path))
	if err != nil {
		return
	}

	fd := int(dh.Fd())
	result = make([]string, 0)

	scratchBuffer := make([]byte, storage.bufferSize)

	for {
		n, err = syscall.ReadDirent(fd, scratchBuffer)
		runtime.KeepAlive(dh)
		if err != nil {
			if r := dh.Close(); r != nil {
				err = r
			}
			return
		}
		if n <= 0 {
			break
		}
		buf := scratchBuffer[:n]
		for len(buf) > 0 {
			de = (*syscall.Dirent)(unsafe.Pointer(&buf[0]))
			buf = buf[de.Reclen:]

			if de.Ino == 0 {
				continue
			}

			reg := int(uint64(de.Reclen) - uint64(unsafe.Offsetof(syscall.Dirent{}.Name)))

			var nameSlice []byte
			header := (*reflect.SliceHeader)(unsafe.Pointer(&nameSlice))
			header.Cap = reg
			header.Len = reg
			header.Data = uintptr(unsafe.Pointer(&de.Name[0]))

			if index := bytes.IndexByte(nameSlice, 0); index >= 0 {
				header.Cap = index
				header.Len = index
			}

			switch len(nameSlice) {
			case 0:
				continue
			case 1:
				if nameSlice[0] == '.' {
					continue
				}
			case 2:
				if nameSlice[0] == '.' && nameSlice[1] == '.' {
					continue
				}
			}
			result = append(result, string(nameSlice))
		}
	}

	if r := dh.Close(); r != nil {
		err = r
		return
	}

	if ascending {
		sort.Slice(result, func(i, j int) bool {
			return result[i] < result[j]
		})
	} else {
		sort.Slice(result, func(i, j int) bool {
			return result[i] > result[j]
		})
	}

	return
}

// CountFiles returns number of items in directory
func (storage PlaintextStorage) CountFiles(path string) (result int, err error) {
	var (
		n  int
		dh *os.File
		de *syscall.Dirent
	)

	dh, err = os.Open(filepath.Clean(storage.Root + "/" + path))
	if err != nil {
		return
	}

	fd := int(dh.Fd())

	scratchBuffer := make([]byte, storage.bufferSize)

	for {
		n, err = syscall.ReadDirent(fd, scratchBuffer)
		runtime.KeepAlive(dh)
		if err != nil {
			if r := dh.Close(); r != nil {
				err = r
			}
			return
		}
		if n <= 0 {
			break
		}
		buf := scratchBuffer[:n]
		for len(buf) > 0 {
			de = (*syscall.Dirent)(unsafe.Pointer(&buf[0]))
			buf = buf[de.Reclen:]
			if de.Ino == 0 || de.Type != syscall.DT_REG {
				continue
			}
			result++
		}
	}

	if r := dh.Close(); r != nil {
		err = r
	}

	return
}

// Exists returns true if absolute path exists
func (storage PlaintextStorage) Exists(path string) (bool, error) {
	var (
		trusted = new(syscall.Stat_t)
		cleaned = filepath.Clean(storage.Root + "/" + path)
		err     error
	)
	err = syscall.Stat(cleaned, trusted)
	if err == nil {
		return true, nil
	} else if os.IsNotExist(err) {
		return false, nil
	} else {
		return false, err
	}
}

// TouchFile creates files given absolute path if file does not already exist
func (storage PlaintextStorage) TouchFile(path string) error {
	cleanedPath := filepath.Clean(storage.Root + "/" + path)
	if err := os.MkdirAll(filepath.Dir(cleanedPath), os.ModePerm); err != nil {
		return err
	}
	f, err := os.OpenFile(cleanedPath, os.O_RDONLY|os.O_CREATE|os.O_EXCL, os.ModePerm)
	if err != nil {
		return err
	}
	defer f.Close()
	return nil
}

// GetFileReader creates file io.Reader
func (storage PlaintextStorage) GetFileReader(path string) (*fileReader, error) {
	f, err := os.OpenFile(filepath.Clean(storage.Root+"/"+path), os.O_RDONLY, os.ModePerm)
	if err != nil {
		return nil, err
	}

	reader := new(fileReader)
	reader.source = f

	return reader, nil
}

// ReadFileFully reads whole file given absolute path
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

// WriteFile writes data given absolute path to a file if that file does not
// already exists
func (storage PlaintextStorage) WriteFile(path string, data []byte) error {
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

// DeleteFile removes file given absolute path if that file does exists
func (storage PlaintextStorage) DeleteFile(path string) error {
	return os.Remove(filepath.Clean(storage.Root + "/" + path))
}

// UpdateFile rewrite file with data given absolute path to a file if that file
// exist
func (storage PlaintextStorage) UpdateFile(path string, data []byte) (err error) {
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
