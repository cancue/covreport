# covreport

[![godoc - documentation](https://godoc.org/github.com/cancue/covreport?status.svg)](https://pkg.go.dev/github.com/cancue/covreport)
[![go report card](https://goreportcard.com/badge/github.com/cancue/covreport)](https://goreportcard.com/report/github.com/cancue/covreport)
[![github action - test](https://github.com/cancue/covreport/workflows/test/badge.svg)](https://github.com/cancue/covreport/actions)

**covreport** is a html coverage reporter for go coverprofile.

## Installation
```go
go install github.com/cancue/covreport
```

## Generate profile (optional)
```shell
go test -coverprofile cover.prof ./...
```

## Example
```shell
covreport -i cover.prof -o cover.html
open cover.html
```

## Screenshots
![screenshots](https://github.com/cancue/covreport/assets/8125241/47b8ceaa-042d-4e4f-b306-90c8b0a09fbe)


## License

[MIT](https://github.com/cancue/covreport/blob/master/LICENSE)
