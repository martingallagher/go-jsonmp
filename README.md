# go-jsonmp
A Go package to facilitate JSON merge patch format and processing rules ([RFC 7386](https://tools.ietf.org/html/rfc7386)).

## Installation
    go get -u github.com/martingallagher/go-jsonmp


# Usage

### Bytes
```go
a := []byte(`{"a":{"b": "c"}}`)
b := []byte(`{"a":{"b":"d","c":null}}`)

res, err := jsonmp.Patch(a, b)

fmt.Println(string(res)) // {"a":{"b":"d"}}
```

### Values
```go
type doc struct {
	Title string `json:"title,omitempty"`
	Body  string `json:"body,omitempty"`
}

a := &doc{Title: "Hello"}
b := &doc{Body: "Dennis"}

var res *doc

jsonmp.PatchValue(a, b, &res) // *doc{Title:"Hello", Body:"Dennis"}
```

### Patcher
```go
// w http.ResponseWriter, r *http.Request
p := jsonmp.NewPatcher(r.Body, w)
d := &Document{...}

err := p.PatchValue(d)

// result written to w / http.ResponseWriter
```

# Contributions
Bug fixes and feature requests welcome.

# Contributors
- [Martin Gallagher](http://martingallagher.com/)