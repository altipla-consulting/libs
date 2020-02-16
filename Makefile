
FILES = $(shell find . -type f -name '*.go' -not -path './vendor/*')

.PHONY: protos

gofmt:
	@gofmt -s -w $(FILES)
	@gofmt -r '&α{} -> new(α)' -w $(FILES)

test: lint
	actools go test ./...

lint:
	@./infra/lint-errors.sh
	tool/linter ./...

update-deps:
	actools go get -u
	actools go mod download
	actools go mod tidy

protos:
	actools protoc --go_out=paths=source_relative:. ./protos/datetime/datetime.proto

data:
	actools rm database redis
	actools start database redis
	bash -c "until actools mysql -h database -u dev-user -pdev-password -e ';' 2> /dev/null ; do sleep 1; done"

datetime-generator: _datetime-generator gofmt

_datetime-generator:
	wget http://www.unicode.org/Public/cldr/27.0.1/core.zip -O /tmp/core.zip
	go install ./cmd/datetime-generator
	datetime-generator -locales en,es,fr,ru,de,it,ja,pt
