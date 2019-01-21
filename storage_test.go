package storage

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var tmpDir = os.TempDir()

func testPad(version int) string {
	return fmt.Sprintf("%010d", version)
}

func TestExists(t *testing.T) {
	file, err := ioutil.TempFile(tmpDir, "existant.*.tmp")
	require.Nil(t, err)
	filename := file.Name()
	defer os.Remove(filename)

	storage := NewStorage(tmpDir)

	var ok bool
	var fail error

	ok, fail = storage.Exists(filepath.Base(filename))
	assert.Nil(t, fail)
	assert.True(t, ok)

	ok, fail = storage.Exists(filepath.Base(filename + "xxx"))
	assert.Nil(t, fail)
	assert.False(t, ok)
}

func TestListDirectory(t *testing.T) {
	tmpdir, err := ioutil.TempDir(tmpDir, "test_storage")
	require.Nil(t, err)
	defer os.RemoveAll(tmpdir)

	storage := NewStorage(tmpDir)

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
		var file, _ = os.Create(tmpdir + "/" + testPad(i))
		file.Close()
	}

	list, err := storage.ListDirectory(filepath.Base(tmpdir), true)
	require.Nil(t, err)

	assert.NotNil(t, list)
	assert.Equal(t, len(items), len(list))
	assert.Equal(t, testPad(items[0]), list[0])
	assert.Equal(t, testPad(items[len(items)-1]), list[len(list)-1])
}

func TestCountFiles(t *testing.T) {
	tmpdir, err := ioutil.TempDir(tmpDir, "test_storage")
	require.Nil(t, err)
	defer os.RemoveAll(tmpdir)

	storage := NewStorage(tmpDir)

	for i := 0; i < 60; i++ {
		file, err := os.Create(tmpdir + "/" + testPad(i) + "F")
		require.Nil(t, err)
		file.Close()
	}

	for i := 0; i < 40; i++ {
		err := os.MkdirAll(tmpdir+"/"+testPad(i)+"D", os.ModePerm)
		require.Nil(t, err)
	}

	numberOfFiles, err := storage.CountFiles(filepath.Base(tmpdir))
	require.Nil(t, err)
	assert.Equal(t, 60, numberOfFiles)
}

func BenchmarkCountFiles(b *testing.B) {
	tmpdir, err := ioutil.TempDir(tmpDir, "test_storage")
	require.Nil(b, err)
	defer os.RemoveAll(tmpdir)

	storage := NewStorage(tmpDir)

	for i := 0; i < 1000; i++ {
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
	tmpdir, err := ioutil.TempDir(tmpDir, "test_storage")
	require.Nil(b, err)
	defer os.RemoveAll(tmpdir)

	storage := NewStorage(tmpDir)

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
	file, err := ioutil.TempFile(tmpDir, "exists.*")
	require.Nil(b, err)
	filename := file.Name()
	defer os.Remove(filename)

	storage := NewStorage(tmpDir)
	basePath := filepath.Base(filename)

	b.ResetTimer()
	b.ReportAllocs()
	for n := 0; n < b.N; n++ {
		storage.Exists(basePath)
	}
}

func BenchmarkUpdateFile(b *testing.B) {
	file, err := ioutil.TempFile(tmpDir, "updated.*")
	require.Nil(b, err)
	filename := file.Name()
	defer os.Remove(filename)

	storage := NewStorage(tmpDir)
	basePath := filepath.Base(filename)
	data := []byte("abcd")

	b.ResetTimer()
	b.ReportAllocs()
	for n := 0; n < b.N; n++ {
		storage.UpdateFile(basePath, data)
	}
}

func BenchmarkAppendFile(b *testing.B) {
	file, err := ioutil.TempFile(tmpDir, "appended.*")
	require.Nil(b, err)
	filename := file.Name()
	defer os.Remove(filename)

	storage := NewStorage(tmpDir)
	basePath := filepath.Base(filename)
	data := []byte("abcd")

	b.ResetTimer()
	b.ReportAllocs()
	for n := 0; n < b.N; n++ {
		storage.AppendFile(basePath, data)
	}
}

func BenchmarkReadFileFully(b *testing.B) {
	file, err := ioutil.TempFile(tmpDir, "updated.*")
	require.Nil(b, err)
	filename := file.Name()
	defer os.Remove(filename)

	storage := NewStorage(tmpDir)
	basePath := filepath.Base(filename)

	require.Nil(b, storage.UpdateFile(basePath, []byte("abcd")))

	b.ResetTimer()
	b.ReportAllocs()
	for n := 0; n < b.N; n++ {
		storage.ReadFileFully(basePath)
	}
}
