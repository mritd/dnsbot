package main

import (
	"time"

	"github.com/allegro/bigcache"
)

var cache *bigcache.BigCache

func init() {
	cache, _ = bigcache.NewBigCache(bigcache.Config{
		Shards:      1024,
		LifeWindow:  3 * time.Minute,
		CleanWindow: 5 * time.Minute,
	})
}
