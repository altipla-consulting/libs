
# libs

[![GoDoc](https://godoc.org/libs.altipla.consulting?status.svg)](https://godoc.org/libs.altipla.consulting)

List of Go utils and shared code for [Altipla Consulting](https://www.altiplaconsulting.com/) projects.

**WARNING:** We do incompatible releases from minor to minor release because it is an internal lib. All backwards incompatible changes will be listed on [CHANGELOG.md](CHANGELOG.md)


### Install

```shell
go get libs.altipla.consulting
```


### Packages

- [arrays](arrays)
- [bigquery](bigquery)
- [collections](collections)
- [connect](connect)
- [content](content)
- [database](database)
- [datetime](datetime)
- [encoding](encoding)
- [errors](errors)
- [frontmatter](frontmatter)
- [geo](geo)
- [grpctest](grpctest)
- [langs](langs)
- [loaders](loaders)
- [mailgun](mailgun)
- [messageformat](messageformat)
- [mjml](mjml)
- [money](money)
- [pagination](pagination)
- [recaptcha](recaptcha)
- [redis](redis)
- [routing](routing)
- [schema](schema)
- [sentry](sentry)
- [services](services)
- [slack](slack)
- [storage](storage)
- [templates](templates)
- [tokensource](tokensource)
- [validation](validation)


### Contributing

You can make pull requests or create issues in GitHub. Any code you send should be formatted using `make gofmt`.


### Running tests

Run the tests:

```shell
make test
```


### License

[MIT License](LICENSE)
