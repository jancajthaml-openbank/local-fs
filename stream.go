// Copyright (c) 2016-2019, Jan Cajthaml <jan.cajthaml@gmail.com>
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
)

type fileReader struct {
	source *os.File
}

func (reader *fileReader) Read(p []byte) (int, error) {
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
