package storage

import (
	"context"
	"data_storage/server/domain"
	"sync"
	"time"
)

// Data is a thread-safe, TTL-backed in-memory repository.
type Data struct {
	mu       sync.RWMutex
	data     map[string]*domain.Entry
	interval time.Duration
	stop     chan struct{}
}

// NewDataRepo creates the in-memory store and immediately
// starts a background goroutine that evicts expired entries.
func NewDataRepo(invTimeInterval time.Duration) *Data {
	d := &Data{
		data:     make(map[string]*domain.Entry),
		interval: invTimeInterval,
		stop:     make(chan struct{}),
	}
	go d.invalidate()
	return d
}

// Get retrieves an entry by key.
// Returns ErrNotFound if the key is missing,
// or ErrExpiredEntry once time.Now() ≥ Expiry.
func (d *Data) Get(ctx context.Context, key string) (*domain.Entry, error) {
	d.mu.RLock()
	entry, ok := d.data[key]
	d.mu.RUnlock()

	if !ok {
		return nil, domain.ErrNotFound
	}
	// inclusive expiration: now ≥ Expiry is expired
	if !entry.Expiry.IsZero() && !time.Now().Before(entry.Expiry) {
		return nil, domain.ErrExpiredEntry
	}
	return entry, nil
}

// Set inserts or updates an entry.
// Returns ErrEmptyKey if key is empty, ErrEmptyEntry if entry is nil.
func (d *Data) Set(ctx context.Context, key string, entry *domain.Entry) error {
	if key == "" {
		return domain.ErrEmptyKey
	}
	if entry == nil {
		return domain.ErrEmptyEntry
	}

	d.mu.Lock()
	d.data[key] = entry
	d.mu.Unlock()
	return nil
}

// Remove deletes the entry for the given key.
// Returns ErrEmptyKey if key is empty.
func (d *Data) Remove(ctx context.Context, key string) error {
	if key == "" {
		return domain.ErrEmptyKey
	}

	d.mu.Lock()
	delete(d.data, key)
	d.mu.Unlock()
	return nil
}

// invalidate runs every d.interval and removes any entries
// whose Expiry ≤ the tick time.
func (d *Data) invalidate() {
	ticker := time.NewTicker(d.interval)
	defer ticker.Stop()

	for {
		select {
		case now := <-ticker.C:
			d.mu.Lock()
			for k, entry := range d.data {
				if !entry.Expiry.IsZero() && !now.Before(entry.Expiry) {
					delete(d.data, k)
				}
			}
			d.mu.Unlock()
		case <-d.stop:
			return
		}
	}
}

// ShutDownInvalidation stops the background cleanup goroutine.
// Call this during graceful shutdown.
func (d *Data) ShutDownInvalidation() {
	close(d.stop)
}
