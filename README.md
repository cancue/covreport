# covreport

[![godoc - documentation](https://godoc.org/github.com/cancue/covreport?status.svg)](https://pkg.go.dev/github.com/cancue/covreport)
[![go report card](https://goreportcard.com/badge/github.com/cancue/covreport)](https://goreportcard.com/report/github.com/cancue/covreport)
[![github action - test](https://github.com/cancue/covreport/workflows/test/badge.svg)](https://github.com/cancue/covreport/actions)
[![codecov - code coverage](https://img.shields.io/codecov/c/github/cancue/covreport.svg?style=flat-square)](https://codecov.io/gh/cancue/covreport)

**covreport** is a html coverage reporter for go coverprofile.

## Installation
```shell
go install github.com/cancue/covreport@v0.3.0
```

### (optional) Generate profile
```shell
go test -coverprofile cover.prof ./...
```

## Example
```shell
# all flags are optional
# covreport && open cover.html

covreport -i cover.prof -o cover.html -cutlines 70,40
```

## Manual
```shell
covreport -h
```

## Screenshots
![screenshots](https://github.com/cancue/covreport/assets/8125241/47b8ceaa-042d-4e4f-b306-90c8b0a09fbe)


## License

[MIT](https://github.com/cancue/covreport/blob/master/LICENSE)
