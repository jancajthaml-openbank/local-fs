package storage

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func getKey() []byte {
	src := []byte("cf434a97e34dc7a7feb918de8dfdbfbe10397bcbdcb84ca6779df518c264ad8d")
	dst := make([]byte, hex.DecodedLen(len(src)))
	n, _ := hex.Decode(dst, src)
	return dst[:n]
}

func TestExistsEncrypted(t *testing.T) {
	tmpDir := os.TempDir()

	file, err := ioutil.TempFile(tmpDir, "existent.*.tmp")
	if err != nil {
		t.Fatalf("unexpected error when creating temp file %+v", err)
	}

	filename := file.Name()
	defer os.Remove(filename)

	storage := NewEncryptedStorage(tmpDir, getKey())

	var ok bool

	ok, err = storage.Exists(filepath.Base(filename))
	if err != nil {
		t.Errorf("unexpected error when calling Exists %+v", err)
	}
	if !ok {
		t.Errorf("expected Exists to return true for existent file")
	}

	ok, err = storage.Exists(filepath.Base(filename + "xxx"))

	if err != nil {
		t.Errorf("unexpected error when calling Exists %+v", err)
	}
	if ok {
		t.Errorf("expected Exists to return false for non existent file")
	}
}

func TestReadFileFullyEncrypted(t *testing.T) {
	tmpDir := os.TempDir()

	file, err := ioutil.TempFile(tmpDir, "readable.*.tmp")
	if err != nil {
		t.Fatalf("unexpected error when creating temp file %+v", err)
	}

	filename := file.Name()
	basePath := filepath.Base(filename)
	defer os.Remove(filename)

	storage := NewEncryptedStorage(tmpDir, getKey())

	bigBuff := make([]byte, 75000)
	rand.Read(bigBuff)

	err = storage.WriteFile(basePath, bigBuff)
	if err != nil {
		t.Fatalf("unexpected error when calling WriteFile %+v", err)
	}

	var data []byte

	data, err = storage.ReadFileFully(basePath)

	if err != nil {
		t.Errorf("unexpected error when calling ReadFileFully %+v", err)
	}
	if len(bigBuff) != len(data) {
		t.Errorf("expected to read %d bytes but red %d instead", len(data), len(bigBuff))
	}
	if string(bigBuff) != string(data) {
		t.Errorf("expected to read %s but got %s instead", string(data), string(bigBuff))
	}
}

func TestListDirectoryEncrypted(t *testing.T) {
	tmpDir := os.TempDir()

	tmpdir, err := ioutil.TempDir(tmpDir, "test_storage")
	if err != nil {
		t.Fatalf("unexpected error when creating temp file %+v", err)
	}
	defer os.RemoveAll(tmpdir)

	storage := NewEncryptedStorage(tmpDir, getKey())

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

	err = os.MkdirAll(tmpdir, os.ModePerm)
	if err != nil {
		t.Fatalf("unexpected error when asserting directories %+v", err)
	}
	defer os.RemoveAll(tmpdir)

	items := NewSlice(0, 10, 1)

	for _, i := range items {
		var file, _ = os.Create(fmt.Sprintf("%s/%010d", tmpdir, i))
		file.Close()
	}

	list, err := storage.ListDirectory(filepath.Base(tmpdir), true)
	if err != nil {
		t.Fatalf("unexpected error when calling ListDirectory %+v", err)
	}

	if list == nil {
		t.Errorf("expected slice got nothing")
	}
	if len(items) != len(list) {
		t.Errorf("expected to get %d files got %d instead", len(items), len(list))
	}
	if fmt.Sprintf("%010d", items[0]) != list[0] {
		t.Errorf("expected first item to be %s got %s instead", fmt.Sprintf("%010d", items[0]), list[0])
	}
	if fmt.Sprintf("%010d", items[len(items)-1]) != list[len(list)-1] {
		t.Errorf("expected last item to be %s got %s instead", fmt.Sprintf("%010d", items[len(items)-1]), list[len(list)-1])
	}
}

