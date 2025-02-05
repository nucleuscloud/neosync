package loki

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"time"
)

/*
These types were pulled from the github.com/grafana/loki/pkg/logcli/client/client.go file
The ResultType stuff can be pulled from github.com/grafana/loki/pkg/loghttp/query.go

We aren't using this package directly because it installs conflicting kube client types
and doesn't work with our current go.mod
*/

// Define a struct that matches the structure of the JSON response you expect from Loki.
// This is a simplified example; adjust according to the actual response structure you need.
// QueryResponse represents the http json response to a label query
type QueryResponse struct {
	Status string            `json:"status"`
	Data   QueryResponseData `json:"data"`
}

// QueryResponseData represents the http json response to a label query
type QueryResponseData struct {
	ResultType ResultType  `json:"resultType"`
	Result     ResultValue `json:"result"`
	// Statistics stats.Result `json:"stats"`
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (q *QueryResponseData) UnmarshalJSON(data []byte) error {
	unmarshal := struct {
		Type   ResultType      `json:"resultType"`
		Result json.RawMessage `json:"result"`
		// Statistics stats.Result    `json:"stats"`
	}{}

	err := json.Unmarshal(data, &unmarshal)
	if err != nil {
		return err
	}

	var value ResultValue

	// unmarshal results
	switch unmarshal.Type {
	case ResultTypeStream:
		var s Streams
		err = json.Unmarshal(unmarshal.Result, &s)
		value = s
	// case ResultTypeMatrix:
	// 	var m Matrix
	// 	err = json.Unmarshal(unmarshal.Result, &m)
	// 	value = m
	// case ResultTypeVector:
	// 	var v Vector
	// 	err = json.Unmarshal(unmarshal.Result, &v)
	// 	value = v
	// case ResultTypeScalar:
	// 	var v Scalar
	// 	err = json.Unmarshal(unmarshal.Result, &v)
	// 	value = v
	default:
		return fmt.Errorf("unknown type: %s", unmarshal.Type)
	}

	if err != nil {
		return err
	}

	q.ResultType = unmarshal.Type
	q.Result = value
	//nolint
	// q.Statistics = unmarshal.Statistics

	return nil
}

// ResultValue interface mimics the promql.Value interface
type ResultValue interface {
	Type() ResultType
}

// ResultType holds the type of the result
type ResultType string

// ResultType values
const (
	ResultTypeStream = "streams"
	ResultTypeScalar = "scalar"
	ResultTypeVector = "vector"
	ResultTypeMatrix = "matrix"
)

// Type implements the promql.Value interface

// Streams is a slice of Stream
type Streams []Stream

func (Streams) Type() ResultType { return ResultTypeStream }

func (s Streams) ToProto() []Stream {
	if len(s) == 0 {
		return nil
	}
	result := make([]Stream, 0, len(s))
	for _, s := range s {
		result = append(result, Stream{Labels: s.Labels, Entries: s.Entries})
	}
	return result
}

// Stream represents a log stream.  It includes a set of log entries and their labels.
type Stream struct {
	Labels  LabelSet `json:"stream"`
	Entries []Entry  `json:"values"`
}

// LabelSet is a key/value pair mapping of labels
type LabelSet map[string]string

// Map coerces LabelSet into a map[string]string. This is useful for working with adapter types.
func (l LabelSet) Map() map[string]string {
	return l
}

// String implements the Stringer interface.  It returns a formatted/sorted set of label key/value pairs.
func (l LabelSet) String() string {
	var b bytes.Buffer

	keys := make([]string, 0, len(l))
	for k := range l {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	b.WriteByte('{')
	for i, k := range keys {
		if i > 0 {
			b.WriteByte(',')
			b.WriteByte(' ')
		}
		b.WriteString(k)
		b.WriteByte('=')
		b.WriteString(strconv.Quote(l[k]))
	}
	b.WriteByte('}')
	return b.String()
}

// Entry represents a log entry.  It includes a log message and the time it occurred at.
type Entry struct {
	Timestamp time.Time
	Line      string
}

type LabeledEntry struct {
	Entry
	Labels LabelSet
}

func (e *Entry) MarshalJSON() ([]byte, error) {
	l, err := json.Marshal(e.Line)
	if err != nil {
		return nil, err
	}
	return []byte(fmt.Sprintf("[\"%d\",%s]", e.Timestamp.UnixNano(), l)), nil
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (e *Entry) UnmarshalJSON(data []byte) error {
	var unmarshal []string

	err := json.Unmarshal(data, &unmarshal)
	if err != nil {
		return err
	}

	t, err := strconv.ParseInt(unmarshal[0], 10, 64)
	if err != nil {
		return err
	}

	e.Timestamp = time.Unix(0, t)
	e.Line = unmarshal[1]

	return nil
}
