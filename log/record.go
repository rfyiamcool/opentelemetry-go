// Copyright The OpenTelemetry Authors
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

package log // import "go.opentelemetry.io/otel/log"

import "time"

// attributesInlineCount is the number of attributes that are efficiently
// stored in an array within a Record. This value is borrowed from slog which
// performed a quantitative survey of log library use and found this value to
// cover 95% of all use-cases (https://go.dev/blog/slog#performance).
const attributesInlineCount = 5

// Record represents a log record.
type Record struct {
	timestamp         time.Time
	observedTimestamp time.Time
	severity          Severity
	severityText      string
	body              Value

	// The fields below are for optimizing the implementation of Attributes and
	// AddAttributes. This design is borrowed from the slog Record type:
	// https://cs.opensource.google/go/go/+/refs/tags/go1.22.0:src/log/slog/record.go;l=20

	// Allocation optimization: an inline array sized to hold
	// the majority of log calls (based on examination of open-source
	// code). It holds the start of the list of attributes.
	front [attributesInlineCount]KeyValue

	// The number of attributes in front.
	nFront int

	// The list of attributes except for those in front.
	// Invariants:
	//   - len(back) > 0 if nFront == len(front)
	//   - Unused array elements are zero-ed. Used to detect mistakes.
	back []KeyValue
}

// Timestamp returns the time when the log record occurred.
func (r *Record) Timestamp() time.Time {
	return r.timestamp
}

// SetTimestamp sets the time when the log record occurred.
func (r *Record) SetTimestamp(t time.Time) {
	r.timestamp = t
}

// ObservedTimestamp returns the time when the log record was observed.
func (r *Record) ObservedTimestamp() time.Time {
	return r.observedTimestamp
}

// SetObservedTimestamp sets the time when the log record was observed.
func (r *Record) SetObservedTimestamp(t time.Time) {
	r.observedTimestamp = t
}

// Severity returns the [Severity] of the log record.
func (r *Record) Severity() Severity {
	return r.severity
}

// SetSeverity sets the [Severity] level of the log record.
func (r *Record) SetSeverity(level Severity) {
	r.severity = level
}

// SeverityText returns severity (also known as log level) text. This is the
// original string representation of the severity as it is known at the source.
func (r *Record) SeverityText() string {
	return r.severityText
}

// SetSeverityText sets severity (also known as log level) text. This is the
// original string representation of the severity as it is known at the source.
func (r *Record) SetSeverityText(text string) {
	r.severityText = text
}

// Body returns the body of the log record.
func (r *Record) Body() Value {
	return r.body
}

// SetBody sets the body of the log record.
func (r *Record) SetBody(v Value) {
	r.body = v
}

// WalkAttributes walks all attributes the log record holds by calling f for
// each on each [KeyValue] in the [Record]. Iteration stops if f returns false.
func (r *Record) WalkAttributes(f func(KeyValue) bool) {
	for i := 0; i < r.nFront; i++ {
		if !f(r.front[i]) {
			return
		}
	}
	for _, a := range r.back {
		if !f(a) {
			return
		}
	}
}

// AddAttributes adds attributes to the log record.
func (r *Record) AddAttributes(attrs ...KeyValue) {
	var i int
	for i = 0; i < len(attrs) && r.nFront < len(r.front); i++ {
		a := attrs[i]
		r.front[r.nFront] = a
		r.nFront++
	}

	// TODO: when Go 1.20 is no longer supported, use slices.Grow instead.
	r.back = grow(r.back, len(attrs[i:]))
	r.back = append(r.back, attrs[i:]...)
}

// grow increases the slice's capacity, if necessary, to guarantee space for
// another n elements. After grow(n), at least n elements can be appended to
// the slice without another allocation.
//
// This is based on [Grow]. It is not available in Go 1.20 so it is reproduced
// here.
//
// [Grow]: https://pkg.go.dev/slices#Grow
func grow(slice []KeyValue, n int) []KeyValue {
	if n -= cap(slice) - len(slice); n > 0 {
		slice = append(slice[:cap(slice)], make([]KeyValue, n)...)[:len(slice)]
	}
	return slice
}

// AttributesLen returns the number of attributes in the log record.
func (r *Record) AttributesLen() int {
	return r.nFront + len(r.back)
}
