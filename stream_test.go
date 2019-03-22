package storage

import (
	"crypto/rand"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFileReader(t *testing.T) {
	tmpDir := os.TempDir()

	file, err := ioutil.TempFile(tmpDir, "readable.*.tmp")
	require.Nil(t, err)
	filename := file.Name()
	basePath := filepath.Base(filename)
	defer os.Remove(filename)

	storage := NewStorage(tmpDir)

	bigBuff := make([]byte, 75000)
	rand.Read(bigBuff)

	err = ioutil.WriteFile(filename, bigBuff, os.ModePerm)
	require.Nil(t, err)

	var data []byte
	var fail error

	reader, fail := storage.GetFileReader(basePath)
	assert.Nil(t, fail)
	assert.NotNil(t, reader)

	data = make([]byte, len(bigBuff))
	n, fail := io.ReadFull(reader, data)
	assert.Equal(t, nil, fail)
	assert.Equal(t, len(bigBuff), n)
	assert.Equal(t, len(bigBuff), len(data))
	assert.Equal(t, bigBuff, data)
}

func BenchmarkFileReader(b *testing.B) {
	tmpDir := os.TempDir()

	file, err := ioutil.TempFile(tmpDir, "readable.*")
	require.Nil(b, err)
	filename := file.Name()
	defer os.Remove(filename)

	storage := NewStorage(tmpDir)
	basePath := filepath.Base(filename)

	bigBuff := make([]byte, 75000)
	rand.Read(bigBuff)

	err = ioutil.WriteFile(filename, bigBuff, os.ModePerm)
	require.Nil(b, err)

	data := make([]byte, len(bigBuff))

	reader, fail := storage.GetFileReader(basePath)
	assert.Nil(b, fail)
	assert.NotNil(b, reader)

	b.ResetTimer()
	b.ReportAllocs()
	b.SetBytes(int64(len(bigBuff)))
	for n := 0; n < b.N; n++ {
		reader, _ := storage.GetFileReader(basePath)
		io.ReadFull(reader, data)
	}
}
