# go-structdiff

[![CI](https://github.com/philiprehberger/go-structdiff/actions/workflows/ci.yml/badge.svg)](https://github.com/philiprehberger/go-structdiff/actions/workflows/ci.yml) [![Go Reference](https://pkg.go.dev/badge/github.com/philiprehberger/go-structdiff.svg)](https://pkg.go.dev/github.com/philiprehberger/go-structdiff) [![License](https://img.shields.io/github/license/philiprehberger/go-structdiff)](LICENSE)

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

### Only compare specific fields

```go
changes := structdiff.Compare(a, b, structdiff.OnlyFields("Name", "Age"))
// Only compares Name and Age, ignores all other fields
```

### Patching

```go
p := Person{Name: "Alice", Age: 30}
changes := structdiff.Compare(p, Person{Name: "Bob", Age: 31})

err := structdiff.Patch(&p, changes)
// p is now {Name: "Bob", Age: 31}
```

Patch supports nested struct fields via dot-notation paths:

```go
p := Person{Name: "Alice", Address: Address{City: "NY"}}
err := structdiff.Patch(&p, []structdiff.Change{
    {Path: "Address.City", New: "LA"},
})
// p.Address.City is now "LA"
```

### Formatted output

```go
changes := structdiff.Compare(a, b)

// Human-readable
fmt.Println(structdiff.Format(changes))
// Name: "Alice" → "Bob"
// Age: 30 → 31

// JSON
data, err := structdiff.FormatJSON(changes)
// [{"path":"Name","old":"Alice","new":"Bob"},{"path":"Age","old":30,"new":31}]
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
| `Patch(target any, changes []Change) error` | Apply changes to a struct pointer by setting fields at each path |
| `Format(changes []Change) string` | Human-readable multi-line diff output with quoted strings |
| `FormatJSON(changes []Change) ([]byte, error)` | JSON array of `{path, old, new}` objects |
| `Ignore(fields ...string) Option` | Skip fields by name |
| `IgnoreTag(tag string) Option` | Skip fields with a specific struct tag value (e.g., `diff:"-"`) |
| `OnlyFields(fields ...string) Option` | Restrict comparison to only the specified fields |
| `Change{Path, Old, New}` | Represents a single field difference |
| `Change.String() string` | Human-readable format: `"Path: old → new"` |

## Development

```bash
go test ./...
go vet ./...
```

## License

MIT
