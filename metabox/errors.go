package metabox

import "errors"

var (
	errCacheNotFound     = errors.New("cache not found")
	errNoAvailableStores = errors.New("no available stores")
)
