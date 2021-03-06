package tracker

import (
	"errors"
	"fmt"
	"sort"
	"time"
)

// Predicate is a condition / matcher function for a tracker item.
type Predicate func(*Item) bool

// List of prebuilt predicates.
var (
	PredicateAll = func(item *Item) bool { return true }
	PredicateTag = func(tag string) Predicate {
		return func(item *Item) bool { return item.Tags.Has(tag) }
	}
)

var (
	errEmptyResult = errors.New("no results found")
)

type SimpleFileDB struct {
	path  string
	table map[string]*Item
}

func NewSimpleFileDB(path string) (*SimpleFileDB, error) {
	items, err := decodeItemsFromFile(path)
	if err != nil {
		return nil, fmt.Errorf("decoding %s: %v", path, err)
	}

	db := &SimpleFileDB{
		path:  path,
		table: make(map[string]*Item),
	}

	for _, item := range items {
		err := db.Put(item.ID, item)
		if err != nil {
			return nil, err
		}
	}

	return db, nil
}

func (db *SimpleFileDB) Exists(key string) bool {
	_, ok := db.table[key]
	return ok
}

func (db *SimpleFileDB) Put(key string, item *Item) error {
	db.table[key] = item
	return nil
}

func (db *SimpleFileDB) Get(key string) (*Item, error) {
	item, ok := db.table[key]
	if !ok {
		return nil, fmt.Errorf("%q not found", key)
	}
	return item, nil
}

func (db *SimpleFileDB) Query(predicates ...Predicate) ([]*Item, error) {
	// If predicates is empty, select all items.
	if len(predicates) == 0 {
		predicates = append(predicates, PredicateAll)
	}

	var items []*Item
	for _, item := range db.table {
		// Pass all predicates or else be excluded from the result.
		var exclude bool
		for _, p := range predicates {
			if !p(item) {
				exclude = true
				break
			}
		}
		if !exclude {
			items = append(items, item)
		}
	}
	return items, nil
}

func (db *SimpleFileDB) QueryLatest(predicates ...Predicate) (*Item, error) {
	items, err := db.Query(predicates...)
	if err != nil {
		return nil, err
	}

	if len(items) == 0 {
		return nil, errEmptyResult
	}

	sort.Slice(items, func(a, b int) bool {
		return time.Time(items[a].Created).Unix() > time.Time(items[b].Created).Unix()
	})

	return items[0], nil
}

func (db *SimpleFileDB) Flush() error {
	items, err := db.Query()
	if err != nil {
		return err
	}
	return encodeItemsToFile(db.path, items)
}
