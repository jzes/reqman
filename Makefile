.PHONY: test coverage

test:
	go test ./...

coverage:
	sh scripts/check-coverage.sh
