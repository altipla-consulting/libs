
FILES = $(shell find . -type f -name '*.go' -not -path './vendor/*')

.PHONY: protos

gofmt:
	@gofmt -s -w $(FILES)
	@gofmt -r '&α{} -> new(α)' -w $(FILES)

test: lint
	go test ./...

lint:
	@./infra/lint-errors.sh
	tool/linter ./...
	go vet ./...
	go install ./...

update-deps:
	go get -u
	go mod download
	go mod tidy

protos:
	actools protoc --go_out=paths=source_relative:. ./protos/datetime/datetime.proto

data:
	docker-compose kill database redis firestore
	docker-compose up -d database redis firestore
	bash -c "until actools mysql -h database -u dev-user -pdev-password -e ';' 2> /dev/null ; do sleep 1; done"

datetime-generator: _datetime-generator gofmt

_datetime-generator:
	wget http://www.unicode.org/Public/cldr/27.0.1/core.zip -O /tmp/core.zip
	go install ./cmd/datetime-generator
	datetime-generator -locales en,es,fr,ru,de,it,ja,pt
