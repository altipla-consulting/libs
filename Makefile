
FILES = $(shell find . -type f -name '*.go' -not -path './vendor/*')

.PHONY: protos

gofmt:
	@gofmt -s -w $(FILES)
	@gofmt -r '&α{} -> new(α)' -w $(FILES)
	@impsort . -p libs.altipla.consulting

test: lint
	go test -race ./...

lint:
	go install ./...
	go vet ./...
	linter ./...

update-deps:
	go get -u
	go mod download
	go mod tidy

protos:
	actools protoc --go_out=paths=source_relative:. ./protos/datetime/datetime.proto
	actools protoc --go_out=paths=source_relative:. ./bigquery/proto/pagination.proto

data:
	docker-compose kill database redis firestore ravendb
	docker-compose up -d database redis firestore ravendb
	bash -c "until actools mysql -h database -u dev-user -pdev-password -e ';' 2> /dev/null ; do sleep 1; done"

datetime-generator: _datetime-generator gofmt

_datetime-generator:
	wget http://www.unicode.org/Public/cldr/27.0.1/core.zip -O /tmp/core.zip
	go install ./cmd/datetime-generator
	datetime-generator -locales en,es,fr,ru,de,it,ja,pt
