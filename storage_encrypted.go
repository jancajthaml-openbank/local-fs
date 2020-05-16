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
)

// EncryptedStorage is a fascade to access encrypted storage
type EncryptedStorage struct {
	underlying    PlaintextStorage
	encryptionKey []byte
}

// NewEncryptedStorage returns new storage over given root
func NewEncryptedStorage(root string, key []byte) EncryptedStorage {
	return EncryptedStorage{
		underlying:    NewPlaintextStorage(root),
		encryptionKey: key,
	}
}

// Encrypt data with encryption key
func (storage EncryptedStorage) Encrypt(data []byte) ([]byte, error) {
	if len(storage.encryptionKey) == 0 {
		return nil, fmt.Errorf("no encryption key setup")
	}
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
	if len(storage.encryptionKey) == 0 {
		return nil, fmt.Errorf("no encryption key setup")
	}
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

func (storage EncryptedStorage) ListDirectory(path string, ascending bool) ([]string, error) {
	return storage.underlying.ListDirectory(path, ascending)
}

func (storage EncryptedStorage) CountFiles(path string) (int, error) {
	return storage.underlying.CountFiles(path)
}

func (storage EncryptedStorage) Exists(path string) (bool, error) {
	return storage.underlying.Exists(path)
}

func (storage EncryptedStorage) TouchFile(path string) error {
	return storage.underlying.TouchFile(path)
}

func (storage EncryptedStorage) DeleteFile(path string) error {
	return storage.underlying.DeleteFile(path)
}

func (storage EncryptedStorage) GetFileReader(path string) (*fileReader, error) {
	return storage.underlying.GetFileReader(path)
}

func (storage EncryptedStorage) ReadFileFully(path string) ([]byte, error) {
	in, err := storage.underlying.ReadFileFully(path)
	if err != nil {
		return nil, err
	}
	return storage.Decrypt(in)
}

func (storage EncryptedStorage) WriteFile(path string, data []byte) error {
	out, err := storage.Encrypt(data)
	if err != nil {
		return err
	}
	return storage.WriteFile(path, out)
}

func (storage EncryptedStorage) UpdateFile(path string, data []byte) error {
	out, err := storage.Encrypt(data)
	if err != nil {
		return err
	}
	return storage.UpdateFile(path, out)
}

func (storage EncryptedStorage) AppendFile(path string, data []byte) error {
	// FIXME inline and use flag
	// |os.O_EXCL

	// FIXME mutex!
	in, err := storage.underlying.ReadFileFully(path)
	if err != nil {
		return err
	}
	head, err := storage.Decrypt(in)
	if err != nil {
		return err
	}

	var tail = make([]byte, len(head)+1)
	tail = append(tail, head...)
	tail = append(tail, data...)

	out, err := storage.Encrypt(tail)
	if err != nil {
		return err
	}
	return storage.UpdateFile(path, out)
}
