package jsonmp

import (
	"bytes"
	"encoding/json"
	"reflect"
	"strings"
	"testing"
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

type testStruct struct {
	Title       string   `json:"title,omitempty"`
	PhoneNumber string   `json:"phoneNumber,omitempty"`
	Content     string   `json:"content,omitempty"`
	Tags        []string `json:"tags,omitempty"`
	Author      *author  `json:"author,omitempty"`
}

type author struct {
	GivenName  string `json:"givenName,omitempty"`
	FamilyName string `json:"familyName,omitempty"`
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

func TestPatchValue(t *testing.T) {
	var a, b, r interface{}

	if err := json.Unmarshal([]byte(testData[15].a), &a); err != nil {
		t.Fatal(err)
	}

	if err := json.Unmarshal([]byte(testData[15].b), &b); err != nil {
		t.Fatal(err)
	}

	if err := json.Unmarshal([]byte(testData[15].result), &r); err != nil {
		t.Fatal(err)
	}

	// Patched result
	var p interface{}

	if err := PatchValue(a, b, &p); err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(p, r) {
		t.Fatal("invalid result")
	}
}

func TestPatchValueWithBytes(t *testing.T) {
	for i, c := range testData {
		var a interface{}

		if err := json.Unmarshal([]byte(c.a), &a); err != nil {
			t.Fatal(err)
		}

		var r interface{}

		if err := PatchValueWithBytes(a, []byte(c.b), &r); err != nil {
			t.Fatal(err)
		}

		var x interface{}

		if err := json.Unmarshal([]byte(c.result), &x); err != nil {
			t.Fatal(err)
		}

		if !reflect.DeepEqual(r, x) {
			t.Fatal("incorrect result for test data", i)
		}
	}
}

func TestPatchValueWithBytesTyped(t *testing.T) {
	var a, p, r *testStruct

	if err := json.Unmarshal([]byte(testData[15].a), &a); err != nil {
		t.Fatal(err)
	}

	if err := json.Unmarshal([]byte(testData[15].result), &r); err != nil {
		t.Fatal(err)
	}

	// Patched result
	b := []byte(testData[15].b)

	if err := PatchValueWithBytes(a, b, &p); err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(p, r) {
		t.Fatal("invalid result")
	}
}

func TestPatchValueWithReader(t *testing.T) {
	for i, c := range testData {
		var a interface{}

		if err := json.Unmarshal([]byte(c.a), &a); err != nil {
			t.Fatal(err)
		}

		var (
			x interface{}
			r = bytes.NewReader([]byte(c.b))
		)

		if err := PatchValueWithReader(a, r, &x); err != nil {
			t.Fatal(err)
		}

		var y interface{}

		if err := json.Unmarshal([]byte(c.result), &y); err != nil {
			t.Fatal(err)
		}

		if !reflect.DeepEqual(x, y) {
			t.Fatal("incorrect result for test data", i)
		}
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

func BenchmarkPatchValue(b *testing.B) {
	b.ReportAllocs()

	var x, y, r interface{}

	json.Unmarshal([]byte(testData[15].a), &x)
	json.Unmarshal([]byte(testData[15].b), &y)

	for i := 0; i < b.N; i++ {
		PatchValue(x, y, &r)
	}
}

func BenchmarkPatchValueWithBytes(b *testing.B) {
	b.ReportAllocs()

	var (
		v, r interface{}
		p    = []byte(testData[15].a)
	)

	json.Unmarshal([]byte(testData[15].a), &v)

	for i := 0; i < b.N; i++ {
		PatchValueWithBytes(v, p, &r)
	}
}

func BenchmarkPatchValueWithBytesTyped(b *testing.B) {
	b.ReportAllocs()

	var (
		x, r *testStruct
		p    = []byte(testData[15].b)
	)

	json.Unmarshal([]byte(testData[15].a), &x)

	for i := 0; i < b.N; i++ {
		PatchValueWithBytes(x, p, &r)
	}
}

func BenchmarkPatchValueWithReader(b *testing.B) {
	b.ReportAllocs()

	var (
		x, y interface{}
		r    = bytes.NewReader([]byte(testData[15].a))
	)

	for i := 0; i < b.N; i++ {
		PatchValueWithReader(x, r, &y)
		r.Seek(0, 0)
	}
}

func BenchmarkPatcher(b *testing.B) {
	b.ReportAllocs()

	// Preallocate for better statistics for Patcher benchmark
	buf := &bytes.Buffer{}
	r := bytes.NewReader([]byte(testData[15].b))
	x := []byte(testData[15].a)

	for i := 0; i < b.N; i++ {
		p := NewPatcher(r, buf)

		p.Patch(x)
		r.Seek(0, 0)
		buf.Reset()
	}
}
