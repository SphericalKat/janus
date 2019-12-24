package janus

import (
	"github.com/allegro/bigcache"
	"github.com/jinzhu/gorm"
	"time"
)

// Janus represents an instance of the Janus middleware
type Janus struct {
	db    *gorm.DB
	cache *bigcache.BigCache
}

func NewJanusMiddleware(db *gorm.DB) (*Janus, error) {
	config := bigcache.Config{
		Shards:             1024,
		LifeWindow:         30 * 24 * time.Hour,
		CleanWindow:        5 * time.Minute,
		MaxEntriesInWindow: 1000 * 10 * 60,
		MaxEntrySize:       500,
		HardMaxCacheSize:   512,
		OnRemove:           nil,
		OnRemoveWithReason: nil,
	}

	cache, err := bigcache.NewBigCache(config)
	if err != nil {
		return nil, err
	}

	return &Janus{
		db:    db,
		cache: cache,
	}, nil
}
