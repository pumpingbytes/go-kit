# dberror/postgres

Postgres/pgx implementation of `github.com/pumpingbytes/go-kit/dberror.Classifier`.

## Requirements

- Go `1.25+`

## Installation

```bash
go get github.com/pumpingbytes/go-kit/dberror/postgres
```

## Usage

```go
import (
    rootdberror "github.com/pumpingbytes/go-kit/dberror"
    postgresdberror "github.com/pumpingbytes/go-kit/dberror/postgres"
)

func newClassifier() rootdberror.Classifier {
    return postgresdberror.New()
}
```

This package is intentionally a separate module so the main `github.com/pumpingbytes/go-kit` module can stay on a lower Go version while the pgx-based classifier tracks newer pgx releases independently.

