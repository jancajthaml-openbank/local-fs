package storage

import (
	"crypto/rand"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExists(t *testing.T) {
	tmpDir := os.TempDir()

	file, err := ioutil.TempFile(tmpDir, "existent.*.tmp")
	require.Nil(t, err)
	filename := file.Name()
	defer os.Remove(filename)

	storage := NewPlaintextStorage(tmpDir)

	var ok bool
	var fail error

	ok, fail = storage.Exists(filepath.Base(filename))
	assert.Nil(t, fail)
	assert.True(t, ok)

	ok, fail = storage.Exists(filepath.Base(filename + "xxx"))
	assert.Nil(t, fail)
	assert.False(t, ok)
}

func TestReadFileFully(t *testing.T) {
	tmpDir := os.TempDir()

	file, err := ioutil.TempFile(tmpDir, "readable.*.tmp")
	require.Nil(t, err)
	filename := file.Name()
	basePath := filepath.Base(filename)
	defer os.Remove(filename)

	storage := NewPlaintextStorage(tmpDir)

	bigBuff := make([]byte, 75000)
	rand.Read(bigBuff)

	err = storage.WriteFile(basePath, bigBuff)
	require.Nil(t, err)

	var data []byte
	var fail error

	data, fail = storage.ReadFileFully(basePath)
	assert.Nil(t, fail)
	assert.Equal(t, len(bigBuff), len(data))
	assert.Equal(t, bigBuff, data)
}

func TestListDirectory(t *testing.T) {
	tmpDir := os.TempDir()

	tmpdir, err := ioutil.TempDir(tmpDir, "test_storage")
	require.Nil(t, err)
	defer os.RemoveAll(tmpdir)

	storage := NewPlaintextStorage(tmpDir)

	NewSlice := func(start, end, step int) []int {
		if step <= 0 || end < start {
			return []int{}
		}
		s := make([]int, 0, 1+(end-start)/step)
		for start <= end {
			s = append(s, start)
			start += step
		}
		return s
	}

	require.Nil(t, os.MkdirAll(tmpdir, os.ModePerm))
	defer os.RemoveAll(tmpdir)

	items := NewSlice(0, 10, 1)

	for _, i := range items {
		var file, _ = os.Create(fmt.Sprintf("%s/%010d", tmpdir, i))
		file.Close()
	}

	list, err := storage.ListDirectory(filepath.Base(tmpdir), true)
	require.Nil(t, err)

	assert.NotNil(t, list)
	assert.Equal(t, len(items), len(list))
	assert.Equal(t, fmt.Sprintf("%010d", items[0]), list[0])
	assert.Equal(t, fmt.Sprintf("%010d", items[len(items)-1]), list[len(list)-1])
}

func TestCountFiles(t *testing.T) {
	tmpDir := os.TempDir()

	tmpdir, err := ioutil.TempDir(tmpDir, "test_storage")
	require.Nil(t, err)
	defer os.RemoveAll(tmpdir)

	storage := NewPlaintextStorage(tmpDir)

	for i := 0; i < 60; i++ {
		file, err := os.Create(fmt.Sprintf("%s/%010dF", tmpdir, i))
		require.Nil(t, err)
		file.Close()
	}

	for i := 0; i < 40; i++ {
		err := os.MkdirAll(fmt.Sprintf("%s/%010dD", tmpdir, i), os.ModePerm)
		require.Nil(t, err)
	}

	numberOfFiles, err := storage.CountFiles(filepath.Base(tmpdir))
	require.Nil(t, err)
	assert.Equal(t, 60, numberOfFiles)
}

func BenchmarkCountFiles(b *testing.B) {
	tmpDir := os.TempDir()

	tmpdir, err := ioutil.TempDir(tmpDir, "test_storage")
	require.Nil(b, err)
	defer os.RemoveAll(tmpdir)

	storage := NewPlaintextStorage(tmpDir)

	for i := 0; i < 10000; i++ {
		file, err := os.Create(fmt.Sprintf("%s%010d", tmpdir, i))
		require.Nil(b, err)
		file.Close()
	}

	basePath := filepath.Base(tmpdir)

	b.ResetTimer()
	b.ReportAllocs()
	for n := 0; n < b.N; n++ {
		storage.CountFiles(basePath)
	}
}

func BenchmarkListDirectory(b *testing.B) {
	tmpDir := os.TempDir()

	tmpdir, err := ioutil.TempDir(tmpDir, "test_storage")
	require.Nil(b, err)
	defer os.RemoveAll(tmpdir)

	storage := NewPlaintextStorage(tmpDir)

	for i := 0; i < 1000; i++ {
		file, err := os.Create(fmt.Sprintf("%s%010d", tmpdir, i))
		require.Nil(b, err)
		file.Close()
	}

	basePath := filepath.Base(tmpdir)

	b.ResetTimer()
	b.ReportAllocs()
	for n := 0; n < b.N; n++ {
		storage.ListDirectory(basePath, true)
	}
}

func BenchmarkExists(b *testing.B) {
	tmpDir := os.TempDir()

	file, err := ioutil.TempFile(tmpDir, "exists.*")
	require.Nil(b, err)
	filename := file.Name()
	defer os.Remove(filename)

	storage := NewPlaintextStorage(tmpDir)
	basePath := filepath.Base(filename)

	b.ResetTimer()
	b.ReportAllocs()
	for n := 0; n < b.N; n++ {
		storage.Exists(basePath)
	}
}

func BenchmarkWriteFile(b *testing.B) {
	tmpDir := os.TempDir()

	file, err := ioutil.TempFile(tmpDir, "updated.*")
	require.Nil(b, err)
	filename := file.Name()
	defer os.Remove(filename)

	storage := NewPlaintextStorage(tmpDir)
	basePath := filepath.Base(filename)
	bigBuff := make([]byte, 1024)
	rand.Read(bigBuff)

	b.ResetTimer()
	b.ReportAllocs()
	b.SetBytes(int64(len(bigBuff)))
	for n := 0; n < b.N; n++ {
		storage.WriteFile(basePath, bigBuff)
	}
}

func BenchmarkAppendFile(b *testing.B) {
	tmpDir := os.TempDir()

	file, err := ioutil.TempFile(tmpDir, "appended.*")
	require.Nil(b, err)
	filename := file.Name()
	defer os.Remove(filename)

	storage := NewPlaintextStorage(tmpDir)
	basePath := filepath.Base(filename)
	bigBuff := make([]byte, 1024)
	rand.Read(bigBuff)

	b.ResetTimer()
	b.ReportAllocs()
	b.SetBytes(int64(len(bigBuff)))
	for n := 0; n < b.N; n++ {
		storage.AppendFile(basePath, bigBuff)
	}
}

func BenchmarkReadFileFully(b *testing.B) {
	tmpDir := os.TempDir()

	file, err := ioutil.TempFile(tmpDir, "readable.*")
	require.Nil(b, err)
	filename := file.Name()
	defer os.Remove(filename)

	storage := NewPlaintextStorage(tmpDir)
	basePath := filepath.Base(filename)

	bigBuff := make([]byte, 1024)
	rand.Read(bigBuff)

	err = ioutil.WriteFile(filename, bigBuff, os.ModePerm)
	require.Nil(b, err)

	b.ResetTimer()
	b.ReportAllocs()
	b.SetBytes(int64(len(bigBuff)))
	for n := 0; n < b.N; n++ {
		storage.ReadFileFully(basePath)
	}
}
