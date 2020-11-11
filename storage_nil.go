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
	"time"
)

// NilStorage is a nil storage fascade
type NilStorage struct {
	Storage
}

// ListDirectory stub
func (storage NilStorage) ListDirectory(path string, ascending bool) ([]string, error) {
	return nil, fmt.Errorf("storage not initialized properly")
}

// CountFiles stub
func (storage NilStorage) CountFiles(path string) (int, error) {
	return 0, fmt.Errorf("storage not initialized properly")
}

// Exists stub
func (storage NilStorage) Exists(path string) (bool, error) {
	return false, fmt.Errorf("storage not initialized properly")
}

// LastModification stub
func (storage NilStorage) LastModification(path string) (time.Time, error) {
	return time.Now(), fmt.Errorf("storage not initialized properly")
}

// TouchFile stub
func (storage NilStorage) TouchFile(path string) error {
	return fmt.Errorf("storage not initialized properly")
}

// DeleteFile stub
func (storage NilStorage) DeleteFile(path string) error {
	return fmt.Errorf("storage not initialized properly")
}

// ReadFileFully stub
func (storage NilStorage) ReadFileFully(path string) ([]byte, error) {
	return nil, fmt.Errorf("storage not initialized properly")
}

// WriteFileExclusive stub
func (storage NilStorage) WriteFileExclusive(path string, data []byte) error {
	return fmt.Errorf("storage not initialized properly")
}

// WriteFile stub
func (storage NilStorage) WriteFile(path string, data []byte) error {
	return fmt.Errorf("storage not initialized properly")
}

// AppendFile stub
func (storage NilStorage) AppendFile(path string, data []byte) error {
	return fmt.Errorf("storage not initialized properly")
}
