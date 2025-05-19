package domain

import "context"

type EntryRepository interface {
	Get(ctx context.Context, key string) (*Entry, error)
	Set(ctx context.Context, key string, entry *Entry) error
	Remove(ctx context.Context, key string) error
}
