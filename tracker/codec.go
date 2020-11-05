package tracker

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"sort"
	"time"

	"github.com/jszwec/csvutil"
)

const separator = ' '

func decodeItemsFromFile(path string) ([]*Item, error) {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_RDONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("opening %s: %v", path, err)
	}

	reader := csv.NewReader(file)
	reader.Comma = separator

	// Disable parsing headers.
	header, err := csvutil.Header(&Item{}, "csv")
	if err != nil {
		return nil, fmt.Errorf("retrieving header: %v", err)
	}
	// Create CSV decoder.
	decoder, err := csvutil.NewDecoder(reader, header...)
	if err == io.EOF {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("csv decode %s: %v", path, err)
	}

	// Collect all decoded items.
	var items []*Item
	for {
		var item Item
		if err := decoder.Decode(&item); err == io.EOF {
			break
		} else if err != nil {
			return nil, fmt.Errorf("malformed entry #%d: %v", len(items)+1, err)
		}
		items = append(items, &item)
	}
	return items, nil
}

func encodeItemsToFile(path string, items []*Item) error {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("opening %s: %v", path, err)
	}

	// Sort the items first by ascending timestamp.
	sort.SliceStable(items, func(i, j int) bool {
		a := time.Time(items[i].Created)
		b := time.Time(items[j].Created)
		return a.Before(b)
	})

	writer := csv.NewWriter(file)
	writer.Comma = separator

	encoder := csvutil.NewEncoder(writer)
	encoder.AutoHeader = false
	for _, item := range items {
		if err := encoder.Encode(item); err != nil {
			return fmt.Errorf("encoding item: %v", err)
		}
	}

	writer.Flush()

	return writer.Error()
}
