
FILES = $(shell find . -type f -name '*.go' -not -path './vendor/*')

.PHONY: protos

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

protos:
	actools protoc --go_out=paths=source_relative:. ./protos/datetime/datetime.proto
	actools protoc --go_out=paths=source_relative:. ./protos/pagination/pagination.proto

data:
	actools rm database redis
	actools start database redis
	bash -c "until actools mysql -h database -u dev-user -pdev-password -e ';' 2> /dev/null ; do sleep 1; done"

datetime-generator:
	go install ./cmd/datetime-generator
	wget http://www.unicode.org/Public/cldr/27.0.1/core.zip -O /tmp/core.zip
	datetime-generator -locales en,es,fr,ru,de,it,ja,pt
	@gofmt -w $(FILES)
	@gofmt -r '&α{} -> new(α)' -w $(FILES)
