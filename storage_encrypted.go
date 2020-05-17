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
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"syscall"
)

// EncryptedStorage is a fascade to access encrypted storage
type EncryptedStorage struct {
	Root          string
	bufferSize    int
	encryptionKey []byte
}

// NewEncryptedStorage returns new storage over given root
func NewEncryptedStorage(root string, key []byte) EncryptedStorage {
	if root == "" || os.MkdirAll(filepath.Clean(root), os.ModePerm) != nil {
		panic("unable to assert root storage directory")
	}
	if len(key) == 0 {
		panic("no encryption key setup")
	}
	return EncryptedStorage{
		Root:          root,
		bufferSize:    8192,
		encryptionKey: key,
	}
}

// Encrypt data with encryption key
func (storage EncryptedStorage) Encrypt(data []byte) ([]byte, error) {
	block, err := aes.NewCipher(storage.encryptionKey)
	if err != nil {
		return nil, err
	}
	ciphertext := make([]byte, aes.BlockSize+len(data))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, err
	}
	cfb := cipher.NewCFBEncrypter(block, iv)
	cfb.XORKeyStream(ciphertext[aes.BlockSize:], []byte(data))
	return ciphertext, nil
}

// Decrypt data with encryption key
func (storage EncryptedStorage) Decrypt(data []byte) ([]byte, error) {
	block, err := aes.NewCipher(storage.encryptionKey)
	if err != nil {
		return nil, err
	}
	if len(data) < aes.BlockSize {
		return nil, fmt.Errorf("invalid blocksize expected %d but actual is %d", aes.BlockSize, len(data))
	}
	plaintext := make([]byte, len(data))
	copy(plaintext, data)
	iv := plaintext[:aes.BlockSize]
	plaintext = plaintext[aes.BlockSize:]
	cfb := cipher.NewCFBDecrypter(block, iv)
	cfb.XORKeyStream(plaintext, plaintext)
	return plaintext, nil
}

// ListDirectory returns sorted slice of item names in given absolute path
// default sorting is ascending
func (storage EncryptedStorage) ListDirectory(path string, ascending bool) ([]string, error) {
	return listDirectory(storage.Root+"/"+path, storage.bufferSize, ascending)
}

// CountFiles returns number of items in directory
func (storage EncryptedStorage) CountFiles(path string) (int, error) {
	return countFiles(storage.Root+"/"+path, storage.bufferSize)
}

// Exists returns true if path exists
func (storage EncryptedStorage) Exists(path string) (bool, error) {
	return nodeExists(storage.Root + "/" + path)
}

// TouchFile creates files given absolute path if file does not already exist
func (storage EncryptedStorage) TouchFile(path string) error {
	return touch(storage.Root + "/" + path)
}

// DeleteFile removes file given absolute path if that file does exists
func (storage EncryptedStorage) DeleteFile(path string) error {
	return os.Remove(filepath.Clean(storage.Root + "/" + path))
}

// ReadFileFully reads whole file given path
func (storage EncryptedStorage) ReadFileFully(path string) ([]byte, error) {
	filename := filepath.Clean(storage.Root + "/" + path)
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
	_, err = syscall.Read(fd, buf)
	if err != nil && err != io.EOF {
		return nil, err
	}
	// FIXME inline
	return storage.Decrypt(buf)
}

// WriteFileExclusive writes data given path to a file if that file does not
// already exists
func (storage EncryptedStorage) WriteFileExclusive(path string, data []byte) error {
	filename := filepath.Clean(storage.Root + "/" + path)
	if err := os.MkdirAll(filename, 0600); err != nil {
		return err
	}
	// FIXME inline
	out, err := storage.Encrypt(data)
	if err != nil {
		return err
	}
	fd, err := syscall.Open(filename, syscall.O_CREAT|syscall.O_WRONLY|syscall.O_EXCL|syscall.O_NONBLOCK, 0600)
	if err != nil {
		return err
	}
	defer syscall.Close(fd)
	if err = syscall.Flock(fd, syscall.LOCK_EX); err != nil {
		return err
	}
	defer syscall.Flock(fd, syscall.LOCK_UN)
	if _, err := syscall.Write(fd, out); err != nil {
		return err
	}
	return nil
}

// WriteFile writes data given absolute path to a file, creates it if it does
// not exist
func (storage EncryptedStorage) WriteFile(path string, data []byte) error {
	filename := filepath.Clean(storage.Root + "/" + path)
	if err := os.MkdirAll(filename, 0600); err != nil {
		return err
	}
	// FIXME inline
	out, err := storage.Encrypt(data)
	if err != nil {
		return err
	}
	fd, err := syscall.Open(filename, syscall.O_CREAT|syscall.O_WRONLY|syscall.O_TRUNC|syscall.O_NONBLOCK, 0600)
	if err != nil {
		return err
	}
	defer syscall.Close(fd)
	if err = syscall.Flock(fd, syscall.LOCK_EX); err != nil {
		return err
	}
	defer syscall.Flock(fd, syscall.LOCK_UN)
	if _, err := syscall.Write(fd, out); err != nil {
		return err
	}
	return nil
}

// AppendFile appens data given absolute path to a file, creates it if it does
// not exist
func (storage EncryptedStorage) AppendFile(path string, data []byte) error {
	filename := filepath.Clean(storage.Root + "/" + path)
	if err := os.MkdirAll(filename, 0600); err != nil {
		return err
	}
	fd, err := syscall.Open(filename, syscall.O_CREAT|syscall.O_WRONLY|syscall.O_TRUNC|syscall.O_NONBLOCK, 0600)
	if err != nil {
		return err
	}
	defer syscall.Close(fd)
	if err = syscall.Flock(fd, syscall.LOCK_EX); err != nil {
		return err
	}
	defer syscall.Flock(fd, syscall.LOCK_UN)
	var fs syscall.Stat_t
	if err = syscall.Fstat(fd, &fs); err != nil {
		return err
	}
	buf := make([]byte, fs.Size)
	_, err = syscall.Read(fd, buf)
	if err != nil && err != io.EOF {
		return err
	}
	// FIXME inline
	head, err := storage.Decrypt(buf)
	if err != nil {
		return err
	}
	var tail = make([]byte, len(head)+1)
	tail = append(tail, head...)
	tail = append(tail, data...)
	// FIXME inline
	out, err := storage.Encrypt(tail)
	if err != nil {
		return err
	}
	if _, err := syscall.Write(fd, out); err != nil {
		return err
	}
	return nil
}
