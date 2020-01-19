package janus

import (
	"context"
	"encoding/json"
	"fmt"
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

// NewJanusMiddleware returns a new instance of the Janus middleware
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

// GetHandler get an instance of the middleware handler
func (j *Janus) GetHandler(next http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		acc := ctx.Value("janus_context").(*Account)

		// try to find account in cache
		jsun, err := j.cache.Get(fmt.Sprintf("%v-%v", acc.CacheKey, acc.OrganizationID))
		if err != nil { // cache miss
			err = j.db.Where("cache_key = ?", acc.CacheKey).Where("organization_id = ?", acc.OrganizationID).Find(acc).Error // try to find in db
			if err == gorm.ErrRecordNotFound {                                                                         // not found in db
				ctx = context.WithValue(ctx, "janus_context", &Account{})
				r = r.WithContext(ctx)
				next.ServeHTTP(w, r)
				return
			}
			// found in db, save to cache
			jsun, _ = jettison.Marshal(acc)
			_ = j.cache.Set(fmt.Sprintf("%v-%v", acc.CacheKey, acc.OrganizationID), jsun)
		}

		_ = json.Unmarshal(jsun, acc)

		ctx = context.WithValue(ctx, "janus_context", acc)
		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)
	}
}

// SetRights set rights of a particular account
func (j *Janus) SetRights(acc *Account) error {
	jsun, _ := jettison.Marshal(acc)
	err := j.cache.Set(fmt.Sprintf("%v-%v", acc.CacheKey, acc.OrganizationID), jsun)
	if err != nil {
		return err
	}

	err = j.db.Save(acc).Error
	if err != nil {
		return err
	}

	return nil
}
