// Copyright (c) 2017-2023, Jan Cajthaml <jan.cajthaml@gmail.com>
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
	"os"
	"time"
)

// Storage represents contract
type Storage interface {
	Chmod(absPath string, mod os.FileMode) error
	ListDirectory(string, bool) ([]string, error)
	CountFiles(string) (int, error)
	Exists(string) (bool, error)
	TouchFile(string) error
	Mkdir( string) error
	ReadFileFully(string) ([]byte, error)
	WriteFileExclusive(string, []byte) error
	WriteFile(string, []byte) error
	Delete(string) error
	AppendFile(string, []byte) error
	LastModification(string) (time.Time, error)
}
