# local-fs

Local File System Database

![Health Check](https://github.com/jancajthaml-openbank/local-fs/workflows/Health%20Check/badge.svg)

[![godoc for jancajthaml-openbank/local-fs](https://godoc.org/github.com/nathany/looper?status.svg)](https://godoc.org/github.com/jancajthaml-openbank/local-fs)

## Usage

both plaintext and encrypted storage have same api

```go
import (
  localfs "github.com/jancajthaml-openbank/local-fs"
)

...

storage := localfs.NewPlaintextStorage("/tmp")

// list nodes at /tmp/foo in ascending order
asc, err := storage.ListDirectory("foo", true)

// list nodes at /tmp/foo in descending order
desc, err := storage.ListDirectory("foo", false)

// count files at /tmp/foo
count, err := storage.CountFiles("foo")

// check if /tmp/foo exists
ok, err := storage.Exists("foo")

// delete file /tmp/foo
err := storage.DeleteFile("foo")

// creates file /tmp/foo if not exists
err := storage.TouchFile("foo")

// ovewrites file /tmp/foo with "abc", creates file if it does not exist
err := storage.WriteFile("foo", []byte("abc"))

// crates and writes file /tmp/foo with "abc", fails if file exists
err := storage.WriteFileExclusive("foo", []byte("abc"))

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

storage := localfs.NewEncryptedStorage("/tmp/data", "/tmp/secrets/key")
err := storage.WriteFile("foo", []byte("pii"))

...

data, err := storage.ReadFileFully("/tmp/data/foo")
```

## License

Licensed under Apache 2.0 see LICENSE.md for details

[![FOSSA Status](https://app.fossa.com/api/projects/git%2Bgithub.com%2Fjancajthaml-openbank%2Flocal-fs.svg?type=large)](https://app.fossa.com/projects/git%2Bgithub.com%2Fjancajthaml-openbank%2Flocal-fs?ref=badge_large)
