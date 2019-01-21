.ONESHELL:

.PHONY: all
all: test

.PHONY: test
test:
	GOMAXPROCS=1 \
	go test -v ./... -benchmem -bench=. -timeout=20s
