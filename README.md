# Close Pool

[![GoDoc](https://pkg.go.dev/badge/github.com/bassosimone/closepool)](https://pkg.go.dev/github.com/bassosimone/closepool) [![Build Status](https://github.com/bassosimone/closepool/actions/workflows/go.yml/badge.svg)](https://github.com/bassosimone/closepool/actions) [![codecov](https://codecov.io/gh/bassosimone/closepool/branch/main/graph/badge.svg)](https://codecov.io/gh/bassosimone/closepool)

The `closepool` Go package allows pooling `io.Closer` instances
and closing them in a single operation, iterating in reverse (LIFO)
order. This is useful in loops that create layered resources (e.g.,
a TCP connection followed by a TLS connection on top of it).

For example:

```Go
import "github.com/bassosimone/closepool"

// Create a pool and defer closing all resources.
pool := &closepool.Pool{}
defer pool.Close()

// Resources are closed in reverse (LIFO) order,
// so the TLS connection is closed before the TCP one.
pool.Add(tcpConn)
pool.Add(tlsConn)
```

## Installation

To add this package as a dependency to your module:

```sh
go get github.com/bassosimone/closepool
```

## Development

To run the tests:
```sh
go test -v .
```

To measure test coverage:
```sh
go test -v -cover .
```

## License

```
SPDX-License-Identifier: GPL-3.0-or-later
```

## History

Adapted from [rbmk-project/rbmk](https://github.com/rbmk-project/rbmk/tree/v0.18.0/pkg/common/closepool/).