func TestCountFilesEncrypted(t *testing.T) {
	tmpDir := os.TempDir()

	tmpdir, err := ioutil.TempDir(tmpDir, "test_storage")
	if err != nil {
		t.Fatalf("unexpected error when creating temp file %+v", err)
	}
	defer os.RemoveAll(tmpdir)

	storage := NewEncryptedStorage(tmpDir, getKey())

	for i := 0; i < 60; i++ {
		file, err := os.Create(fmt.Sprintf("%s/%010dF", tmpdir, i))
		if err != nil {
			t.Fatalf("unexpected error when creating temp file %+v", err)
		}
		file.Close()
	}

	for i := 0; i < 40; i++ {
		err := os.MkdirAll(fmt.Sprintf("%s/%010dD", tmpdir, i), os.ModePerm)
		if err != nil {
			t.Fatalf("unexpected error when asserting directories %+v", err)
		}
	}

	numberOfFiles, err := storage.CountFiles(filepath.Base(tmpdir))
	if err != nil {
		t.Fatalf("unexpected error when calling CountFiles %+v", err)
	}
	if numberOfFiles != 60 {
		t.Errorf("expected to count 60 files, counted %d instead", numberOfFiles)
	}
}

func BenchmarkCountFilesEncrypted(b *testing.B) {
	tmpDir := os.TempDir()

	tmpdir, err := ioutil.TempDir(tmpDir, "test_storage")
	if err != nil {
		b.Fatalf("unexpected error when creating temp file %+v", err)
	}
	defer os.RemoveAll(tmpdir)

	storage := NewEncryptedStorage(tmpDir, getKey())

	for i := 0; i < 10000; i++ {
		file, err := os.Create(fmt.Sprintf("%s%010d", tmpdir, i))
		if err != nil {
			b.Fatalf("unexpected error when creating temp file %+v", err)
		}
		file.Close()
	}

	basePath := filepath.Base(tmpdir)

	b.ResetTimer()
	b.ReportAllocs()
	for n := 0; n < b.N; n++ {
		storage.CountFiles(basePath)
	}
}

func BenchmarkListDirectoryEncrypted(b *testing.B) {
	tmpDir := os.TempDir()

	tmpdir, err := ioutil.TempDir(tmpDir, "test_storage")
	if err != nil {
		b.Fatalf("unexpected error when creating temp file %+v", err)
	}
	defer os.RemoveAll(tmpdir)

	storage := NewEncryptedStorage(tmpDir, getKey())

	for i := 0; i < 1000; i++ {
		file, err := os.Create(fmt.Sprintf("%s%010d", tmpdir, i))
		if err != nil {
			b.Fatalf("unexpected error when creating temp file %+v", err)
		}
		file.Close()
	}

	basePath := filepath.Base(tmpdir)

	b.ResetTimer()
	b.ReportAllocs()
	for n := 0; n < b.N; n++ {
		storage.ListDirectory(basePath, true)
	}
}

func BenchmarkExistsEncrypted(b *testing.B) {
	tmpDir := os.TempDir()

	file, err := ioutil.TempFile(tmpDir, "exists.*")
	if err != nil {
		b.Fatalf("unexpected error when creating temp file %+v", err)
	}
	filename := file.Name()
	defer os.Remove(filename)

	storage := NewEncryptedStorage(tmpDir, getKey())
	basePath := filepath.Base(filename)

	b.ResetTimer()
	b.ReportAllocs()
	for n := 0; n < b.N; n++ {
		storage.Exists(basePath)
	}
}

func BenchmarkWriteFileEncrypted(b *testing.B) {
	tmpDir := os.TempDir()

	file, err := ioutil.TempFile(tmpDir, "updated.*")
	if err != nil {
		b.Fatalf("unexpected error when creating temp file %+v", err)
	}
	filename := file.Name()
	defer os.Remove(filename)

	storage := NewEncryptedStorage(tmpDir, getKey())
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

func BenchmarkAppendFileEncrypted(b *testing.B) {
	tmpDir := os.TempDir()

	file, err := ioutil.TempFile(tmpDir, "appended.*")
	if err != nil {
		b.Fatalf("unexpected error when creating temp file %+v", err)
	}
	filename := file.Name()
	defer os.Remove(filename)

	storage := NewEncryptedStorage(tmpDir, getKey())
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

func BenchmarkReadFileFullyEncrypted(b *testing.B) {
	tmpDir := os.TempDir()

	file, err := ioutil.TempFile(tmpDir, "readable.*")
	if err != nil {
		b.Fatalf("unexpected error when creating temp file %+v", err)
	}
	filename := file.Name()
	defer os.Remove(filename)

	storage := NewEncryptedStorage(tmpDir, getKey())
	basePath := filepath.Base(filename)

	bigBuff := make([]byte, 1024)
	rand.Read(bigBuff)

	err = ioutil.WriteFile(filename, bigBuff, os.ModePerm)
	if err != nil {
		b.Fatalf("unexpected error when writing to file %+v", err)
	}

	b.ResetTimer()
	b.ReportAllocs()
	b.SetBytes(int64(len(bigBuff)))
	for n := 0; n < b.N; n++ {
		storage.ReadFileFully(basePath)
	}
}
