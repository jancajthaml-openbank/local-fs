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
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"sort"
	"syscall"
	"unsafe"
)

func nameFromDirent(dirent *syscall.Dirent) []byte {
	reg := int(uint64(dirent.Reclen) - uint64(unsafe.Offsetof(syscall.Dirent{}.Name)))

	var name []byte
	header := (*reflect.SliceHeader)(unsafe.Pointer(&name))
	header.Cap = reg
	header.Len = reg
	header.Data = uintptr(unsafe.Pointer(&dirent.Name[0]))

	if index := bytes.IndexByte(name, 0); index >= 0 {
		header.Cap = index
		header.Len = index
	}

	return name
}

func listDirectory(abspath string, bufferSize int, ascending bool) (result []string, err error) {
	var (
		n  int
		dh *os.File
		de *syscall.Dirent
	)

	dh, err = os.Open(filepath.Clean(abspath))
	if err != nil {
		return
	}

	fd := int(dh.Fd())
	result = make([]string, 0)

	scratchBuffer := make([]byte, bufferSize)

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

func countFiles(absPath string, bufferSize int) (result int, err error) {
	var (
		n  int
		dh *os.File
		de *syscall.Dirent
	)

	dh, err = os.Open(filepath.Clean(absPath))
	if err != nil {
		return
	}

	fd := int(dh.Fd())

	scratchBuffer := make([]byte, bufferSize)

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

func nodeExists(absPath string) (bool, error) {
	var (
		trusted = new(syscall.Stat_t)
		cleaned = filepath.Clean(absPath)
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

func touch(absPath string) error {
	cleanedPath := filepath.Clean(absPath)
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
