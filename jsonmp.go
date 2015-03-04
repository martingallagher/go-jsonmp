// Copyright Praegressus Limited. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package jsonmp

import (
	"encoding/json"
	"io"
)

// Patch patches a JSON byte value in accordance with RFC 7386.
func Patch(a, b []byte) ([]byte, error) {
	// Optimization: combine and create a single JSON value
	// BenchmarkPatchOld  50000   27684 ns/op  3812 B/op  85 allocs/op
	// BenchmarkPatchNew  100000  12476 ns/op  4860 B/op  18 allocs/op
	a = append(a, 0)

	copy(a[1:], a[0:])

	a[0] = '['
	a = append(a, ',')
	a = append(a, b...)
	a = append(a, ']')

	var i []interface{}

	if err := json.Unmarshal(a, &i); err != nil {
		return nil, err
	}

	return json.Marshal(patch(i[0], i[1]))
}

// PatchValue patches the interface values,
// the destination interface recieves the result.
func PatchValue(a, b, dst interface{}) error {
	var err error

	if a, err = coerce(a); err != nil {
		return err
	}

	if b, err = coerce(b); err != nil {
		return err
	}

	return marshalValue(a, b, dst)
}

// PatchValueWithBytes patches an interface value
// with a byte value. The destination interface
// recieves the result.
func PatchValueWithBytes(a interface{}, c []byte, dst interface{}) error {
	var err error

	if a, err = coerce(a); err != nil {
		return err
	}

	var b interface{}

	if err = json.Unmarshal(c, &b); err != nil {
		return err
	}

	return marshalValue(a, b, dst)
}

// PatchValueWithBytes patches an interface value
// with an io.Reader. The destination interface
// recieves the result.
func PatchValueWithReader(a interface{}, r io.Reader, dst interface{}) error {
	var err error

	if a, err = coerce(a); err != nil {
		return err
	}

	var b interface{}

	if err = json.NewDecoder(r).Decode(&b); err != nil {
		return err
	}

	return marshalValue(a, b, dst)
}

// marshalValue is a helper function for patching
// and putting the result to a destination interface.
func marshalValue(a, b, dst interface{}) error {
	c, err := json.Marshal(patch(a, b))

	if err != nil {
		return err
	}

	return json.Unmarshal(c, &dst)
}

// Patcher reads the patch data from the Reader
// and writes the patched result to the Writer
// when Patch() is called.
type Patcher struct {
	r io.Reader
	w io.Writer
}

// NewPatcher returns a new Patcher.
func NewPatcher(r io.Reader, w io.Writer) *Patcher {
	return &Patcher{r, w}
}

// Patch patches the JSON byte value with
// the content of the Patcher's Reader.
// The result is then written to the Writer.
func (p *Patcher) Patch(c []byte) error {
	b, err := read(p.r)

	if err != nil {
		return err
	}

	var a interface{}

	if err := json.Unmarshal(c, &a); err != nil {
		return err
	}

	return p.write(a, b)
}

// PatchValue patches the interface with
// the content of the Patcher's Reader.
// The result is then written to the Writer.
func (p *Patcher) PatchValue(a interface{}) error {
	var err error

	if a, err = coerce(a); err != nil {
		return err
	}

	b, err := read(p.r)

	if err != nil {
		return err
	}

	return p.write(a, b)
}

func read(r io.Reader) (interface{}, error) {
	var i interface{}

	err := json.NewDecoder(r).Decode(&i)

	return i, err
}

func (p *Patcher) write(a, b interface{}) error {
	return json.NewEncoder(p.w).Encode(patch(a, b))
}

// coerces the original interface into a vanilla interface.
func coerce(i interface{}) (interface{}, error) {
	if i == nil {
		return i, nil
	}

	// Assume vanilla interface, exit early
	switch i.(type) {
	case map[string]interface{}:
		return i, nil

	case []interface{}:
		return i, nil
	}

	b, err := json.Marshal(i)

	if err != nil {
		return nil, err
	}

	var v interface{}

	return v, json.Unmarshal(b, &v)
}

// patches A with B and returns the result
func patch(a, b interface{}) interface{} {
	if m, ok := a.(map[string]interface{}); ok {
		return handleMap(m, b)
	}

	if m, ok := b.(map[string]interface{}); ok {
		return removeNull(m)
	}

	return b
}

// removeNull recursively removes null entries in a map.
func removeNull(m map[string]interface{}) map[string]interface{} {
	for k, v := range m {
		if v == nil {
			delete(m, k)

			continue
		}

		if c, ok := v.(map[string]interface{}); ok {
			m[k] = removeNull(c)
		}
	}

	return m
}

// handleMap handles patching of map values.
func handleMap(m map[string]interface{}, p interface{}) interface{} {
	c, ok := p.(map[string]interface{})

	// Patch type over-rules
	if !ok {
		return p
	}

	for k, v := range c {
		// New entry
		if _, exists := m[k]; !exists {
			if v == nil {
				continue
			}

			if n, ok := v.(map[string]interface{}); ok {
				m[k] = removeNull(n)
			} else {
				m[k] = v
			}

			continue
		}

		// Old entry, null value = remove
		if v == nil {
			delete(m, k)

			continue
		}

		if n, ok := v.(map[string]interface{}); ok {
			m[k] = patch(m[k], n)
		} else {
			m[k] = v
		}
	}

	return m
}
