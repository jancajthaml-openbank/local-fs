.ONESHELL:

.PHONY: all
all: sync test lint sec

.PHONY: lint
lint:
	@docker-compose run --rm lint --pkg local-fs || :

.PHONY: sec
sec:
	@docker-compose run --rm sec --pkg local-fs || :

.PHONY: sync
sync:
	@docker-compose run --rm sync --pkg local-fs

.PHONY: test
test:
	@docker-compose run --rm test --pkg local-fs
