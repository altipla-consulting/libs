
FILES = $(shell find . -type f -name '*.go' -not -path './vendor/*')

gofmt:
	@gofmt -w $(FILES)
	@gofmt -r '&α{} -> new(α)' -w $(FILES)

test: gofmt
	revive -formatter friendly
	go test ./...

update-deps:
	go get -u
	go mod download
	go mod tidy
