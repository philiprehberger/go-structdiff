# go-structdiff

[![CI](https://github.com/philiprehberger/go-structdiff/actions/workflows/ci.yml/badge.svg)](https://github.com/philiprehberger/go-structdiff/actions/workflows/ci.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/philiprehberger/go-structdiff.svg)](https://pkg.go.dev/github.com/philiprehberger/go-structdiff)
[![License](https://img.shields.io/github/license/philiprehberger/go-structdiff)](LICENSE)

Field-level struct comparison for Go with dot-notation change paths and zero dependencies

## Installation

```bash
go get github.com/philiprehberger/go-structdiff
```

## Usage

### Compare two structs

```go
package main

import (
    "fmt"
    "github.com/philiprehberger/go-structdiff"
)

type Address struct {
    Street string
    City   string
}

type Person struct {
    Name    string
    Age     int
    Address Address
}

func main() {
    a := Person{Name: "Alice", Age: 30, Address: Address{Street: "Main St", City: "NY"}}
    b := Person{Name: "Alice", Age: 31, Address: Address{Street: "Main St", City: "LA"}}

    changes := structdiff.Compare(a, b)
    for _, c := range changes {
        fmt.Println(c)
    }
    // Output:
    // Age: 30 → 31
    // Address.City: NY → LA
}
```

### Ignore fields

```go
changes := structdiff.Compare(a, b, structdiff.Ignore("Age"))
// Only returns Address.City change
```

### Ignore by struct tag

```go
type Config struct {
    Name     string
    Internal string `diff:"-"`
}

a := Config{Name: "v1", Internal: "secret1"}
b := Config{Name: "v2", Internal: "secret2"}

changes := structdiff.Compare(a, b, structdiff.IgnoreTag("diff"))
// Only returns Name change; Internal is skipped
```

### Check equality

```go
if structdiff.Equal(a, b) {
    fmt.Println("structs are identical")
}
```

## API

| Function / Type | Description |
|---|---|
| `Compare(a, b any, opts ...Option) []Change` | Deep compare two structs, returns list of field-level changes |
| `Equal(a, b any, opts ...Option) bool` | Returns true if structs are deeply equal |
| `Ignore(fields ...string) Option` | Skip fields by name |
| `IgnoreTag(tag string) Option` | Skip fields with a specific struct tag value (e.g., `diff:"-"`) |
| `Change{Path, Old, New}` | Represents a single field difference |
| `Change.String() string` | Human-readable format: `"Path: old → new"` |

## Features

- All primitive types, strings, slices, maps, nested structs, pointers
- Dot-notation paths: `"Address.City"`, `"Items[2].Name"`
- Handles nil pointers and type mismatches
- Skips unexported fields automatically
- Zero external dependencies

## Development

```bash
go test ./...
go vet ./...
```

## License

MIT
