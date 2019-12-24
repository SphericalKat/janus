package janus

import (
	"context"
	"encoding/json"
	"github.com/allegro/bigcache"
	"github.com/jinzhu/gorm"
	"github.com/wI2L/jettison"
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

func (j *Janus) GetHandler(next http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		acc := ctx.Value("janus_context").(*Account)

		jsun, err := j.cache.Get(acc.Key)
		if err != nil {
			j.db.Where("key = ?", acc.Key).Find(acc)
			jsun, _ := jettison.Marshal(acc)
			_ = j.cache.Set(acc.Key, jsun)
		}

		_ = json.Unmarshal(jsun, acc)

		ctx = context.WithValue(ctx, "janus_context", acc)
		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)
	}
}
