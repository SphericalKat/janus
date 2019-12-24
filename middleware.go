package janus

import (
	"github.com/allegro/bigcache"
	"github.com/jinzhu/gorm"
	"net/http"
	"time"
)

// Janus represents an instance of the Janus middleware
type Janus struct {
	db    *gorm.DB
	cache *bigcache.BigCache
}

// NewJanusMiddleware return a new instance of the Janus middleware
func NewJanusMiddleware(db *gorm.DB) (*Janus, error) {
	db.AutoMigrate(&Account{})

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

func (j *Janus) saveAccount(account *Account) {
	j.db.Save(account)
}

func JanusHandler(next http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

	}
}

