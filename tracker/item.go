package tracker

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"strconv"
	"strings"
	"time"
)

const emptyTagsMarker = "-"

// Tags is a []string wrapper with custom csv encode/decode.
type Tags []string

func (t Tags) Has(tag string) bool {
	for _, s := range t {
		if tag == s {
			return true
		}
	}
	return false
}

func (t Tags) MarshalCSV() ([]byte, error) {
	// Workaround serialization if no tags are available.
	if len(t) == 0 {
		t = append(t, emptyTagsMarker)
	}

	var buf bytes.Buffer

	w := csv.NewWriter(&buf)
	if err := w.Write(t); err != nil {
		return nil, err
	}
	w.Flush()

	return []byte(strings.TrimSpace(buf.String())), w.Error()
}

func (t *Tags) UnmarshalCSV(data []byte) error {
	r := csv.NewReader(bytes.NewBuffer(data))
	tags, err := r.Read()
	if err != nil {
		return err
	}

	// If found an empty tags in the first element, ignore it.
	if tags[0] == emptyTagsMarker {
		tags = tags[1:]
	}

	*t = tags

	return nil
}

// Time is a time.Time wrapper with custom CSV encode/decode.
type Time time.Time

func (t Time) MarshalCSV() ([]byte, error) {
	s := fmt.Sprintf("%d", time.Time(t).Unix())
	return []byte(s), nil
}

func (t *Time) UnmarshalCSV(data []byte) error {
	ts, err := strconv.ParseInt(string(data), 10, 64)
	if err != nil {
		return err
	}
	*t = Time(time.Unix(ts, 0))
	return nil
}

type Item struct {
	ID      string `csv:"hash"`
	Created Time   `csv:"created_time"`
	Author  string `csv:"author"`
	Tags    Tags   `csv:"tags"`
}
