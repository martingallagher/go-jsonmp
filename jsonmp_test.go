package jsonmp

import (
	"bytes"
	"encoding/json"
	"reflect"
	"strings"
	"testing"
	"time"
)

var testData = []struct{ a, b, result string }{
	// RFC 7386 Example Test Cases
	{`{"a":"b"}`, `{"a":"c"}`, `{"a":"c"}`},
	1: {`{"a":"b"}`, `{"b":"c"}`, `{"a":"b","b":"c"}`},
	{`{"a":"b"}`, `{"a":null}`, `{}`},
	{`{"a":"b","b":"c"}`, `{"a":null}`, `{"b":"c"}`},
	{`{"a":["b"]}`, `{"a":"c"}`, `{"a":"c"}`},
	5: {`{"a":"c"}`, `{"a":["b"]}`, `{"a":["b"]}`},
	{`{"a":{"b": "c"}}`, `{"a":{"b":"d","c":null}}`, `{"a":{"b":"d"}}`},
	{`{"a":[{"b":"c"}]}`, `{"a":[1]}`, `{"a":[1]}`},
	{`["a","b"]`, `["c","d"]`, `["c","d"]`},
	{`{"a":"b"}`, `["c"]`, `["c"]`},
	10: {`{"a":"foo"}`, `null`, `null`},
	{`{"a":"foo"}`, `"bar"`, `"bar"`},
	{`{"e":null}`, `{"a":1}`, `{"e":null,"a":1}`},
	{`[1,2]`, `{"a":"b","c":null}`, `{"a":"b"}`},
	{`{}`, `{"a":{"bb":{"ccc":null}}}`, `{"a":{"bb":{}}}`},

	// RFC 7386 Example Document
	{`{
    "title": "Goodbye!",
    "author": {
        "givenName": "John",
        "familyName": "Doe"
    },
    "tags": ["example", "sample"],
    "content": "This will be unchanged"
}`, `{
    "title": "Hello!",
    "phoneNumber": "+01-123-456-7890",
    "author": {
        "familyName": null
    },
    "tags": ["example"]
}`, `{
    "title": "Hello!",
    "author": {
        "givenName": "John"
    },
    "tags": ["example"],
    "content": "This will be unchanged",
    "phoneNumber": "+01-123-456-7890"
}`},
}

// fmtResult encodes the result JSON in accordance with
// encoding/json for field-ordering reasons
func fmtResult(s string) string {
	var i interface{}

	json.Unmarshal([]byte(s), &i)

	b, _ := json.Marshal(i)

	return string(b)
}

func TestPatch(t *testing.T) {
	for i, c := range testData {
		p, err := Patch([]byte(c.a), []byte(c.b))

		if err != nil {
			t.Fatal(err)
		}

		if r := fmtResult(c.result); r != string(p) {
			t.Fatalf("incorrect result (%d): %s != %s", i, p, r)
		}
	}
}

type document struct {
	Title     string    `json:"title,omitempty"`
	Body      string    `json:"body,omitempty"`
	Author    string    `json:"author,omitempty"`
	Score     int       `json:"score,omitempty"`
	Published time.Time `json:"published,omitempty"`
}

func TestPatchValue(t *testing.T) {
	d := &document{
		Title: "On the Origin of Species",
		Body:  "...",
	}

	d.Published, _ = time.Parse("02 Jan 2005", "24 Nov 1859")
	body := "From the strong principle of inheritance, any selected variety will tend to propagate its new and modified form."

	p := &document{
		Body:      body,
		Author:    "Charles Darwin",
		Score:     10,
		Published: time.Time{},
	}

	var result *document

	if err := PatchValue(d, p, &result); err != nil {
		t.Fatal(err)
	}

	if !result.Published.IsZero() {
		t.Fatal("invalid published value")
	} else if result.Body != body {
		t.Fatal("invalid body value")
	} else if result.Score != 10 {
		t.Fatal("invalid score value")
	}
}

func TestPatcher(t *testing.T) {
	buf := &bytes.Buffer{}

	for i, c := range testData {
		buf.Reset()

		r := strings.NewReader(c.b)
		p := NewPatcher(r, buf)

		if err := p.Patch([]byte(c.a)); err != nil {
			t.Fatal(err)
		}

		var a interface{}

		if err := json.NewDecoder(buf).Decode(&a); err != nil {
			t.Fatal(err)
		}

		var b interface{}

		if err := json.Unmarshal([]byte(c.result), &b); err != nil {
			t.Fatal(err)
		}

		if !reflect.DeepEqual(a, b) {
			t.Fatal("incorrect result for test data", i)
		}
	}

	for i, c := range testData {
		buf.Reset()

		r := strings.NewReader(c.b)
		p := NewPatcher(r, buf)

		var a interface{}

		if err := json.Unmarshal([]byte(c.a), &a); err != nil {
			t.Fatal(err)
		}

		if err := p.PatchValue(a); err != nil {
			t.Fatal(err)
		}

		var b interface{}

		if err := json.NewDecoder(buf).Decode(&b); err != nil {
			t.Fatal(err)
		}

		var res interface{}

		if err := json.Unmarshal([]byte(c.result), &res); err != nil {
			t.Fatal(err)
		}

		if !reflect.DeepEqual(res, b) {
			t.Fatal("incorrect result for test data", i)
		}
	}
}

func BenchmarkPatch(b *testing.B) {
	b.ReportAllocs()

	x, y := []byte(testData[15].a), []byte(testData[15].b)

	for i := 0; i < b.N; i++ {
		Patch(x, y)
	}
}

func BenchmarkPatcher(b *testing.B) {
	b.ReportAllocs()

	buf := &bytes.Buffer{}

	for i := 0; i < b.N; i++ {
		buf.Reset()

		r := strings.NewReader(testData[15].b)
		p := NewPatcher(r, buf)

		p.Patch([]byte(testData[15].a))
	}
}
