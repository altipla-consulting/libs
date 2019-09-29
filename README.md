
# libs

[![GoDoc](https://godoc.org/libs.altipla.consulting?status.svg)](https://godoc.org/libs.altipla.consulting)

List of Go utils and shared code for [Altipla Consulting](https://www.altiplaconsulting.com/) projects.

**WARNING:** We do incompatible releases from minor to minor release because it is an internal lib. All backwards incompatible changes will be listed on [CHANGELOG.md](CHANGELOG.md)


### Install

```shell
go get libs.altipla.consulting
```


### Packages

- [arrays](arrays/README.md)
- [bigquery](bigquery/README.md)
- [collections](collections/README.md)
- [connect](connect/README.md)
- [content](content/README.md)
- [crypt](crypt/README.md)
- [database](database/README.md)
- [datetime](datetime/README.md)
- [encoding](encoding/README.md)
- [errors](errors/README.md)
- [frontmatter](frontmatter/README.md)
- [geo](geo/README.md)
- [grpctest](grpctest/README.md)
- [langs](langs/README.md)
- [loaders](loaders/README.md)
- [mailgun](mailgun/README.md)
- [messageformat](messageformat/README.md)
- [mjml](mjml/README.md)
- [money](money/README.md)
- [pagination](pagination/README.md)
- [recaptcha](recaptcha/README.md)
- [redis](redis/README.md)
- [routing](routing/README.md)
- [schema](schema/README.md)
- [sentry](sentry/README.md)
- [services](services/README.md)
- [slack](slack/README.md)
- [storage](storage/README.md)
- [templates](templates/README.md)
- [tokensource](tokensource/README.md)
- [validation](validation/README.md)


### Contributing

You can make pull requests or create issues in GitHub. Any code you send should be formatted using `make gofmt`.


### Running tests

Run the tests:

```shell
make test
```


### License

[MIT License](LICENSE)
