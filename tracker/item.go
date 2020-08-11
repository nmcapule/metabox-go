package tracker

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// Tags is a []string wrapper with custom csv encode/decode.
type Tags []string

func (t Tags) MarshalCSV() ([]byte, error) {
	var buf bytes.Buffer

	w := csv.NewWriter(&buf)
	if err := w.Write(t); err != nil {
		return nil, err
	}
	w.Flush()

	return []byte(strings.TrimSpace(buf.String())), w.Error()
}

func (t *Tags) UnmarshalCSV(data []byte) (err error) {
	r := csv.NewReader(bytes.NewBuffer(data))
	*t, err = r.Read()
	return
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
