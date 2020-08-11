package tracker

import (
	"fmt"
)

type Predicate func(*Item) bool

var (
	PredicateAll Predicate = func(item *Item) bool { return true }
)

type DB interface {
	Put(key string, item *Item) error
	Get(key string) (*Item, error)
	Query(predicate Predicate) ([]*Item, error)
}

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

func (db *SimpleFileDB) Query(predicate Predicate) ([]*Item, error) {
	var items []*Item
	for _, item := range db.table {
		if predicate(item) {
			items = append(items, item)
		}
	}
	return items, nil
}

func (db *SimpleFileDB) Flush() error {
	items, err := db.Query(PredicateAll)
	if err != nil {
		return err
	}
	return encodeItemsToFile(db.path, items)
}
