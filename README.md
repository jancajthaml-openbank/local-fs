# local-fs

Local File System Database

![Health Check](https://github.com/jancajthaml-openbank/local-fs/workflows/Health%20Check/badge.svg)

[![godoc for jancajthaml-openbank/local-fs](https://godoc.org/github.com/nathany/looper?status.svg)](https://godoc.org/github.com/jancajthaml-openbank/local-fs)

## Usage

```go
import (
  localfs "github.com/jancajthaml-openbank/local-fs"
)

...

storage := localfs.NewStorage("/tmp")

// list nodes at /tmp/foo in ascending order
asc, err := storage.ListDirectory("foo", true)

// list nodes at /tmp/foo in descengin order
desc, err := storage.ListDirectory("foo", false)

// check if /tmp/foo exists
ok, err := storage.Exists("foo")

// delete file /tmp/foo
err := storage.DeleteFile("foo")

// creates file /tmp/foo if not exists
err := storage.TouchFile("foo")

// ovewrites file /tmp/foo with "abc", creates file if it does not exist
err := storage.WriteFile("foo", []byte("abc"))

// ovewrites file /tmp/foo with "abc", fails if file does not exist
err := storage.UpdateFile("foo", []byte("abc"))

// appends "abc" to end of /tmp/foo file, fails if file does not exist
err := storage.AppendFile("foo", []byte("abc"))

// read all bytes of file /tmp/foo
data, err := storage.ReadFileFully("tmp")

// read all bytes of file /tmp/foo
data, err := storage.ReadFileFully("tmp")

// returns reader for /tmp/foo
fd, err := storage.GetFileReader("tmp")
```

## Encryption of data at rest

Generate some key

```bash
openssl rand -hex 32 | xargs --no-run-if-empty echo -n > /tmp/secrets/key
```

Write and read encrypted data at /tmp/data/foo

```go
import (
  localfs "github.com/jancajthaml-openbank/local-fs"
)

...

storage := localfs.NewStorage("/tmp/data")
storage.SetEncryptionKey("/tmp/secrets/key")

out, err := storage.Encrypt([]byte("pii"))
err := storage.WriteFile("foo", out)

...

in, err := storage.ReadFileFully("/tmp/data/foo")
data, err := storage.Decrypt(in)
```
